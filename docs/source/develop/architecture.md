# 架构说明

本页面向需要修改 orvix 本身的开发者。日常使用请从 [快速上手](../start/quickstart.md) 开始。

## 代码布局

```
cmd/                  cobra 命令定义（submit / generate / status / watch / kill / version）
spec/jobspec/         jobspec 契约权威所在：schema、VERSION、conformance 向量与
                      golden，以及 go:embed 访问包（spec.go）
internal/
├── directive/        #ORVIX 指令解析器
├── scheduler/        调度器后端接口与实现（slurm / donau / local）
├── script/           脚本改写（生成 .submit 内容）
├── submit/           提交管线编排
├── jobinfo/          .info.yaml sidecar 读写
├── watch/            状态轮询
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

`orvix generate` 只执行 1–3 并写出 `.submit` 文件。两者的产物差异：

| 阶段 | generate | submit |
|---|---|---|
| 解析 / 选后端 / 渲染 | 是 | 是 |
| 写 `.submit` | 是 | 是（`--dry-run` 时改写 stdout） |
| 调用调度器提交 | 否 | 是 |
| 写 `.info.yaml` | 否 | 是 |
| 写 `.submit.log` | 仅失败时 | 成功与失败都写（`--no-log` 关闭） |
| 轮询状态 | 否 | 仅 `--watch` |

产物格式见 [文件格式](../reference/file-formats.md)。

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

用户视角的语法规则见 [编写作业脚本](../guide/writing-scripts.md)，
指令全集见 [指令参考](../reference/directives.md)。实现上需要记住的三点：

- 解析在第一个非空、非注释行处停止；行首空白允许；`#ORVIX` 后必须跟空白或行尾。
- 条件指令 `#ORVIX [scheduler=<name>] key=value` 在解析期按最终后端过滤，
  同键后出现者覆盖先出现者。
- 未知键静默丢弃，不做透传；`KnownDirectives` 由 `spec/jobspec` 的 schema 派生校验
  （`spec_test.go`），不会与契约漂移（见 [jobspec 契约](../reference/jobspec.md)）。

## 测试约定

- 测试使用 `stretchr/testify`（`assert` / `require`），测试文件与被测代码同目录。
- 完整的分层策略（单元 / golden / CI / HPC 实跑）与各层跑法见 [测试策略](testing.md)。
