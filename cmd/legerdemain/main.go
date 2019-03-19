package main

import (
	"os"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
	"github.com/juju/gnuflag"
	"path"
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
	parseLedger( parseLedgerRC )
}

func parseLedgerRC(string file) string {
	return "./foolscap.ledger"
}

func initFilePath() string {
	if usr, err := os.user.Current(); err != nil {
		rcInHome := path.join(usr.HomeDir, ledgerRCFileName)
		if _, err := os.Stat(rcInHome); err != nil {
			return rcInHome
		}
	}
	rcInCurrentDirectory := path.join(os.Getwd(), ledgerRCFileName)
	if _, err := os.Stat(rcInCurrentDirectory); err != nil {
		return rcInCurrentDirectory
	}
	return ""
}

func parseLedger(string ledgerFile) {
	if _, err := os.Stat(ledgerFile); os.IsNotExist(err) {
		checkf(err, "ldgFile doesn't exist.  %s ", *ledgerFile)
	}
	cnt, err := ioutil.ReadFile(*ledgerFile)
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}

	tree := parse.New(*ledgerFile, string(cnt))
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
		printer := print.New(t)

		buf := &bytes.Buffer{}
		err = printer.Print(buf)
		if err != nil {
			log.Fatalln("rendering ledger file:", err)
		}

		var dest io.Writer
		dest = os.Stdout
		if inFile != "" && *writeOutput {
			destFile, err := os.Create(inFile)
			if err != nil {
				log.Fatalln("Couldn't write to file:", inFile)
			}
			dest = destFile
			defer destFile.Close()
		}

		_, err = dest.Write(buf.Bytes())
		if err != nil {
			log.Fatalln("Error writing to file:", err)

		}
	}
}
