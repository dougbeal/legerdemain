package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
	"github.com/juju/gnuflag"
)

var (
	verbose  = gnuflag.Bool("verbose", false, "")
	debug    = gnuflag.Bool("debug", false, "")
	journal  = gnuflag.String("file", "", "Path of ledger file. (default: value from .ledgerrc.)")
	ledgerrc = gnuflag.String("init-file", "", "Path of init file.  (default: search ~/.ledgerrc, ./.ledgerrc)")
	output   = gnuflag.String("output", "/dev/stdout", "Path of output file.")
	currency = gnuflag.String("currency", "USD", "Default currency if none set.")
)

const LedgerRCFileName string = ".ledgerrc"

func main() {
	parseLedger(parseLedgerRC(""))
}

func parseLedgerRC(file string) string {
	return "./foolscap.ledger"
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
		// 	destFile, err := os.Create(inFile)
		// 	if err != nil {
		// 		log.Fatalln("Couldn't write to file:", inFile)
		// 	}
		// 	dest = destFile
		// 	defer destFile.Close()
		// }

		_, err = dest.Write(buf.Bytes())
		if err != nil {
			log.Fatalln("Error writing to file:", err)

		}
	}
}
