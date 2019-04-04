package plaid

import (
	"github.com/plaid/plaid-go/plaid"
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