module github.com/dougbeal/legerdemain

require (
	github.com/abourget/ledger v0.0.0-20170120042254-bdeca00c3c45
	github.com/fatih/gomodifytags v0.0.0-20180914191908-141225bf62b6 // indirect
	github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d
	github.com/juju/loggo v0.0.0-20190212223446-d976af380377
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/plaid/plaid-go v0.0.0-20190218043405-1109794cef7b
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/tools v0.0.0-20190330180304-aef51cc3777c // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/abourget/ledger => ../goledger

go 1.13
