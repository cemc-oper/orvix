# jobspec 契约

orvix 的 `#ORVIX` 指令集不是私有词汇，而是 **jobspec 契约** 的命令行投影——
这份契约由本仓库权威持有。

## 什么是 jobspec 契约

jobspec 是一份**语言中立的作业运行资源描述契约**：一份 JSON Schema（19 个字段），
描述一个 HPC 作业运行所需的全部资源与调度选项。工作流生成器（takflow 及其上层应用）
把作业资源渲染成 `#ORVIX` 指令，orvix 再把指令翻译给具体调度器——契约是两者之间的
接缝，保证"生成端写的"和"提交端认的"始终一致。

- **契约的唯一所有者（single source of truth）是 orvix**：
  `spec/jobspec/jobspec.schema.json`（仓库根下，公开可见），版本号在
  `spec/jobspec/VERSION`（当前 `1.0.0`）。Go 访问包 `spec/jobspec/spec.go`
  通过 `go:embed` 直接嵌入同目录的权威文件，仓内只有这一份副本。
- **takflow vendor 了一份只读副本**（`takflow/spec/jobspec/`，pin 到某个 orvix
  提交/发布）。该副本由 takflow 侧的一致性检查守护，**不要指望在 takflow 改契约**；
  契约变更一律在 orvix 发起，takflow 升 pin 同步。

## 键名对应关系

契约中的字段名为 snake_case（如 `job_name`），作为 `#ORVIX` 指令使用时改为连字符形式
（如 `job-name`）。orvix 解析器识别的指令集合（`internal/directive/parser.go` 的
`KnownDirectives`）由 `spec/jobspec` 的 schema 派生校验（`spec_test.go`），
两者不会漂移。

契约相对 orvix 早期指令词汇的一个新增键是 **`requeue`**（v1 引入）：
`requeue=false` 时 SLURM 后端生成 `#SBATCH --no-requeue`，Donau 和 local 忽略。

## 一致性保障

契约的正确性由本仓库的 golden 测试端到端保障：
`spec/jobspec/conformance/vectors/` 下的测试向量（携带 `#ORVIX` 指令的脚本）
经 `orvix generate --scheduler {slurm,donau,local}` 等价生成管线处理后，与
`conformance/golden/<scheduler>/` 的基线逐字节比对（`conformance_test.go`，
CI 的 `jobspec golden` job）。有意改变输出时用 `make regen-golden` 重新生成基线，
diff 随同一 commit 评审。

takflow 侧另有一道消费方验收门：pin 住某个 orvix 版本，断言自己的渲染器能复现
vendor 的向量、且 pin 版本的 `orvix generate` 输出与 vendor 的 golden 一致。

新增或修改指令键的完整清单见 [新增一个指令](../develop/adding-directive.md)。

## 相关链接

- 契约 schema：`spec/jobspec/jobspec.schema.json`（本仓库，权威）
- conformance 资产：`spec/jobspec/conformance/`
- takflow 仓库（消费方）：`framework/takflow`（本工作区内）
