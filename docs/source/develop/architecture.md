# 架构说明

本页面向需要修改 orvix 本身的开发者。日常使用请从 [快速上手](../quickstart.md) 开始。

## 代码布局

```
cmd/                  cobra 命令定义（submit / generate / status / watch / kill / version）
internal/
├── directive/        #ORVIX 指令解析器
├── scheduler/        调度器后端接口与实现（slurm / donau / local）
├── script/           脚本改写（生成 .submit 内容）
├── submit/           提交管线编排
├── jobinfo/          .info.yaml sidecar 读写
├── watch/            状态轮询
├── spec/             vendor 的 takflow jobspec 契约（只读）
├── version/          版本推导
└── log/              调试日志（--debug）
```

## 提交管线

`internal/submit/submit.go` 是核心编排器，`orvix submit` 的完整流程：

1. **解析** — `directive.Parse` 从脚本头部提取 `#ORVIX key=value` 行
   （在第一个非空、非注释行处停止），过滤条件指令与未知键。
2. **选择后端** — `scheduler.For` 根据 `scheduler=` 指令或 `--scheduler` 覆盖值
   选定 slurm / donau / local 后端。
3. **改写脚本** — `script.Render` 生成可运行脚本：保留 shebang、插入调度器指令块、
   删除 `#ORVIX` 行、其余内容原样保留。
4. **提交** — 后端执行原生提交命令（如 `sbatch --parsable`），捕获作业 ID。
5. **记录** — `jobinfo.Write` 在原脚本旁写入 `.info.yaml` sidecar
   （作业 ID、调度器、提交命令、解析后的指令等）。
6. **监视**（可选）— `--watch` 时 `watch.Run` 轮询状态直到终止状态。

`orvix generate` 只执行 1–3 并写出 `.submit` 文件。

## 调度器后端接口

所有后端实现 `scheduler.Scheduler` 接口（`internal/scheduler/scheduler.go`）：

```go
type Scheduler interface {
    Name() string
    PreambleFor(d *directive.Set) ([]string, error)
    Submit(scriptPath string) (jobID string, command string, err error)
    Status(jobID string) (string, error)
    Kill(jobID string, sig os.Signal) error
    NormalizeState(raw string) JobState
}
```

- `PreambleFor` 负责指令翻译，返回调度器指令行（如 `#SBATCH` 行）。
- `NormalizeState` 把调度器原生状态归一化为 `JobState`
  （PENDING / RUNNING / COMPLETED / FAILED / CANCELLED / TIMEOUT / UNKNOWN），
  `watch` 只依赖归一化状态判断终止。

## Generator 模式

每个后端用一组 `Generator` 函数构建指令块：

```go
type Generator func(d *directive.Set) (string, bool)
```

每个 generator 查询指令集中自己负责的键，命中则产出一行调度器指令，
未命中返回 `("", false)` 被跳过。新增指令映射是纯声明式的：
在 `PreambleFor` 的 generator 列表里加一项即可。

部分键需要定制 generator 而不是简单的一对一映射，例如 Donau 后端：

- `cpus-per-task` + `memory` 合并为一条 `-R` 行；
- `project` + `application` 合并为一条 `-d` 行；
- `time` 的 `HH:MM:SS` 自动转换为秒。

## 指令解析规则要点

- 指令必须在脚本初始注释块内；标记为恰好大写 `#ORVIX`，`#` 与 `ORVIX` 之间无空格。
- 支持条件指令 `#ORVIX [scheduler=<name>] key=value`（目前仅支持 scheduler 条件），
  条件指令覆盖同键的无条件指令。
- 未知键静默丢弃，不做透传。指令全集与 takflow jobspec 契约一致
  （见 [jobspec 契约](../reference/jobspec.md)）。

## 测试约定

- 测试使用 `stretchr/testify`（`assert` / `require`）。
- 解析器测试（`internal/directive/parser_test.go`）覆盖引号值、裸键、条件过滤、
  覆盖行为、畸形输入等边界。
- 各后端测试验证指令映射与特殊逻辑（如 Donau 的时间转秒）。
- 运行：`make test`，单包 `make test PKG=./internal/directive`。
