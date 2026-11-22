# 快速开始

## 安装

使用 Go 1.27.0 或更高版本。创建应用模块、添加 nemo 并安装生成器：

```sh
mkdir nemo-example
cd nemo-example
go mod init example.com/nemo-example
go get github.com/longxi-go/nemo
go install github.com/longxi-go/nemo/cmd/nemoctl@latest
```

生成器安装到 `GOBIN`；未设置时使用 `$(go env GOPATH)/bin`。将该目录加入 `PATH`。
应用依赖与生成器应固定为相同的 nemo 版本，以便复现构建结果。

## 声明并加载配置

`// +nemo:config:prefix=app` 选择需要生成绑定方法的 `Config` 类型，并提供默认绑定前缀。
`required` 和 `default` 字段标签配置生成的 Binder。
支持的注解和标签见[注释注解指南](annotations.md)。

将以下程序保存为 `main.go`：

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

解析模块依赖、生成绑定方法并运行程序：

```sh
go mod tidy
nemoctl generate --roots .
go run .
nemoctl check --roots .
```

输出为 `nemo 8080`。

生成器在选中的包内写入 `zz_nemo.gen.go`。首次生成时，即使方法尚不存在，
也能分析 `binding.Bind[Config]` 调用；其他类型错误仍会使生成失败。
分发包时应包含生成源码，运行时不需要生成器。

传给 `binding.Bind` 的空前缀采用注解中的 `app` 前缀；
直接调用 `config.BindProperties(snapshot, "")` 则绑定根路径。
标签和支持的类型见[类型化绑定](binding.md)。

## 加载文件

使用 YAML 时，将静态 Provider 的创建过程替换为：

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

导入 `github.com/longxi-go/nemo/pkg/codec`，删除不再使用的 gosr 和 maptypes 导入。
将配置保存为 `application.yaml`：

```yaml
app:
  name: nemo
  port: 9090
```

程序输出 `nemo 9090`。`false` 表示文件必需存在；可选文件允许缺失，
但仍拒绝非法内容和其他读取错误。内置文件 Provider 不会自动监听文件。
显式发现与配置源优先级见[配置语义](configuration.md)，刷新与监听生命周期见[架构](architecture.md)。

仓库的 [config 示例](https://github.com/longxi-go/nemo/tree/main/examples/config)
还包含时长和嵌套配置。在仓库内执行 `go run ./examples/config`，输出为 `nemo 8080 3s localhost`。
