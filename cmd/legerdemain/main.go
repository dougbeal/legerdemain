package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"os"
	"os/user"
	"path"

	"log"

	rdebug "runtime/debug"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
	"github.com/juju/gnuflag"
	"github.com/plaid/plaid-go/plaid"
	"gopkg.in/yaml.v2"
)

var (
	verbose         = gnuflag.Bool("verbose", false, "")
	debug           = gnuflag.Bool("debug", false, "")
	journal         = gnuflag.String("file", "", "Path of ledger file. (default: value from .ledgerrc.)")
	ledgerrc        = gnuflag.String("init-file", "", "Path of init file.  (default: search ~/.ledgerrc, ./.ledgerrc)")
	output          = gnuflag.String("output", "/dev/stdout", "Path of output file.")
	currency        = gnuflag.String("currency", "USD", "Default currency if none set.")
	configDir       = gnuflag.String("config", "~/.legerdmain", "Path of legerdmain config")
	plaidMode       = gnuflag.Bool("plaid", false, "Intake plaid data")
	skipLedger      = gnuflag.Bool("skip", false, "Skip parsing existing ledger")
	plaidConfigFile = gnuflag.String("plaidconf", "", "Set location of plain config file")
)

type PlaidEnvironment struct {
	Name   string `yaml:"name"`
	Secret string `yaml:"secret"`
}

type PlaidInstitution struct {
	Name        string `yaml:"name"`
	ItemId      string `yaml:"item_id"`
	AccessToken string `yaml:"access_token"`
}

type User struct {
	LedgerFileName string             `yaml:"ledger_file_name"`
	Institutions   []PlaidInstitution `yaml:"institutions"`
}

type PlaidConfig struct {
	ClientID     string             `yaml:"client_id"`
	PublicKey    string             `yaml:"public_key"`
	Environments []PlaidEnvironment `yaml:"environments"`
	Users        []User             `yaml:"users"`
}

const LedgerRCFileName string = ".ledgerrc"

func abortOnError(err error) {
	if err != nil {
		log.Print("abortOnError: ")
		log.Fatalln(err)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("stacktrace from panic: %s\n", rdebug.Stack())
		}
	}()
	gnuflag.Parse(true)

	if *verbose {
		log.Println(os.Args)
		log.Print(gnuflag.Args())
	}
	if !*skipLedger {
		parseLedger(parseLedgerRC(""))
	}
	if *plaidMode {
		if *plaidConfigFile == "" {
			*plaidConfigFile = path.Join(*configDir, "plaid.yaml")
		}
		data, err := ioutil.ReadFile(*plaidConfigFile)
		if *debug {
			log.Print("Plaid config file: ")
			log.Println(string(data))
		}
		abortOnError(err)

		plaidConfig := &PlaidConfig{}
		err = yaml.Unmarshal(data, plaidConfig)
		if *debug {
			log.Print("Unmarshal'd plaid config file: ")
			log.Printf("%+v\n:", plaidConfig)

		}
		abortOnError(err)

		clientOptions := plaid.ClientOptions{
			ClientID:    plaidConfig.ClientID,
			Secret:      plaidConfig.Environments[0].Secret,
			PublicKey:   plaidConfig.PublicKey,
			Environment: plaid.Development, //   * development: Test your integration with live credentials; you will need to request access before you can access our Development environment
			HTTPClient:  &http.Client{},
		}
		client, err := plaid.NewClient(clientOptions)
		abortOnError(err)

		accountsR, err := client.GetAccounts(plaidConfig.Users[0].Institutions[0].AccessToken)

		if plaidError := err.(plaid.Error); plaidError.ErrorType == "ITEM_ERROR" && plaidError.ErrorCode == "ITEM_LOGIN_REQUIRED" {
			// requires user interaction
			PlaidLink(Settings{"transactions", plaidConfig.Environments[0].Name, plaidConfig.PublicKey}, client)
		}

		abortOnError(err)
		log.Printf("%+v\n:", accountsR)

	} else {
	}

}

type Settings struct {
	PlaidProducts    string
	PlaidEnvironment string
	PlaidPublicKey   string
}

func PlaidLink(settings Settings, client *plaid.Client) {
	var accessToken string
	t, err := template.ParseFiles("templates/index.html")
	abortOnError(err)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			log.Printf("http url: %+v\n", r.URL)
			log.Printf("http rqst: %+v\n", r)
		}
		t.Execute(w, settings)

	})

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/get_access_token", func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			log.Printf("http url: %+v\n", r.URL)
			log.Printf("http rqst: %+v\n", r)
		}
		if r.Method == "POST" {
			r.ParseForm()

			public_token := r.Form["public_token"]
			accessTokenResponse, err := client.ExchangePublicToken(public_token[0])
			log.Println("Public token -> Access Token", accessTokenResponse.AccessToken, "for item:", accessTokenResponse.ItemID)
			accessToken = accessTokenResponse.AccessToken
			if err != nil {
				log.Print("abortOnError: ")
				log.Println(err)
			}
		}

	})

	// http.HandleFunc("/auth", methods=["GET"])
	// http.HandleFunc("/transactions", methods=["GET"])
	// http.HandleFunc("/identity", methods=["GET"])
	// http.HandleFunc("/balance", methods=["GET"])
	// http.HandleFunc("/accounts", methods=["GET"])
	// http.HandleFunc("/assets", methods=["GET"])
	// http.HandleFunc("/item", methods=["GET"])
	// http.HandleFunc("/set_access_token", methods=["POST"])

	http.ListenAndServe(":8080", nil)

}

func parseLedgerRC(file string) string {
	return "/Users/dougbeal/git.private/finances/foolscap/foolscap.ledger"
}

func initFilePath() string {
	home := func() (string, error) {
		if usr, err := user.Current(); err != nil {
			return "", err
		} else {
			return usr.HomeDir, err
		}
	}
	pathFunctions := []func() (string, error){home, os.Getwd}

	for _, fn := range pathFunctions {
		if directory, err := fn(); err != nil {
			filePath := path.Join(directory, LedgerRCFileName)
			if _, err := os.Stat(filePath); err != nil {
				return directory
			}
		}
	}
	return ""
}

func parseLedger(ledgerFile string) {
	if _, err := os.Stat(ledgerFile); os.IsNotExist(err) {
		log.Fatalln(err, "ldgFile doesn't exist.  %s ", ledgerFile)
	}
	cnt, err := ioutil.ReadFile(ledgerFile)
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}

	tree := parse.New(ledgerFile, string(cnt))
	err = tree.Parse()
	if err != nil {
		log.Fatalln("Parsing error:", err)
	}

	for _, nodeIface := range tree.Root.Nodes {
		switch node := nodeIface.(type) {
		case *parse.XactNode:
			fmt.Println(node)
		case *parse.CommentNode:

		case *parse.SpaceNode:

		default:
			fmt.Printf("unprintable node type %T\n", nodeIface)
		}
	}
	if *verbose {
		printer := print.New(tree)

		buf := &bytes.Buffer{}
		err = printer.Print(buf)
		if err != nil {
			log.Fatalln("rendering ledger file:", err)
		}

		var dest io.Writer
		dest = os.Stdout
		// if inFile != "" && *writeOutput {
		//	destFile, err := os.Create(inFile)
		//	if err != nil {
		//		log.Fatalln("Couldn't write to file:", inFile)
		//	}
		//	dest = destFile
		//	defer destFile.Close()
		// }

		_, err = dest.Write(buf.Bytes())
		if err != nil {
			log.Fatalln("Error writing to file:", err)

		}
	}
}
