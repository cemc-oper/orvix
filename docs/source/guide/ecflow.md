# 与 ecFlow 集成

ecFlow 负责任务编排与状态跟踪，orvix 负责把作业脚本里的 `#ORVIX` 指令翻译成调度器指令
并提交/终止作业。两者通过 `ECF_JOB_CMD` 与 `ECF_KILL_CMD` 两个变量对接。

> **HPC-only**：完整集成需要 ecFlow server 与真实调度器。没有集群时，仍可用
> `orvix generate` 对 ecFlow 生成的作业脚本做静态检查。

## 集成原理

```
ecflow_server
  │  ECF_JOB_CMD  = %ORVIX_BIN% submit --scheduler %ORVIX_SCHEDULER% %ECF_JOB%
  │  ECF_KILL_CMD = %ORVIX_BIN% kill %ECF_JOB%.info.yaml
  ▼
orvix submit 读取 %ECF_JOB%（head/tail 已展开、变量已替换的作业脚本）
  解析 #ORVIX 指令 → 生成 %ECF_JOB%.submit → sbatch/dsub 提交
  → 生成 %ECF_JOB%.info.yaml（作业元数据 sidecar）
  ▼
作业在计算节点运行：head.h 的 ecflow_client --init 置为 active，
  tail.h 的 --complete（或 ERR trap 的 --abort）回报最终状态
  ▼
kill 时：orvix kill 读取 .info.yaml 中的 scheduler + job_id，
  向对应调度器发取消（scancel / djob -T）
```

**sidecar 是 kill 链路的关键**：ecFlow 只知道任务路径（如 `/suite/family/task`），
不知道调度器作业号。submit 时 orvix 把 `scheduler`、`job_id` 等写进
`%ECF_JOB%.info.yaml`，`ECF_KILL_CMD` 据此找到并终止正确的作业。
sidecar 缺失（submit 失败、被清理）时 kill 无从下手——这是 kill 链路排查的第一检查点。

## suite 中的配置

```bash
edit ORVIX_BIN       'orvix'      # server 进程 PATH 找不到时改成绝对路径
edit ORVIX_SCHEDULER 'slurm'      # 或 donau

edit ECF_JOB_CMD  '%ORVIX_BIN% submit --scheduler %ORVIX_SCHEDULER% %ECF_JOB%'
edit ECF_KILL_CMD '%ORVIX_BIN% kill %ECF_JOB%.info.yaml'
```

部署要点：

- `ORVIX_BIN` 必须对 **ecflow_server 进程**可用；不确定时写绝对路径。
- `ECF_HOME` 目录要提前创建，server 不会自动创建。
- 作业脚本（`.ecf`）里的 `#ORVIX` 指令同样遵守解析规则：写在开头注释区，
  用 `queue=` 而不是已废弃的 `partition=`。
- 平台需要 `project` / `application` 时再启用对应行；**空值会在 SLURM 上生成
  `#SBATCH --wckey=` 这样的空旗标**，可能直接导致提交被拒。

## 作业头尾（head.h / tail.h）的要求

orvix 不管 ecFlow 的状态上报，头尾文件仍按 ecFlow 惯例编写。与 kill 链路相关的两点：

1. `--init` 时把调度器作业号作为 `ECF_RID` 上报，便于在 ecFlow 与调度器之间对账：

   ```bash
   export ECF_RID="${SLURM_JOB_ID:-$$}"
   ecflow_client --init="$ECF_RID"
   ```

2. 用 ERR trap 上报失败，并保证脚本能响应 TERM——SLURM 的默认 kill 会先
   `scancel --full --signal=TERM`，给 trap 留出执行时间（默认 30 秒宽限期）：

   ```bash
   set -eEuo pipefail
   trap 'rc=$?; ecflow_client --abort="task failed rc=$rc"; exit "$rc"' ERR
   ```

   `tail.h` 在正常结束时调用 `ecflow_client --complete`。

清理逻辑特别长的任务，可在 suite 的 kill 命令中调大宽限期，例如
`ORVIX_SLURM_KILL_GRACE=120 %ORVIX_BIN% kill %ECF_JOB%.info.yaml`。

## 运行后检查什么

任务运行后，`%ECF_HOME%` 下应能看到：

```
task.job1             ecFlow 生成的作业脚本（head/tail 已展开）
task.job1.submit      orvix 翻译后的脚本：含 #SBATCH/#DSUB 行、无 #ORVIX 行
task.job1.info.yaml   作业元数据 sidecar
task.1.out            作业输出（ECF_JOBOUT）
```

## 集成排错

| 现象 | 检查点 |
|---|---|
| 任务 submitted 后不动 | `ORVIX_BIN` 是否在 server 进程 PATH 中；`%ECF_JOB%.submit` 是否生成 |
| 作业被调度器拒绝 | 队列名是否合法；`.submit.log` 中的错误；是否有空旗标 |
| `--kill` 后调度器作业仍在 | `.info.yaml` 是否存在且 `job_id` 正确；用 `squeue -u $USER` 复核 |
| 任务一直 active 不变 complete | `tail.h` 是否被包含；作业输出里是否有 `ecflow_client` 报错 |

更多通用症状见 [故障排查](troubleshooting.md)。

## 可运行示例

工作区的 `runtime/orvix-devel/repo/orvix-hpc-test/examples/ecflow/` 提供一个最小套件
（`suite.def` + `include/head.h` / `tail.h` + 两个任务：串行任务与供手动 kill 的长任务），
其 `docs/ecflow.md` 给出部署步骤与手动验证 runbook，包括 kill 链路的逐步检查方法。

## 相关页面

- [监视与终止作业](monitoring-and-killing.md) — SLURM 分阶段 kill 与 trap 的关系
- [被工作流生成器使用](workflow-generators.md) — takflow 等生成器如何产出 `#ORVIX` 指令
- [文件格式](../reference/file-formats.md)
