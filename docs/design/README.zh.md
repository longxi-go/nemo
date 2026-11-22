# 文档站点

[English](README.md) | [简体中文](README.zh.md)

nemo 手册使用 MkDocs Material，包含英文和简体中文页面。
构建需要 Python 3.12 和 uv，依赖由 `uv.lock` 固定。

## 构建与预览

在仓库根目录执行：

```sh
make docs-setup
make docs-serve
make docs-build
```

`docs-serve` 的英文预览地址为 `http://127.0.0.1:8000/nemo/`，
中文为 `http://127.0.0.1:8000/nemo/zh/`。`docs-build` 执行严格构建和站点检查。
在 `docs/design` 中检查已构建的站点：

```sh
uv run --locked python -m unittest discover -s tests
```

Mermaid 图表在浏览器预览时渲染。除静态检查外，还应审查图表渲染和双语标签。

## 站点文件

| `docs/design` 内的路径 | 用途 |
| --- | --- |
| `mkdocs.yml` | 主题、导航、仓库和语言配置 |
| `docs/` | 英文页面及配对的 `*.zh.md` 中文译文 |
| `uv.lock` | 锁定的文档依赖 |
| `tests/` | 构建后站点回归检查 |

页面按相同文件名配对，并加入两份导航。生成的 `site/` 和 `.venv/` 目录不纳入版本控制。
通过以 `/` 结尾的 `DOCS_SITE_URL` 覆盖基础 URL。
保留指向 `longxi-go/nemo` 的 Material 原生仓库组件，用于显示仓库元数据。

发布过程见[维护者发布指南](../maintainers/releases.zh_CN.md)。
