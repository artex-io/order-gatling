module github.com/alexppxela/order-gatling

go 1.20

require (
	github.com/google/uuid v1.3.0
	github.com/prometheus/client_golang v1.14.0
	github.com/quickfixgo/enum v0.1.0
	github.com/quickfixgo/field v0.1.0
	github.com/quickfixgo/quickfix v0.7.0
	github.com/quickfixgo/tag v0.1.0
	github.com/rs/zerolog v1.29.1
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.15.0
	github.com/sylr/quickfixgo-fix50sp2/executionreport v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/newordersingle v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/ordercancelreject v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/ordercancelreplacerequest v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/ordermasscancelreport v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/ordermasscancelrequest v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/quote v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/quotecancel v0.0.0-20220401195242-281940b8a21e
	github.com/sylr/quickfixgo-fix50sp2/quotestatusreport v0.0.0-20220401195242-281940b8a21e
	sylr.dev/fix v0.1.0
)

replace (
	github.com/quickfixgo/enum => github.com/sylr/quickfixgo-enum v0.0.0-20220401193143-29a559514373
	github.com/quickfixgo/field => github.com/sylr/quickfixgo-field v0.0.0-20220401193046-ca4cd16301d2
	github.com/quickfixgo/quickfix => github.com/sylr/quickfix-go v0.6.1-0.20221223080000-9e4d31ed9df6
	github.com/quickfixgo/tag => github.com/sylr/quickfixgo-tag v0.0.0-20220401193001-96cf7367fdfa
)

require (
	filippo.io/age v1.1.1 // indirect
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/armon/go-proxyproto v0.0.0-20210323213023-7e956b284f0a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/go-set v0.1.13 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.40.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/quickfixgo/fixt11 v0.1.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sylr.dev/yaml/age/v3 v3.0.0-20221203153010-eb6b46db8d90 // indirect
	sylr.dev/yaml/v3 v3.0.0-20220527135632-500fddf2b049 // indirect
)

replace sylr.dev/fix => /Users/alex/git/fix
