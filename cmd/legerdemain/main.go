package main

import (
	"bytes"
	"fmt"

	"io"
	"io/ioutil"
	"net/http"

	"os"
	"os/user"
	"path"



	rdebug "runtime/debug"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
	"github.com/juju/gnuflag"

	"gopkg.in/yaml.v2"
	"github.com/dougbeal/legerdemain/pkg/plaid"

	"log"
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


func abortOnError(err error) {
	if err != nil {
		log.Print("abortOnError: ")
		log.Fatalln(err)
	}
}
const LedgerRCFileName string = ".ledgerrc"



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

		plaidConfig := &plaid.PlaidConfig{}
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
			plaid.PlaidLink(Settings{"transactions", plaidConfig.Environments[0].Name, plaidConfig.PublicKey}, client)
		}

		abortOnError(err)
		log.Printf("%+v\n:", accountsR)

	} else {
	}

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
