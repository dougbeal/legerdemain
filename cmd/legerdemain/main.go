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

	"log"

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

type PlaidConfig struct {
	ClientID  string `json:"client_id"`
	Secret    string `json:"secret"`
	PublicKey string `json:"public_key"`
}

const LedgerRCFileName string = ".ledgerrc"

func abortOnError(err error) {
	log.Fatalln(err)
}

func main() {
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
		abortOnError(err)

		plaidConfig := &PlaidConfig{}
		err = yaml.Unmarshal(data, plaidConfig)
		abortOnError(err)

		clientOptions := plaid.ClientOptions{
			ClientID:    plaidConfig.ClientID,
			Secret:      plaidConfig.Secret,
			PublicKey:   plaidConfig.PublicKey,
			Environment: plaid.Development,
			HTTPClient:  &http.Client{},
		}
		_, err = plaid.NewClient(clientOptions)
		abortOnError(err)

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
