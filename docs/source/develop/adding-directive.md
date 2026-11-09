# 新增一个指令

`#ORVIX` 指令集是 jobspec 契约的投影，而契约由本仓库权威持有——新增一个指令键
是**单仓库**变更，按以下清单逐项落实，缺一不可。唯一的跨仓库动作在最后：
通知消费方（takflow）升 pin 同步。

## 变更清单

### 1. 契约（本仓库 `spec/jobspec/`）

- `spec/jobspec/jobspec.schema.json`：新增 snake_case 字段（如 `new_key`），
  写明类型与说明；
- 按契约版本规则（语义化版本：新增可选键为 MINOR）更新 `spec/jobspec/VERSION`；
- `spec/jobspec/spec_test.go` 会随即变红（解析器词汇表与新 schema 不一致），
  这正是契约门的提示。

### 2. 解析器（本仓库）

- `internal/directive/parser.go`：在 `KnownDirectives` 中加入连字符形式的键
  （如 `new-key`）；
- `internal/directive/parser_test.go`：补充解析测试（带值 / 裸键 / 条件形式，
  视键的语义而定）。

### 3. 各调度器后端（本仓库）

为每个后端决定该键的映射，并在对应 `PreambleFor` 的 generator 列表中声明：

- `internal/scheduler/slurm.go` — 简单映射用 `sbatch` / `sbatchQ` / `sbatchBare`，
  特殊语义（如 `requeue` 的布尔反转）写定制 generator；
- `internal/scheduler/donau.go` — 注意合并类指令（`-R`、`-d`）的既有模式；
- `internal/scheduler/local.go` — 大多数资源键对 local 无意义，确认无需处理即可；
- 某后端无对应能力时**显式决定忽略**，并在文档中注明（如 `ntasks` 之于 Donau）。

为每个后端的映射补充单元测试。

### 4. conformance 向量与 golden（本仓库）

- 在 `spec/jobspec/conformance/vectors/` 增加（或修改）携带新指令的脚本向量；
- 跑 `make regen-golden` 重新生成 `golden/{slurm,donau,local}/` 基线，
  diff 随同一 commit 评审；
- CI 的 `jobspec golden` job（`TestConformanceGolden`）是这道门的常态守卫。

### 5. 文档（本仓库）

- `docs/source/reference/directives.md`：指令全集表格与后端支持矩阵各加一行；
- `docs/source/reference/scheduler-mapping.md`：各后端映射表加一行（或注明忽略）。

同时检查相关指南页是否需要同步：新键若有使用场景或易错写法，更新
`guide/writing-scripts.md`（配方/常见错误）、`guide/choosing-scheduler.md`（后端差异）与
`guide/troubleshooting.md`（症状条目）。页面归属与同步要求见
[文档编写指南](documentation.md)。

### 6. 通知消费方（takflow 仓库）

契约合并（并发布/打 tag）后，通知 takflow 升 pin：在 takflow 侧运行
`scripts/sync_jobspec.py --orvix-ref <新版 orvix ref>`，vendor 副本与
`ORVIX_VERSION` pin 一并更新。takflow 渲染器如需从应用层配置生成新键
（`TaskResource` → `ResourceSpec` 编译），在 takflow 那个 PR 里一并处理。

## 命名约定

- schema 字段用 snake_case（`new_key`），`#ORVIX` 指令用连字符（`new-key`），
  两者由 `spec/jobspec/spec.go` 的 `DirectiveKeys()` 机械转换；
- 调度器中立的布尔键默认值要为所有后端可解释的语义
  （参考 `requeue`：默认 `true`，`false` 时才产出调度器指令）。
