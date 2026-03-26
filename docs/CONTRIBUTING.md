# 贡献指南 - 蓝鲸监控平台

## 1. 工具安装

### Go 安装

```shell
# 安装 gvm (Go 版本管理器)
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source ~/.gvm/scripts/gvm
gvm install go1.24.13 -B
gvm use go1.24.13
```

### uv 安装 (Python 包管理器)

```shell
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### GTM 安装 (Git Task Manager)

```shell
uv tool install "bk-gtm>0.7,<0.8" --index-url=https://mirrors.tencent.com/pypi/simple --extra-index-url=https://mirrors.tencent.com/repository/pypi/tencent_pypi/simple
```

### GPG 安装

```shell
# 测试是否已安装，如果未安装，根据系统选择安装方式
gpg --version

sudo apt-get install gnupg # Ubuntu/Debian
brew install gnupg # macOS
sudo yum install gnupg # CentOS/RHEL
```

## 2. 开发环境初始化

### GitHub ssh 配置（如果已配置则跳过）

```shell
ssh-keygen -t rsa -b 4096
cat ~/.ssh/id_rsa.pub # 将输出配置到 GitHub 上
```
- 生成公私钥后复制公钥文件（id_rsa.pub）的内容到 [GitHub SSH](https://github.com/settings/ssh/new) 进行配置。

### 项目基础配置

```shell
# Fork 项目到个人仓库后克隆
git clone <您的个人仓库地址>
cd bkmonitor-ecosystem/
git remote add upstream git@github.com:TencentBlueKing/bkmonitor-ecosystem.git
git config user.name "您的GitHub用户名"
git config user.email "您的GitHub邮箱"
make init # 初始化开发环境
source .venv/bin/activate
```

### GPG 密钥配置

- 生成 GPG 密钥对
```shell
gpg --full-generate-key
# 推荐配置选项
# Please select what kind of key you want: 4    # (4) RSA (sign only)
# What keysize do you want? (3072) 4096
# Key is valid for? 3y
# Is this correct? y
# Real name: （github 用户名）
# Email address: （github 邮箱）
# Comment: 为空
# Change (N)ame, (C)omment, (E)mail or (O)kay/(Q)uit? o
# Passphrase: 为空然后一直回车
```

- 查看 GPG 密钥
```shell
gpg --list-secret-keys --keyid-format=long
```

输出示例：
```shell
sec   rsa4096/XXXXXXXXXXXXXXXX 2024-12-09 [SC] [expires: 2027-12-09]
      YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
uid                 [ultimate] Your Name <your@email.com>
```
其中 `XXXXXXXXXXXXXXXX` 即为你的 <GPG_Key_ID>（GPG 密钥 ID），后续配置中需要使用。

- 导出公钥并添加到 Git 平台
```shell
gpg --armor --export <您的GPG密钥ID>
```
将输出的公钥内容（以 `-----BEGIN PGP PUBLIC KEY BLOCK-----` 开头）完整复制到 [GitHub gpg 配置页面](https://github.com/settings/gpg/new)。

- 配置 Git 使用 GPG 签名
```shell
git config user.signingkey <您的GPG密钥ID>
git config commit.gpgsign true
git config tag.gpgsign true
# 验证配置是否生效
git config --list | grep gpg
```

### PreCI 工具配置（仅内部员工使用）

```shell
preci server --start # 按提示输入验证信息，有可能还需要申请权限，按工具输出指引申请即可
preci init
preci info --update # 打开配置编辑界面
```

在弹出的编辑窗口中，添加以下安全检查配置：
```yaml
checkerSetBasesList:
  - checkerSetId: standard_scc
    checkerSetLang: Other
    enable: true
  - checkerSetId: opensource_sensitive_other
    checkerSetLang: Other
    enable: true
```

执行以下命令验证 PreCI
```shell
git add .
preci scan --pre-commit
```

### GTM 配置

在仓库根目录创建 `.gtm.yaml` 文件：

```yaml
github:
  access_token: "<github_access_token>"
tapd:
  api_user: "<tapd_api_user>"
  api_password: "<tapd_api_password>"
  workspace_id: "10158081"
  username: "<tapd_username>"
  milestone_id: "1010158081002236619"
  tapd_story:
    status_done: "for_test"
    status_release: "status_3,status_9"
    status_doing: "developing,status_7"
  tapd_bug:
    status_done: "resolved"
    status_release: "verified"
    status_doing: "assigned,in_progress"
project:
  default_label: "project/apm"
  reviewers: ["liuwenping", "ZhuoZhuoCrayon", "joker-joker-yuan"]
```
- github_access_token: 访问 [GitHub Token](https://github.com/settings/tokens/new) 申请。
- tapd_api_user, tapd_api_password, tapd_username 请咨询同事获取。

## 3. 贡献指南

使用 GTM（Git Task Manager）工具来简化和标准化开发流程，确保代码质量和协作效率：

### 创建任务单据

```shell
gtm create
# 参考示例：
#? 请选择创建的类型 单据事件, 包含TAPD需求/缺陷和Github Issue
#? 请选择操作 创建新单据
#? 请选择需求类型 feat: A new feature. Correlates with MINOR in SemVer
#? 本次改动的范围 (类名或文件名, 回车可跳过)
#? 本次改动的内容描述 自定义指标上报优化
#? 请选择关联的里程碑 1010158081002236619. 自定义指标上报优化（项目默认里程碑）
#单据创建中...
#? 请输入目标分支 main
#? 请输入新的开发分支名称，请取一个可读的名称方便理解，系统会自动补充前缀和后缀，即 (feat/$input/#1010158081132974932)，input:  contributing_add_dev
```

### 提交代码

```shell
make render
gtm commit
```

### 创建 Pull Request

```shell
gtm pr
```

## 4. 使用变量

- 基础结构
```shell
templates/              →  渲染引擎（Jinja2）  →  docs/open/     （开源版文档）
    *.md（模板文件）                              docs/inner/    （内部版文档，仅内部环境生成）
tools/formatter/
    context/
        base.py         （基础框架：Field、FieldManager、FieldMeta）
        open.py         （开源版变量定义，开源环境和内部环境均存在）
        inner.py        （内部版变量定义，仅内部环境存在，文件存在 -> 处于内部环境）
    main.py             （渲染入口）
```

- 变量定义示例
```python
# from . import base  # 在实际代码中导入 base 模块
# class EcosystemRepositoryUrl(metaclass=base.FieldMeta):
#     class Meta:
#         name = "ECOSYSTEM_REPOSITORY_URL"          # 模板中通过 `{{name}}` 引用的变量名。
#         scope = "open"                             # 变量的作用域，`OPEN` 表示开源版，`INNER` 表示内部版。
#         value = "https://github.com/TencentBlueKing/bkmonitor-ecosystem"  # 变量的值，支持字符串、字典等类型。
```

## 5. 内外部仓库协作模式

```text
┌─────────────────────────────────────────────────┐
│                 开源版仓库（主仓库）                │
│                                                 │
│  feature branch ──PR──▶ main（合并）              │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
              流水线自动触发（单向同步）
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│                 内部版仓库（从仓库）                │
│                                                 │
│  同步分支（自动 merge）──手动 PR──▶ main           │
│       +                                         │
│  <内部版专属 PR> ──PR──▶ main（合并）              │
└─────────────────────────────────────────────────┘
```
- 新增 / 修改内部版专属变量，添加内部版特定模板和修改 `inner` 目录内容时，请额外提交内部版专属 PR。

