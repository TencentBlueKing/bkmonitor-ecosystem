# Contributing to 蓝鲸监控平台

蓝鲸团队秉持开放的态度，欢迎志同道合的开发者一起贡献项目。在开始之前，请先阅读以下指引。

## 目录

- [1. GPG 配置](#1-gpg-配置)
- [2. 环境配置](#2-环境配置)
- [3. Pull Request](#3-pull-request)
- [4. Git Commit Message 规范](#4-git-commit-message-规范)
- [5. GTM 规范](#5-GTM-规范)
- [6. 文档变量及渲染方式](#6-文档变量及渲染方式)
- [7. 内外部仓库协作模式](#7-内外部仓库协作模式)

## 1. GPG 配置

为确保提交的真实性和完整性，所有贡献者必须对 Git Commit 进行 GPG 签名。

### 1.1 生成 GPG 密钥

如果您还没有 GPG 密钥，可通过以下命令生成：

```sh
gpg --full-generate-key
```

推荐选项：
- 密钥类型：RSA
- 密钥长度：4096
- 有效期：3y（根据需要设置）
- 姓名和邮箱：确保和 git config 配置的 user.name 和 user.email 保持一致
- 密码：以后每次加密操作都会让你输入密码，可以为空

### 1.2 查看 GPG 密钥

```sh
gpg --list-secret-keys --keyid-format=long
```

输出示例：
```shell
sec   rsa4096/XXXXXXXXXXXXXXXX 2024-12-09 [SC] [expires: 2027-12-09]
      YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
uid                 [ultimate] Your Name <your@email.com>
```

其中 `XXXXXXXXXXXXXXXX` 即为你的 <GPG_Key_ID>（GPG 密钥 ID）。

### 1.3 导出公钥并添加到 Git 平台

```sh
gpg --armor --export <GPG_Key_ID>
```

将输出的公钥内容复制并添加到你的 Git 平台的 GPG Keys 设置中（如 Github -> Settings -> SSH and GPG keys -> New GPG key）。

### 1.4 配置 Git 使用 GPG 签名

```sh
# 设置 GPG 签名密钥
git config user.signingkey <GPG_Key_ID>

# 开启自动签名默认对所有 commit/tag 签名
git config commit.gpgsign true
git config tag.gpgsign true
```

## 2. 环境配置

本项目使用 [pre-commit](https://pre-commit.com/) 框架，在提交代码前自动执行一系列检查和格式化操作，确保代码质量和文档规范。

### 2.1 环境初始化

执行以下命令完成依赖安装和 pre-commit hook 注册（需要 [Go](https://go.dev/) 语言和 Python 库管理工具 [uv](https://github.com/astral-sh/uv)）：

```sh
make init
```

该命令会完成以下操作：
- 安装 Go 格式化工具（`gofumpt`、`goimports-reviser`）
- 通过 `uv sync` 安装 Python 依赖
- 通过 `uv run pre-commit install` 注册 Git pre-commit hook

### 2.2 PreCI 说明（外部贡献者可忽略）

PreCI 是内部 CI 预检查工具。在 pre-commit 阶段通过 `tools/scripts/pre-commit/pre_check.sh` 调用：

- 如果存在内部脚本（`tools/scripts/pre-commit/inner/pre_check.sh`），将优先执行内部脚本。
- 如果环境中已安装 `preci` 命令行工具，则执行 `preci scan --pre-commit` 进行预检查。
- 如果未安装 `preci`，将自动跳过该检查。
- Windows 平台（MSYS/Cygwin）会自动跳过。

### 2.3 本地接收端

项目在 `examples/common/ob-all-in-one` 下提供了基于 Docker Compose 的本地可观测性接收端，可用于快速验证 demo 的数据上报逻辑。
快速开始 👉 <a href="../examples/common/ob-all-in-one/README.md" target="_blank">ob-all-in-one</a>。
建议优先使用蓝鲸监控平台的线上环境进行数据上报，本地接收端仅作为开发调试的辅助手段。

## 3. Pull Request

Pull Request（简称 PR）是我们使用的主流协作方式，请参考 [Github 文档](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests) 了解 PR 的功能。

我们采用 ["Fork and Pull"](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/getting-started/about-collaborative-development-models#fork-and-pull-model) 的开发模式，即开发者首先向自己的 fork 仓库提交变更，然后向上游仓库发起 PR。

### 3.1 基本原则

- 目标分支：除非有特殊约定，所有 PR 的目标分支应始终选择 `main` 分支。
- 关联 Issue：提交 PR 时必须在标题或描述中关联对应的 Issue 编号。
- No Merge-Commit：本仓库遵循 `no merge-commit` 原则，遇到代码冲突时，应使用 `rebase`（而非 `merge`）进行处理。
- Commit 整理：请将个人仓库（分支）下产生的零碎 commit 压缩合并，并填写有意义的 commit message。

### 3.2 提交步骤

1. Fork 代码仓库（仅首次操作）；
2. 创建 Issue，描述需求或问题；
3. 在个人仓库下新建分支进行开发，分支命名建议按照 `<分类>/<issue id>` 格式，例如：`feature/#1`；
4. 进行开发，此阶段可根据开发进度自由设置 commit；
5. 完成自测，确保相关 Issue 描述的功能已实现；
6. 使用 `git rebase` 整理 commit，严格按照 [Git Commit Message 规范](#git-commit-message-规范) 填写提交信息；
7. 向主仓库发起 PR（PR 标题默认使用 commit message），并指定 Code Review 负责人；
8. 根据 CR 意见修改代码。此阶段允许提交修复类的小 commit，一般使用 `fixup` 作为 commit 前缀，以避免 CR 负责人反复拉取分支；
9. CR 通过；
10. PR 提交者将个人仓库下的 `fixup` commit 通过 `git rebase` 压缩，确保最终只保留一个 commit；
11. 仓库 Maintainer 完成合并（仓库仅开放 Rebase Merging 选项）；
12. 关闭 Issue。

## 4. Git Commit Message 规范

为统一不同团队的提交信息格式，所有 commit message 须遵循以下规范：

```shell
git commit -m '<标记>: <概要说明> #<Issue ID>'
```

示例：

```shell
git commit -m 'fix: 修复数据上报异常问题 #29'
```

### 4.1 标记说明

| 标记           | 说明                         |
|:-------------|:---------------------------|
| feat         | 新增功能特性                     |
| fix          | 缺陷修复                       |
| refactor     | 代码重构（不涉及功能变更）              |
| test         | 增加或修改测试代码                  |
| docs         | 文档编写或更新                    |
| merge        | 分支合并及冲突解决                  |

## 5. GTM 规范

GTM（Git Task Manager）是蓝鲸内部的 Git 任务管理工具，用于简化 Issue 创建、代码提交和 PR 管理等日常开发流程。

### 5.1 安装

```shell
pip install bk-gtm --index-url=https://mirrors.tencent.com/pypi/simple --extra-index-url=https://mirrors.tencent.com/repository/pypi/tencent_pypi/simple
```

### 5.2 常用命令

| 命令                 | 说明                                    |
|:-------------------|:--------------------------------------|
| `gtm create`       | 交互式创建 Issue 或 Task（根据引导选择类型）          |
| `gtm create issue` | 创建 Issue                              |
| `gtm create task`  | 创建 Task                               |
| `gtm commit`       | 交互式提交代码，自动关联 Issue 并规范 commit message |
| `gtm pr`           | 基于当前分支创建 Pull Request，自动填充 PR 信息      |

## 6. 文档变量及渲染方式

本项目的文档采用 `Jinja2 模板` 和 `变量上下文` 的方式管理，支持同一套模板渲染出面向不同环境（开源版 / 内部版）的文档。

### 6.1 整体流程

```text
templates/              →  渲染引擎（Jinja2）  →  docs/open/     （开源版文档）
    *.md（模板文件）                              docs/inner/    （内部版文档，仅内部环境生成）
        +
tools/formatter/
    context/
        base.py         （基础框架：Field、FieldManager、FieldMeta）
        open.py         （开源版变量定义）
        inner.py        （内部版变量定义，仅内部环境存在）
    main.py             （渲染入口）
```

执行渲染：

```sh
make render
```

### 6.2 模板语法

模板文件位于 `templates/` 目录，使用标准 Jinja2 语法引用变量。例如：

```markdown
# 服务快速接入指引（Go）

{{QUICK_START_OVERVIEW}}

## 初始化示例 demo

git clone {{ECOSYSTEM_REPOSITORY_URL}}
cd {{ECOSYSTEM_REPOSITORY_NAME}}/examples/go-examples/helloworld
```

- `{{VARIABLE_NAME}}`：引用简单变量，渲染时会被替换为对应值。
- `{{access_config.otlp.http_endpoint}}`：支持嵌套字典属性访问。
- 非 `.md` 文件（如图片）将直接拷贝到目标目录，不做渲染。

### 6.3 变量定义

变量通过 Python 元类（`FieldMeta`）注册到全局的 `FieldManager` 中。每个变量定义为一个类，包含内部 `Meta` 类：

```python
from . import base

class EcosystemRepositoryUrl(metaclass=base.FieldMeta):
    class Meta:
        name = "ECOSYSTEM_REPOSITORY_URL"          # 模板中引用的变量名
        scope = base.ScopeType.OPEN.value          # 作用域：open / inner
        value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem"  # 变量值
```

- `name`：模板中通过 `{{name}}` 引用的变量名。
- `scope`：变量的作用域，`OPEN` 表示开源版，`INNER` 表示内部版。
- `value`：变量的值，支持字符串、字典等类型。

变量定义文件：
- 开源版变量：`tools/formatter/context/open.py`
- 内部版变量：`tools/formatter/context/inner.py`（仅内部环境存在）

### 6.4 环境区分

渲染引擎会根据是否存在 `inner.py` 模块自动判断当前环境：

- 开源环境（open）：仅渲染开源版变量，输出到 `docs/open/`。模板路径中包含 `inner` 的文件会被跳过。
- 内部环境（inner）：同时渲染开源版和内部版，分别输出到 `docs/open/` 和 `docs/inner/`。

### 6.5 Markdown 链接转换

渲染过程中，模板和输出文件中的 Markdown 格式链接（`[text](url)`）会被自动转换为 HTML `<a>` 标签（带 `target="_blank"`），以便在各平台上获得更好的阅读体验。

## 7. 内外部仓库协作模式

本项目同时维护 `开源版` 和 `内部版` 两个仓库，二者通过流水线进行单向同步。

### 7.1 仓库区分

| 仓库  | 说明        | 特征                                      |
|-----|-----------|-----------------------------------------|
| 开源版 | 面向社区的公开仓库 | 不包含任何带 `inner` 字眼的文件夹或文件                |
| 内部版 | 内部使用的仓库   | 在开源版基础上，额外包含 `inner` 相关的文件（如内部模板、内部变量等） |

### 7.2 协作流程

当内部版有一个新需求时，需要同时在开源版和内部版分别提交 PR：

1. 开源版 PR：实现需求的核心功能，包含所有可以公开的代码和文档。
2. 内部版 PR：仅补充开源版无法包含的内容（如内部模板、内部变量等），不重复实现开源版已有的功能。

代码合并的完整流程如下：

```text
┌─────────────────────────────────────────────────┐
│                 开源版仓库                        │
│                                                 │
│  feature branch ──PR──▶ main（合并）              │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
              流水线自动触发（merge）
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│                 内部版仓库                        │
│                                                 │
│  同步分支（自动 merge）──手动 PR──▶ main           │
│       +                                         │
│  内部版补充 PR ──PR──▶ main（合并）                │
└─────────────────────────────────────────────────┘
```

具体步骤：

1. 在开源版仓库基于需求创建分支，完成开发后提交 PR 并合并到主分支。
2. 开源版主分支合并后，流水线自动触发，将开源版的变更以 `merge` 的方式合并到内部版仓库的某个同步分支。
3. 在内部版仓库中，基于该同步分支手动向主分支提交 PR，完成代码合入。
4. 如果需求涉及内部专属内容（如 `inner` 相关变量或配置），则在内部版仓库另行提交补充 PR，合并到主分支。

### 7.3 注意事项

- 开源版不应包含任何内部信息：文件名、目录名或内容中带有 `inner` 字眼的，均不得出现在开源版仓库中。
- 内部版以开源版为基础：内部版的改动应保持为开源版的超集，避免在内部版中修改开源版已有的逻辑。
- 冲突处理：如果流水线自动 merge 时产生冲突，需在同步分支或开源版上解决后再向主分支提 PR。
