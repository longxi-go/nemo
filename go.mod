module github.com/longxi-go/nemo

go 1.27.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/longxi-go/golayout v0.1.4
	github.com/longxi-go/gosr v0.1.1
	github.com/longxi-go/gosr/tools v0.1.0
	github.com/magiconair/properties v1.8.10
	go.yaml.in/yaml/v3 v3.0.5
	golang.org/x/tools v0.49.0
	gopkg.in/ini.v1 v1.67.1
	sigs.k8s.io/controller-tools v0.20.1
)

require (
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	k8s.io/apimachinery v0.35.0 // indirect
	mvdan.cc/gofumpt v0.12.0 // indirect
)

tool github.com/longxi-go/golayout/cmd/golayout
