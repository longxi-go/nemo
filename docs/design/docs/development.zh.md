# 开发指南

## 仓库工具与质量门禁

开发需要 Git、Bash、Make、Python 3.12 及[质量工作流](https://github.com/longxi-go/nemo/blob/main/.github/workflows/quality.yml)
选择的 Go 工具链。`go.mod` 声明最低 Go 版本。
安装工作流指定版本的 golangci-lint 并加入 `PATH`，或通过 `GOLANGCI_LINT` 指定其可执行文件。

在仓库根目录运行：

```sh
make setup
make license-headers format
make generate
GOWORK=off make check
```

`setup` 解析固定版本的 golayout 工具，不安装格式化器二进制，也不编辑模块文件。
`license-headers` 将 `codequality/HEADER` 应用于仓库内的 Go 文件。
`format` 执行共享 golayout 策略；仅使用上游 gofmt 或 gofumpt 不能落实完整代码布局规则。

`check` 是只读门禁，检查源码文件头、格式、格式化器集成、模块依赖一致性、vet、覆盖率测试、
随机顺序 race 测试、golangci-lint 和生成文件一致性；它会按需下载工具依赖。
使用 `make help` 查看单独目标。修改生成器后应重新生成示例，并通过 `make check-generated` 检查输出。

`make build` 将开发 CLI 写入 `.tmp/bin/nemoctl`。
在仓库外使用时，采用[快速开始](quick-start.md)中的公开安装命令。

## 源码与测试夹具约定

公开 API 位于 `pkg`，实现位于 `internal`，轻量可执行入口位于 `cmd`。
每个仓库内的包提供包文档、可运行示例和单元测试。运行时绑定不得使用反射或 unsafe。

格式化选择符合条件的工程 Go 文件，以及完整且可解析的 `testdata/**/*.go.tmpl`。
普通 testdata、缓存、vendor 目录和符号链接不参与。文件头刷新还排除 testdata，
并遵循 `codequality/HEADER` 中声明的版权所有者。

生成器在类型验证和输出前调用 golayout 共享格式化 API。
可选文件头模板由目标模块拥有，不能将 nemo 的声明复制到消费方生成文件。
较长的测试程序使用按场景组织的源码夹具；负向测试需要的非法语法必须保持原样。

## 多模块工作区 {#local-workspace}

联合开发时，将需要修改的工程克隆为同级目录。例如在父目录运行：

```sh
go work init ./golayout ./gosr ./gosr/tools ./nemo
```

已有工作区使用 `go work use` 扩展。`gosr/tools` 有独立 `go.mod`，必须单独加入；
其他模块按需要加入。`go.work` 和 `go.work.sum` 保持本地使用，
生产模块不包含同级目录替换。

在 Nemo 目录执行 `make check WORKSPACE=../go.work` 可验证指定工作区，临时测试
使用方也采用其中选定的模块。未指定 `WORKSPACE` 时，检查使用已发布的依赖图。
检查拒绝进程或持久配置中的 `GOFLAGS` 覆盖。

## 文档

双语站点使用 Python 3.12 和 uv：

```sh
make docs-setup
make docs-serve
make docs-build
```

英文预览地址为 `http://127.0.0.1:8000/nemo/`，中文为 `http://127.0.0.1:8000/nemo/zh/`。
源码位于 `docs/design/docs`；新增页面同时提供配对的 `page.md` 与 `page.zh.md`，并更新两份导航。
两种语言中的命令、API 语义和约束都应依据源码审查。
Mermaid 图表还需在浏览器中检查渲染；Markdown 构建成功不能验证图表语法或可读性。

站点文件见[文档维护指南](https://github.com/longxi-go/nemo/blob/main/docs/design/README.zh.md)，
发布过程见[维护者发布指南](https://github.com/longxi-go/nemo/blob/main/docs/maintainers/releases.zh_CN.md)。
