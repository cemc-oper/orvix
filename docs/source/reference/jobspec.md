# jobspec 契约

orvix 的 `#ORVIX` 指令集不是私有词汇，而是 **takflow jobspec 契约** 的命令行投影。

## 什么是 jobspec 契约

jobspec 是 takflow 框架维护的一份**语言中立的作业运行资源描述契约**：
一份 JSON Schema（19 个字段），描述一个 HPC 作业运行所需的全部资源与调度选项。
工作流生成器（takflow）把作业资源渲染成 `#ORVIX` 指令，orvix 再把指令翻译给具体调度器——
契约是两者之间的接缝，保证"生成端写的"和"提交端认的"始终一致。

- **契约的唯一所有者（single source of truth）是 takflow**：
  `framework/takflow/spec/jobspec/jobspec.schema.json`，版本号在 `spec/jobspec/VERSION`
  （当前 `1.0.0`）。
- **orvix 只是 vendor 了一份只读副本**：`internal/spec/jobspec.schema.json` 和
  `internal/spec/VERSION`。该副本由一致性测试守护，**不要手工编辑**；
  需要变更时先在 takflow 侧修改契约，再重新 vendor 到 orvix。

## 键名对应关系

契约中的字段名为 snake_case（如 `job_name`），作为 `#ORVIX` 指令使用时改为连字符形式
（如 `job-name`）。orvix 解析器识别的指令集合（`internal/directive/parser.go` 的
`KnownDirectives`）由 vendor 的 schema 派生校验，两者不会漂移。

契约相对 orvix 早期指令词汇的一个新增键是 **`requeue`**（v1 引入）：
`requeue=false` 时 SLURM 后端生成 `#SBATCH --no-requeue`，Donau 和 local 忽略。

## 一致性保障

契约的正确性由 takflow 仓库的 conformance 测试端到端保障：
`takflow/spec/jobspec/conformance/` 下的测试向量（携带 `#ORVIX` 指令的脚本）
经 `orvix generate --scheduler {slurm,donau,local}` 生成后与 golden 文件比对。
新增或修改指令键时必须同步更新契约、vendor 副本与 conformance 向量，
完整清单见 [新增一个指令](../develop/adding-directive.md)。

## 相关链接

- takflow 仓库：`framework/takflow`（本工作区内）
- 契约 schema：`framework/takflow/spec/jobspec/jobspec.schema.json`
- orvix vendor 副本：`internal/spec/`
