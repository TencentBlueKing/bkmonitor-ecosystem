# AI 导航

本文件只承担导航职责。

目标：帮助模型按当前任务找到正确入口，避免重复读目录，或改错文档源文件。

## 0x01 项目定位

- `bkmonitor-ecosystem` 提供最小化、开箱即用的观测数据接入 demo，覆盖 Go、Python、Java、JavaScript 和 C++。
- 项目同时包含文档模板、渲染脚本和多语言示例。

## 0x02 判断是否处于内部版

- 如果存在 [tools/formatter/context/inner.py](tools/formatter/context/inner.py)，说明当前处于内部版环境。
- 处于内部版环境时，优先阅读 `docs/inner/` 路径，而不是 `docs/open/` 路径。
- 如果某条导航只给了开源版路径，先将路径中的 `open` 替换为 `inner`，再检查对应文档是否存在。

## 0x03 参与项目贡献所需知识

适用场景：修改文档、模板、渲染逻辑、示例代码，或准备提交贡献。

优先阅读这些文件：

| 路径 | 内部版路径 | 用途 |
| --- | --- | --- |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | - | 查看开发环境初始化、标准命令、协作流程和模板渲染机制。 |
| [.github/copilot-instructions.md](.github/copilot-instructions.md) | - | 查看项目级约束，尤其是文档、示例和开源内容边界。 |
| [Makefile](Makefile) | - | 查看项目提供的标准开发命令。 |
| [.pre-commit-config.yaml](.pre-commit-config.yaml) | - | 查看提交前会自动执行哪些检查。 |
| [tools/formatter/main.py](tools/formatter/main.py) | - | 查看 [templates/](templates/) 如何渲染到 [docs/](docs/)。 |
| [tools/formatter/context/open.py](tools/formatter/context/open.py) | [tools/formatter/context/inner.py](tools/formatter/context/inner.py) | 查看模板变量定义。 |
| [examples/common/ob-all-in-one/README.md](examples/common/ob-all-in-one/README.md) | - | 查看本地开发和联调 demo 上报链路的最小环境。 |

## 0x04 理解项目已有沉淀

适用场景：快速了解项目提供了哪些文档、样例和接入方案。

优先阅读这些文件：

| 路径 | 内部版路径 | 用途 |
| --- | --- | --- |
| [README.md](README.md) | - | 了解项目目标和总入口。 |
| [docs/open/README.md](docs/open/README.md) | [docs/inner/README.md](docs/inner/README.md) | 查看 APM 接入知识总索引，一览按语言整理的接入文档和样例。 |
| [docs/open/cookbook/README.md](docs/open/cookbook/README.md) | [docs/inner/cookbook/README.md](docs/inner/cookbook/README.md) | 查看 cookbook 知识总索引，一览按语言整理的自定义事件和自定义指标文档。 |
| [docs/open/common/examples/helloworld.md](docs/open/common/examples/helloworld.md) | [docs/inner/common/examples/helloworld.md](docs/inner/common/examples/helloworld.md) | 了解项目希望用什么样的最小示例帮助用户理解接入流程。 |
