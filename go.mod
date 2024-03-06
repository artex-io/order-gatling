module github.com/alexppxela/order-gatling

go 1.21

toolchain go1.21.4

require (
	github.com/google/uuid v1.6.0
	github.com/prometheus/client_golang v1.19.0
	github.com/quickfixgo/enum v0.1.0
	github.com/quickfixgo/field v0.1.0
	github.com/quickfixgo/fix50sp2 v0.1.0
	github.com/quickfixgo/quickfix v0.9.0
	github.com/quickfixgo/tag v0.1.0
	github.com/rs/zerolog v1.32.0
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2
	sylr.dev/fix v0.1.1-0.20230220140741-b9e365fa1f2c
)

replace (
	github.com/quickfixgo/enum => github.com/alexppxela/quickfixgo-enum v0.0.0-20240123152700-3da353f95834
	github.com/quickfixgo/field => github.com/alexppxela/quickfixgo-field v0.0.0-20240123152653-f0a896744ceb
	github.com/quickfixgo/fix50sp2 => github.com/alexppxela/quickfixgo-fix50sp2 v0.0.0-20240123165040-5354803752b9
	github.com/quickfixgo/quickfix => github.com/alexppxela/quickfixgo v0.0.0-20240123152459-2877207c30e2
	github.com/quickfixgo/tag => github.com/alexppxela/quickfixgo-tag v0.0.0-20240123152647-cf0a4d9525ed
	sylr.dev/fix => github.com/alexppxela/fix v0.0.0-20240124083433-6f543f7a6628
)

require (
	filippo.io/age v1.1.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/armon/go-proxyproto v0.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hashicorp/go-set v0.1.14 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.49.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/quickfixgo/fixt11 v0.1.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sylr.dev/yaml/age/v3 v3.0.0-20221203153010-eb6b46db8d90 // indirect
	sylr.dev/yaml/v3 v3.0.0-20220527135632-500fddf2b049 // indirect
)

//
//replace sylr.dev/fix => /Users/alex/git/fix
