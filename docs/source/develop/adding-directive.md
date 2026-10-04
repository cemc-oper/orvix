# 新增一个指令

`#ORVIX` 指令集是 takflow jobspec 契约的投影，因此新增一个指令键是**跨仓库**的变更，
需要按以下清单逐项落实，缺一不可。

## 变更清单

### 1. 契约（takflow 仓库）

在 takflow 侧修改契约，这是唯一权威来源：

- `framework/takflow/spec/jobspec/jobspec.schema.json`：新增 snake_case 字段
  （如 `new_key`），写明类型与说明；
- 按契约版本规则更新 `framework/takflow/spec/jobspec/VERSION`；
- 更新 takflow 的 jobspec 文档与 `TaskResource` → `ResourceSpec` 的编译逻辑
  （如果该键需要从应用层配置生成）。

### 2. vendor 副本（orvix 仓库）

把更新后的 `jobspec.schema.json` 和 `VERSION` 复制到 `internal/spec/`。
`internal/spec/spec_test.go` 会校验 vendor 副本与解析器指令集一致。

### 3. 解析器（orvix 仓库）

- `internal/directive/parser.go`：在 `KnownDirectives` 中加入连字符形式的键
  （如 `new-key`）；
- `internal/directive/parser_test.go`：补充解析测试（带值 / 裸键 / 条件形式，
  视键的语义而定）。

### 4. 各调度器后端（orvix 仓库）

为每个后端决定该键的映射，并在对应 `PreambleFor` 的 generator 列表中声明：

- `internal/scheduler/slurm.go` — 简单映射用 `sbatch` / `sbatchQ` / `sbatchBare`，
  特殊语义（如 `requeue` 的布尔反转）写定制 generator；
- `internal/scheduler/donau.go` — 注意合并类指令（`-R`、`-d`）的既有模式；
- `internal/scheduler/local.go` — 大多数资源键对 local 无意义，确认无需处理即可；
- 某后端无对应能力时**显式决定忽略**，并在文档中注明（如 `ntasks` 之于 Donau）。

为每个后端的映射补充单元测试。

### 5. conformance 向量（takflow 仓库）

在 `framework/takflow/spec/jobspec/conformance/vectors/` 增加携带新指令的脚本向量，
更新 `golden/{slurm,donau,local}/` 的期望输出，跑通一致性比对。

### 6. 文档

- `docs/source/guide/directives.md`：指令全集表格加一行；
- `docs/source/guide/schedulers.md`：各后端映射表加一行（或注明忽略）。

## 命名约定

- schema 字段用 snake_case（`new_key`），`#ORVIX` 指令用连字符（`new-key`），
  两者由 `internal/spec/spec.go` 的 `DirectiveKeys()` 机械转换；
- 调度器中立的布尔键默认值要为所有后端可解释的语义
  （参考 `requeue`：默认 `true`，`false` 时才产出调度器指令）。
