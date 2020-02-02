package plaid

import (

	"net/http"

	template "github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/juju/loggo"
	"github.com/plaid/plaid-go/plaid"
)

//go:generate go run assets/generate.go

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

type Settings struct {
	PlaidProducts    string
	PlaidEnvironment string
	PlaidPublicKey   string
}


type Plaid struct {
	*plaid.Client
}

var log loggo.Logger

func init() {
	log = loggo.GetLogger("legerdemain/pkg/plaid")
}

func NewPlaid(config PlaidConfig) (p *Plaid, err error) {
	clientOptions := plaid.ClientOptions{
		ClientID:    config.ClientID,
		Secret:      config.Environments[0].Secret,
		PublicKey:   config.PublicKey,
		Environment: plaid.Development, //   * development: Test your integration with live credentials; you will need to request access before you can access our Development environment
		HTTPClient:  &http.Client{},
	}
	client, err := plaid.NewClient(clientOptions)
	return &Plaid{client}, err

}

// TODO: load assets into binary https://tech.townsourced.com/post/embedding-static-files-in-go/ (probably with vfsgen, but maybe filb0x)
func (p *Plaid) PlaidLink(settings Settings, tokChan chan<- plaid.ExchangePublicTokenResponse, errChan chan<- error)  {
	var handler = http.NewServeMux()
	var server = &http.Server { Addr: ":8080",
		Handler: handler}
	
	
	t, err := template.ParseFiles(assets, nil, "/templates/index.html")
	if err != nil {
		log.Errorf("%s\n", err)
		panic(err)
	}	

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("http url: %+v\n", r.URL)
		log.Infof("http rqst: %+v\n", r)
		t.Execute(w, settings)

	})

	fs := http.FileServer(http.Dir("static"))
	handler.Handle("/static/", http.StripPrefix("/static/", fs))

	handler.HandleFunc("/get_access_token", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("http url: %+v\n", r.URL)
		log.Infof("http rqst: %+v\n", r)

		if r.Method == "POST" {
			r.ParseForm()

			public_token := r.Form["public_token"]
			accessTokenResponse, err := p.ExchangePublicToken(public_token[0])
			if err != nil {
				log.Errorf("%s\n", err)
				errChan <- err				
			}
			log.Debugf("Public token -> Access Token", accessTokenResponse.AccessToken, "for item:", accessTokenResponse.ItemID)
			tokChan <- accessTokenResponse
			close(tokChan)
			close(errChan)
			server.Close()
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


	server.ListenAndServe()

}
