module github.com/shadowapi/shadowapi/backend

go 1.24

require (
	github.com/caarlos0/env/v9 v9.0.0
	github.com/go-faster/errors v0.7.1
	github.com/go-faster/jx v1.1.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/gotd/contrib v0.21.0
	github.com/gotd/td v0.120.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/mdp/qrterminal v1.0.1
	github.com/nats-io/nats.go v1.38.0
	github.com/ogen-go/ogen v1.10.0
	github.com/ory/kratos-client-go v1.2.1
	github.com/ory/ladon v1.3.0
	github.com/phsym/console-slog v0.3.1
	github.com/prometheus/client_golang v1.21.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/samber/do/v2 v2.0.0-beta.7
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	go.mau.fi/whatsmeow v0.0.0-20250221160813-35b965ceadf1
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/metric v1.34.0
	go.opentelemetry.io/otel/trace v1.34.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/oauth2 v0.24.0
	google.golang.org/api v0.213.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/auth v0.13.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.6 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.12 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/go-faster/yaml v0.4.6 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/ory/pagination v0.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/samber/go-type-to-string v1.7.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.mau.fi/libsignal v0.1.2 // indirect
	go.mau.fi/util v0.8.5 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/exp v0.0.0-20250215185904-eff6e970281f // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241216192217-9240e9c98484 // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	rsc.io/qr v0.2.0 // indirect
)

tool github.com/ogen-go/ogen/cmd/ogen
