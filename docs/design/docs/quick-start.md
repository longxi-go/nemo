# Quick start

## Install

Use Go 1.27.0 or later. Create an application module, add nemo and install its generator:

```sh
mkdir nemo-example
cd nemo-example
go mod init example.com/nemo-example
go get github.com/longxi-go/nemo
go install github.com/longxi-go/nemo/cmd/nemoctl@latest
```

The generator is installed in `GOBIN`, or `$(go env GOPATH)/bin` when `GOBIN` is
unset. Add that directory to `PATH`. Pin the application dependency and generator
to the same nemo version for reproducible builds.

## Declare and load configuration

`// +nemo:config:prefix=app` selects `Config` for generation and supplies the
default binding prefix. The `required` and `default` field tags configure its
generated binder. The [annotation guide](annotations.md) documents every supported
annotation and tag.

Save this program as `main.go`:

```go
package main

import (
    "context"
    "fmt"

    gosr "github.com/longxi-go/gosr/pkg/environment"
    "github.com/longxi-go/nemo/pkg/binding"
    "github.com/longxi-go/nemo/pkg/environment"
    "github.com/longxi-go/nemo/pkg/maptypes"
    "github.com/longxi-go/nemo/pkg/source"
)

// +nemo:config:prefix=app
type Config struct {
    Name string `required:"true"`
    Port int    `default:"8080"`
}

func main() {
    properties, err := source.NewMap("application", maptypes.ValueMap{
        "app.name": gosr.String("nemo"),
    })
    if err != nil {
        panic(err)
    }
    owner, err := environment.New(
        context.Background(),
        environment.WithProviders(source.Static(properties)),
    )
    if err != nil {
        panic(err)
    }
    config, err := binding.Bind[Config](owner.Snapshot(), "")
    if err != nil {
        panic(err)
    }
    fmt.Println(config.Name, config.Port)
}
```

Resolve the application imports, generate the binder and run the program:

```sh
go mod tidy
nemoctl generate --roots .
go run .
nemoctl check --roots .
```

Output: `nemo 8080`.

The generator writes `zz_nemo.gen.go` in the selected package. It can analyze the
first invocation of `binding.Bind[Config]` before that method exists. Unrelated
type errors still fail generation. Include generated source when distributing
the package; consumers do not need the generator at runtime.

The empty prefix passed to `binding.Bind` selects the annotation's `app` prefix.
A direct `config.BindProperties(snapshot, "")` call instead binds the root.
See [typed binding](binding.md) for tags and supported types.

## Load a file

To use YAML, replace the static-provider setup with:

```go
decoder, err := codec.ForFormat("yaml")
if err != nil {
    panic(err)
}
provider, err := source.NewFile("application", "application.yaml", decoder, false)
if err != nil {
    panic(err)
}
owner, err := environment.New(context.Background(), environment.WithProviders(provider))
if err != nil {
    panic(err)
}
```

Import `github.com/longxi-go/nemo/pkg/codec` and remove the unused gosr and maptypes imports.
Save the configuration in `application.yaml`:

```yaml
app:
  name: nemo
  port: 9090
```

The program prints `nemo 9090`. `false` makes this file required; optional
files tolerate absence but still reject malformed content and other read errors.
Built-in file providers do not watch files automatically. See
[configuration](configuration.md) for explicit discovery and source precedence,
and [architecture](architecture.md) for refresh and watch lifecycle.

The repository's [config example](https://github.com/longxi-go/nemo/tree/main/examples/config)
also includes durations and nested configuration. Run it from a checkout with
`go run ./examples/config`; its output is `nemo 8080 3s localhost`.
