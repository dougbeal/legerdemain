package main

import (
	"bytes"
	"fmt"

	"io"
	"io/ioutil"

	"os"
	"os/user"
	"path"

	rdebug "runtime/debug"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
	"github.com/juju/gnuflag"

	plaid "github.com/dougbeal/legerdemain/pkg/plaid"
	plaidapi "github.com/plaid/plaid-go/plaid"
	"gopkg.in/yaml.v2"

	"log"

	"github.com/pkg/browser"
	"time"
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
	plaidConfigFile = gnuflag.String("plaidconf", "", "Set location of plaid config file")
)

func abortOnError(err error) {
	if err != nil {
		log.Print("abortOnError: ")
		log.Println(err)
		log.Fatalf("%s\n", rdebug.Stack())
	}
}

const LedgerRCFileName string = ".ledgerrc"

func main() {
	var plaidConfig = &plaid.PlaidConfig{}
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("stacktrace from panic (%s): %s\n", r, rdebug.Stack())
		}
	}()
	gnuflag.Parse(true)

	if *verbose {
		log.Println(os.Args)
		gnuflag.Visit(func(f *gnuflag.Flag) { log.Print(f) })
		log.Println(gnuflag.Args())
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
		if *debug {
			log.Print("Plaid config file: ")
			log.Println(string(data))
		}

		err = yaml.Unmarshal(data, plaidConfig)
		if *debug {
			log.Printf("Unmarshal'd plaid config file (%s): \n", *plaidConfigFile)
			log.Printf("%+v\n:", plaidConfig)

		}
		abortOnError(err)

		client, err := plaid.NewPlaid(*plaidConfig)

		abortOnError(err)

		accountsResp, err := client.GetAccounts(plaidConfig.Users[0].Institutions[0].AccessToken)

		if err != nil {
			if plaidError, ok := err.(plaidapi.Error); ok && plaidError.ErrorType == "ITEM_ERROR" && plaidError.ErrorCode == "ITEM_LOGIN_REQUIRED" {
				respChan := make(chan plaidapi.ExchangePublicTokenResponse)
				errChan := make(chan error)
				// requires user interaction
				go client.PlaidLink(plaid.Settings{"transactions", plaidConfig.Environments[0].Name, plaidConfig.PublicKey}, respChan, errChan)
				go browser.OpenURL("http://localhost:8080")
				select {
				case response := <-respChan:
					log.Printf("Exchanged for access token: %s\n", response)
				case err := <-errChan:
					abortOnError(err)
				}
			}
		}

		log.Printf("accounts %+v\n:", accountsResp)
		var accountIDs []string
		for _, account := range accountsResp.Accounts {
			accountIDs = append(accountIDs, account.AccountID)
		}
		options := plaidapi.GetTransactionsOptions{
			AccountIDs: accountIDs,
			StartDate: time.Now().AddDate(-10, 0, 0).Format("2006-02-01"), // must be YYYY-MM-DD
			EndDate:   time.Now().Format("2006-02-01"),
			Count:     500,
			Offset:    0,
		}
		log.Printf("getting transactions %+v\n", options)
		resp, err := client.GetTransactionsWithOptions(plaidConfig.Users[0].Institutions[0].AccessToken, options)
		if plaidError, ok := err.(plaidapi.Error); ok {
			abortOnError(plaidError)
		}
		log.Printf("transactions %+v\n:", resp)
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
			log.Fatalf("Fatal error: rendering ledger file:%s\n%s\n", err, rdebug.Stack())
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
