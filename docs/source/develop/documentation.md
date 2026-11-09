# 文档编写指南

本页面向修改 orvix 文档的贡献者，说明文档站的结构约定、写作规范与构建方法。
修改 orvix 代码时请同时遵守本页的"同步更新"要求。

## 文档站结构

文档站分为四区，顺序固定为 **开始 → 指南 → 开发 → 参考**，由 `docs/source/index.rst`
的顶层 toctree 定义。**顶层 toctree 只列四个区的落地页**（`start/index`、`guide/index`、
`develop/index`、`reference/index`），各区的子页面挂在该区落地页内的 ```{toctree}` 中——
这样顶部导航只有四项，侧边栏只展开当前区，不会把全部页面平铺出来。

**新增页面时**：放入对应区目录，并在该区 `index.md` 的 toctree 与导览列表中各加一条；
不要直接加到根 `index.rst`。

四区文体不同，写作前先判断内容属于哪一区：

| 区 | 目录 | 读者 | 文体 | 判断问题 |
|---|---|---|---|---|
| 开始 | `start/` | 第一次接触 orvix 的人 | 引导式教程，有起点有终点 | "读者照着做能跑通吗？" |
| 指南 | `guide/` | 日常使用 orvix 的人 | 任务导向的操作指引 | "读者要做成某件事，该看哪页？" |
| 开发 | `develop/` | 修改 orvix 本身的人 | 架构说明 + 贡献者清单 | "改代码需要知道什么？" |
| 参考 | `reference/` | 所有人 | 面向查找的表格与字段定义 | "读者要查某个键/选项/字段的准确含义？" |

归属规则：

- 表格型的全集定义（指令键、命令选项、后端映射、状态映射、文件字段）一律放**参考**；
  指南页只写"怎么用"，并链接到参考页。
- 教程只在**开始**区出现；指南页不重复安装步骤与入门示例。
- 页面以任务或主题命名，文件名用连字符小写（如 `monitoring-and-killing.md`）。
- 文档只写用户与开发者需要的内容，不记录重构过程、阶段状态或历史决策；过程材料放工作区的
  `doc/` 目录。

## 单一事实来源（SSOT）

每一项事实只在一个页面详写，其余页面写一句话结论并链接。当前分配如下：

| 事实 | 权威页 | 数据来源 |
|---|---|---|
| 指令键全集、类型、默认值 | `reference/directives.md` | `spec/jobspec/jobspec.schema.json` |
| 指令 → 调度器映射 | `reference/scheduler-mapping.md` | `internal/scheduler/{slurm,donau,local}.go` |
| 原始状态 → 归一化状态、终止态 | `reference/scheduler-mapping.md` | `internal/scheduler/state.go` 及各后端 `NormalizeState` |
| 命令选项、输出、退出码 | `reference/cli.md` | `cmd/*.go`、`main.go` |
| `.info.yaml` / `.submit` / `.submit.log` 格式 | `reference/file-formats.md` | `internal/jobinfo/`、`internal/script/render.go`、`internal/submit/` |
| 安装方法 | `start/installation.md` | `Makefile`、`.goreleaser.yaml`、`internal/version/` |
| 指令语法规则 | `guide/writing-scripts.md` | `internal/directive/parser.go` |
| 环境变量与默认值 | `reference/configuration.md` | 各源码文件 |
| 契约版本与 vendor 关系 | `reference/jobspec.md` | `spec/jobspec/VERSION` |
| 提交管线、接口定义 | `develop/architecture.md` | `internal/submit/`、`internal/scheduler/scheduler.go` |

**不要手抄易变值**：契约版本号、orvix 版本号等一律写"以 `<文件>` 为准"，不在散文里写死具体值。

## 写作规范

1. **页面骨架**：`# 标题` → 一段"本页解决什么、适合谁" → 正文 → 末尾"相关页面"链接（参考页可省）。
2. **指南页以任务命名**（"与 ecFlow 集成"而非"ecFlow 说明"），先给可复制执行的最短路径，
   再讲变体与注意事项。
3. **参考页以表格为主**，不写教程段落；指令表列保持 `指令 | 类型/默认 | 说明 | 示例`。
4. **代码块**：命令用 ```` ```bash ````；命令与输出同框时命令带 `$ ` 前缀；文件内容标注对应语言；
   流程图用 ```` ```{mermaid} ````。
5. **HPC-only 提示**：依赖真实调度器或 ecFlow 的示例，段首统一写
   `> **HPC-only**：…`，并尽量给出无需调度器的替代验证方式（如 `orvix generate`）。
6. **可核对性**：含版本、默认值、映射表的页面，末尾加一行"核对：`<源码路径>` 或 `<命令>`"。
7. **以代码为准**：所有命令行为、退出码、状态映射必须与实现一致。发现文档与代码不符时，
   以代码为准修文档，并在 PR 描述中列出这些修正。
8. **链接**：站内用 MyST 相对路径链接（如 `[指令参考](../reference/directives.md)`）；
   跨仓库内容（takflow、orvix-hpc-test）用文字说明其在工作区中的相对位置，不写绝对路径。

## 同步更新要求

- 修改用户可见行为（命令选项、输出、退出码、生成的文件、后端行为）时，**同一个 PR** 更新
  对应指南页与参考页。
- 新增或修改指令键时，按 [新增一个指令](adding-directive.md) 的清单执行，其中文档步骤即
  更新 `reference/directives.md` 与 `reference/scheduler-mapping.md`，并检查相关指南页与
  故障排查条目是否需要同步。
- 修改 CI、构建或发布流程时，同步更新 `develop/setup.md` / `develop/testing.md` /
  `develop/releasing.md`。

## 生成的指令表格

`reference/directives.md` 的**分组表与后端支持矩阵不是手写的**，而是生成后 include 的：

| 角色 | 文件 |
|---|---|
| 键集合、类型、默认值（权威） | `spec/jobspec/jobspec.schema.json` |
| 中文说明、示例、后端摘要、分组 | `docs/data/directives.json` |
| 生成脚本 | `docs/tools/generate_directive_tables.py`（仅用 Python 标准库） |
| 生成产物（不要手改） | `docs/source/reference/_generated/directive-tables.md` |

```bash
make docs-directives   # 重新生成
make docs-check        # 校验产物是否与 schema/数据一致，不一致即失败并给出 diff
make docs-html         # 严格构建文档（先跑 docs-check）
```

生成脚本会校验两侧键集合完全一致、分组无遗漏无重复、每键都有说明/示例/三后端摘要，
因此**契约加了键而文档数据没跟上会直接报错**。修改契约或指令说明后，
把重新生成的片段随同一个 commit 提交。

其余页面仍按 SSOT 表手工维护；`_generated/` 目录已在 `conf.py` 的 `exclude_patterns`
中排除，不会被当成独立文档。

## 本地构建

```bash
pip install -r docs/requirements.txt
cd docs && make html           # 输出 docs/build/html/
```

严格模式（警告即失败）加指令表漂移校验，提交前建议执行（在仓库根目录）：

```bash
make docs-html
```

等价于：

```bash
cd docs
sphinx-build -b html -W --keep-going -a -E source build/html
```

线上版本由 Read the Docs 构建（配置 `.readthedocs.yaml`，Sphinx 配置 `docs/source/conf.py`）；
CI 的 `docs` job 也会执行漂移校验与严格构建。
