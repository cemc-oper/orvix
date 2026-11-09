# jobspec 契约

orvix 的 `#ORVIX` 指令集不是私有词汇，而是 **jobspec 契约** 的命令行投影——
这份契约由本仓库权威持有。

## 什么是 jobspec 契约

jobspec 是一份**语言中立的作业运行资源描述契约**：一份 JSON Schema（19 个字段），
描述一个 HPC 作业运行所需的全部资源与调度选项。工作流生成器（takflow 及其上层应用）
把作业资源渲染成 `#ORVIX` 指令，orvix 再把指令翻译给具体调度器——契约是两者之间的
接缝，保证"生成端写的"和"提交端认的"始终一致。

- **契约的唯一所有者（single source of truth）是 orvix**：
  `spec/jobspec/jobspec.schema.json`（仓库根下，公开可见），版本号以
  `spec/jobspec/VERSION` 文件为准。Go 访问包 `spec/jobspec/spec.go` 通过 `go:embed`
  直接嵌入同目录的权威文件，仓内只有这一份副本。
- **消费方持有只读 vendor 副本**，pin 到某个 orvix 提交/发布，由各自的一致性检查守护。
  **不要在消费方仓库改契约**；契约变更一律在 orvix 发起，消费方升 pin 同步。

## 键名对应关系

契约中的字段名为 snake_case（如 `job_name`），作为 `#ORVIX` 指令使用时改为连字符形式
（如 `job-name`）。orvix 解析器识别的指令集合（`internal/directive/parser.go` 的
`KnownDirectives`）由 `spec/jobspec` 的 schema 派生校验（`spec_test.go`），两者不会漂移。

19 个键的语义、类型与默认值见 [指令参考](directives.md)；各后端的映射见
[调度器映射](scheduler-mapping.md)。

## 版本与 vendor 关系

```
orvix  spec/jobspec/                     ← 权威：schema + VERSION + conformance 资产
  │
  │  vendor（只读副本 + pin）
  ├──► takflow                src/takflow/spec/jobspec/ + ORVIX_VERSION
  │      同步：takflow 侧 scripts/sync_jobspec.py --orvix-ref <orvix ref>
  │
  └──► orvix-hpc-test         spec/jobspec/ + ORVIX_REF
         同步：scripts/sync_conformance.sh；scripts/ci.sh drift 检查漂移
```

契约版本遵循语义化版本：新增可选键为 MINOR，修正描述等不改变行为的变化为 PATCH。

## 一致性保障

契约的正确性由本仓库的 golden 测试端到端保障：
`spec/jobspec/conformance/vectors/` 下的测试向量（携带 `#ORVIX` 指令的脚本）
经 `orvix generate --scheduler {slurm,donau,local}` 等价生成管线处理后，与
`conformance/golden/<scheduler>/` 的基线逐字节比对（`conformance_test.go`，
CI 的 `jobspec golden` job）。有意改变输出时用 `make regen-golden` 重新生成基线，
diff 随同一 commit 评审。

消费方另设验收门：

- **takflow** 断言自己的渲染器能复现 vendor 的向量，且 pin 版本的 `orvix generate`
  输出与 vendor 的 golden 一致。
- **orvix-hpc-test** 为每个向量自动合成 `conf_*` 用例，在真实环境（含无调度器的
  local 平台）对真实 orvix 二进制做 golden diff。

新增或修改指令键的完整清单见 [新增一个指令](../develop/adding-directive.md)。

## 相关链接

- 契约 schema：`spec/jobspec/jobspec.schema.json`（本仓库，权威）
- 契约版本：`spec/jobspec/VERSION`
- conformance 资产：`spec/jobspec/conformance/`
- 消费方仓库：`framework/takflow`、`runtime/orvix-devel/repo/orvix-hpc-test`（本工作区内）
