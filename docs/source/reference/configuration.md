# 环境变量与默认值

## 环境变量

orvix 本体只识别一个环境变量：

| 变量 | 作用域 | 说明 |
|---|---|---|
| `ORVIX_SLURM_KILL_GRACE` | `orvix kill`（SLURM 后端，未指定 `-s` 时） | TERM 信号阶段之后的等待宽限期，单位**秒**，默认 30；`0` 表示不等待、立即升级为普通 `scancel`；非法值被忽略并回退到默认值（`--debug` 有日志） |

```bash
ORVIX_SLURM_KILL_GRACE=60 orvix kill myjob.sh.info.yaml
```

分阶段终止的完整过程见 [调度器映射](scheduler-mapping.md)。

> ecFlow 集成中常见的 `%ORVIX_BIN%`、`%ORVIX_SCHEDULER%` 等是 **ecFlow suite 变量**，
> 由 suite 定义提供，不是 orvix 读取的环境变量。

## 默认值汇总

| 项 | 默认值 | 修改方式 |
|---|---|---|
| 调度器后端 | `local` | 脚本 `scheduler=` 或命令行 `--scheduler` |
| `submit --watch-interval` / `watch --interval` | `5s` | 对应命令选项 |
| `status --format` | `text` | `--format json` |
| `submit` 的提交日志 | 开启（成功/失败都写） | `--no-log` 关闭 |
| 生成脚本路径 | `<脚本>.submit` | `--output-script` |
| 元数据 sidecar 路径 | `<脚本>.info.yaml` | `--output-info` |
| 提交日志路径 | `<脚本>.submit.log` | 固定，不可配置 |
| 生成脚本权限 | 0755 | 固定 |
| `.info.yaml` / `.submit.log` 权限 | 0644 | 固定 |
| `requeue` | `true`（保持调度器默认行为） | `#ORVIX requeue=false` |
| SLURM 默认 kill 宽限期 | 30 秒 | `ORVIX_SLURM_KILL_GRACE` |
| SLURM kill 期间的状态轮询步长 | 1 秒 | 固定 |
| 指令解析 / 脚本改写的单行长度上限 | 1 MiB | 固定 |

## 相关页面

- [命令参考](cli.md) — 各命令选项的完整列表
- [指令参考](directives.md) — 指令键的类型与默认值
- [调度器映射](scheduler-mapping.md) — 后端行为差异

核对：`internal/scheduler/slurm.go`、`internal/directive/parser.go`、`cmd/*.go`、`internal/submit/`。
