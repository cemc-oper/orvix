# 新增调度器后端

新增一个后端（例如接入新的调度系统）会同时触及**代码、契约与消费方**，
按下面顺序做，不要只做代码部分。改内部实现前请先读 [架构说明](architecture.md)。

## 1. 实现 Scheduler 接口

在 `internal/scheduler/<name>.go` 新建后端，实现 `scheduler.Scheduler`：

| 方法 | 实现要求 |
|---|---|
| `Name() string` | 规范后端名，小写，与契约 `scheduler` 枚举值一致 |
| `PreambleFor(d *directive.Set) ([]string, error)` | 把指令集翻译成该调度器的指令行，不带换行符 |
| `Submit(scriptPath string) (jobID, command string, err error)` | 执行原生提交命令，返回作业号与实际执行的命令字符串（会写进 `.info.yaml`） |
| `Status(jobID string) (string, error)` | 返回调度器**原始**状态字符串；已结束作业也尽量能查到（参考 SLURM 的 sacct 回退） |
| `Kill(jobID string, sig os.Signal) error` | `sig == nil` 时用后端默认行为；对已结束作业应视为成功（幂等）；不支持自定义信号时显式报错 |
| `NormalizeState(raw string) JobState` | 把原始状态映射到归一化状态，未知值返回 `UNKNOWN` |

## 2. 用 Generator 组织指令翻译

指令翻译是声明式的：为每个键写一个 `Generator`，在 `PreambleFor` 里按固定顺序列出。

```go
type Generator func(d *directive.Set) (string, bool)   // 未命中返回 ("", false)
```

参考现有后端的辅助函数与特殊处理：

- SLURM：`sbatch` / `sbatchQ`（含空白加引号）/ `sbatchBare`（裸旗标），
  以及 `requeue` 这类布尔反转的定制 generator；
- Donau：合并类指令（`cpus-per-task` + `memory` → 一条 `-R`，
  `project` + `application` → 一条 `-d`）、时间格式转换。

需要显式决定并记录：**每个键在新后端是映射还是忽略**；
空值是跳过还是生成空旗标（现有两后端行为相反，选哪种都可以，但必须在文档写清楚）。

## 3. 注册后端

- `internal/scheduler/scheduler.go` 的 `ByName`：在 switch 中加分支；
- 若后端需要在构造时读取指令集（例如 local 需要 `output`/`error` 来做输出重定向），
  在 `For` 中像 `NewLocal(d)` 那样处理；`ByName` 路径用于 status/kill，拿不到指令集，
  实现时不要依赖它一定存在。

## 4. 状态与终止语义

- 归一化状态只有七种，终止态是 COMPLETED / FAILED / CANCELLED / TIMEOUT；
  映射表在 `internal/scheduler/state.go` 与各后端的 `NormalizeState`。
- `watch` 只依赖归一化状态判断结束，所以**终止态映射不全会导致 watch 永远不退出**。
- 信号处理的平台相关代码放在 `signal_unix.go` / `signal_windows.go`、
  `local_unix.go` / `local_windows.go` 这类带 build tag 的文件中。

## 5. 契约与消费方

新后端名需要进入契约的 `scheduler` 枚举：

- `spec/jobspec/jobspec.schema.json` 的 `scheduler.enum` 加上新名字；
- 按语义化版本更新 `spec/jobspec/VERSION`（新增后端为 MINOR）；
- 合并发布后通知消费方升 pin：takflow（`scripts/sync_jobspec.py`）与
  orvix-hpc-test（`scripts/sync_conformance.sh`）。

## 6. 测试

- `internal/scheduler/<name>_test.go`：指令映射、合并/转换逻辑、状态归一化、kill 行为；
- `spec/jobspec/conformance_test.go` 的 `conformanceSchedulers` 加上新后端，
  并准备对应的 `conformance/golden/<name>/` 目录：

  ```bash
  make regen-golden     # 生成新后端的基线，diff 随同一 commit 评审
  ```

- CI（`.github/workflows/ci.yml`）与发布（`.goreleaser.yaml`）如需覆盖新平台，一并修改；
- 真实环境验证：在 `orvix-hpc-test` 增加平台配置与用例，实跑提交/状态/kill，
  见 [测试策略](testing.md)。

## 7. 文档

- `reference/directives.md`：后端支持矩阵加一列，逐键标注映射或忽略；
- `reference/scheduler-mapping.md`：新增后端小节（指令输出格式、提交/状态/终止命令、
  原始状态映射、预处理规则）；
- `reference/cli.md`、`reference/configuration.md`：kill 与信号相关行为、新环境变量；
- `guide/choosing-scheduler.md`：适用场景与限制；
- `guide/monitoring-and-killing.md`、`guide/troubleshooting.md`：行为差异与常见症状；
- 页面归属与同步要求见 [文档编写指南](documentation.md)。

## 相关页面

- [架构说明](architecture.md)
- [新增一个指令](adding-directive.md)
- [测试策略](testing.md)
