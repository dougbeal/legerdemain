module github.com/dougbeal/legerdmain

require (
	github.com/abourget/ledger v0.0.0-20170120042254-bdeca00c3c45
	github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d
	github.com/plaid/plaid-go v0.0.0-20190218043405-1109794cef7b
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/abourget/ledger => ../goledger
