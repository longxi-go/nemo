# 仓库脚本

[English](README.md) | [简体中文](README.zh_CN.md)

使用 Python 3.12 和兼容 `go.mod` 的 Go 工具链，在仓库根目录运行 Make 目标。
格式化还需要 Git 和 Bash；GitHub 发布预检查需要已认证的 GitHub CLI。

| 脚本 | 职责 | 目标 |
| --- | --- | --- |
| `golayout.py` | 解析声明的 golayout 模块，将文件选择、格式化和文件头处理委托给其共享脚本 | `setup`、`format`、`check-format`、`test-format`、`license-headers`、`check-license-headers` |
| `release.py` | 验证稳定标签、模块身份、依赖版本和发布说明；按需检查 Pages 设置及上游 Release | `validate-release`、`preflight-release` |
| `publish_release.py` | 核对标签与已检查的提交，创建或复用内容一致的 Release，并等待 Latest 可查询 | Release 工作流 |
| `go_environment.py` | 拒绝进程或持久配置中可能替换模块依赖图的 Go 标志 | `verify-environment` |
| `go_module.py` | 通过 `go mod edit -json` 解析模块文件，不解析依赖 | 被其他脚本导入 |

`GO` 选择 Go 可执行文件。格式化语言版本读取本模块的 Go 指令。
`make check` 禁用 workspace 和自动工具链切换，公开依赖版本由 `go.mod`、`go.sum` 声明。
验证入口拒绝非空 `GOFLAGS`，包括持久 Go 配置中的标志，避免替换模块依赖图。
代理和校验服务的配置保持有效。

`make test-scripts` 运行 Python 单元测试。模块元数据测试使用真实 Go。
只读检查的回归测试也调用真实 Go，验证缺少校验记录时报错且不修改 `go.mod`、`go.sum`。
部分子进程和 API 边界使用受控响应，测试非法数据、缺失脚本、拒绝标签部署的策略、
未发布 Release 等情况，不发送网络请求。发布器测试还覆盖 Latest 延迟、已发布内容冲突，
以及重试不得重复写入的约束。格式化适配器测试检查委托行为；
`make test-format` 使用真实 golayout 验证格式化及文件选择。
GitHub 预检查的单元测试不能证明远程仓库配置正确，实际检查应执行
`make preflight-release TAG=vX.Y.Z`。

发布顺序和公开安装验证见[发布指南](../docs/maintainers/releases.zh_CN.md)。
