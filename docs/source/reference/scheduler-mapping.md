# 调度器映射

本页是三个后端的**指令输出格式**与**状态映射**的权威定义。
各键的语义见 [指令参考](directives.md)。

`scheduler` 键只用于选择后端，自身不翻译成任何调度器指令行。

| 后端 | 提交命令 | 状态查询 | 终止命令 |
|---|---|---|---|
| `slurm` | `sbatch --parsable <脚本>` | `squeue -j <id> -h -o %T`，失败或为空回退 `sacct -j <id> -n -X -o State` | `scancel`（分阶段，见下） |
| `donau` | `dsub -s <脚本>` | `djob -L <id>.0`，取 `STATE` 字段 | `djob -T <id>` |
| `local` | 直接启动子进程，作业号为 PID | 对 PID 发信号 0 探测 | 向进程组发信号 |

## SLURM

翻译结果为 `#SBATCH` 行，按固定的内部顺序输出（**与脚本中指令的书写顺序无关**），
未设置的键不产生任何行。

| orvix 指令 | 输出 | 值处理 |
|---|---|---|
| `job-name` | `#SBATCH --job-name=<值>` | 含空白自动加双引号 |
| `output` | `#SBATCH --output=<值>` | 含空白加双引号 |
| `error` | `#SBATCH --error=<值>` | 含空白加双引号 |
| `nodes` | `#SBATCH --nodes=<值>` | 原样 |
| `ntasks` | `#SBATCH --ntasks=<值>` | 原样 |
| `ntasks-per-node` | `#SBATCH --ntasks-per-node=<值>` | 原样 |
| `cpus-per-task` | `#SBATCH --cpus-per-task=<值>` | 原样 |
| `time` | `#SBATCH --time=<值>` | 原样 |
| `queue` | `#SBATCH --partition=<值>` | 含空白加双引号 |
| `account` | `#SBATCH --account=<值>` | 含空白加双引号 |
| `project` | `#SBATCH --wckey=<值>` | 含空白加双引号 |
| `application` | `#SBATCH --comment=<值>` | 含空白加双引号 |
| `exclusive` | `#SBATCH --exclusive` 或 `--exclusive=<值>` | 裸键 / 带值 |
| `nodelist` | `#SBATCH --nodelist=<值>` | 含空白加双引号 |
| `memory` | `#SBATCH --mem=<值>` | 原样 |
| `dependency` | `#SBATCH --dependency=<值>` | 原样 |
| `requeue=false` | `#SBATCH --no-requeue` | 仅显式 `false`（大小写不敏感）时生成；`true` 或未设置时不生成 |
| `job-type` | 无对应指令 | 忽略 |

**空值会生成空旗标**：`#ORVIX project=` 产出 `#SBATCH --wckey=`。不需要的键请整行删除。

`sbatch --parsable` 返回的 `<jobid>;<cluster>` 形式只取作业号部分。

### 默认 kill 的分阶段过程

1. 若作业为 PENDING / CONFIGURING，直接 `scancel <作业号>`（控制器可能对排队作业拒绝信号阶段）。
2. 否则 `scancel --full --signal=TERM <作业号>`：`--full` 把 TERM 同时送达 batch 脚本及其子进程，
   使脚本的信号 trap（例如 ecFlow `head.h` 的 `ecflow_client --abort`）有机会执行；
   裸 `scancel` 只信号 batch shell，脚本前台等待子进程时 trap 会被推迟。
3. 每秒轮询一次状态，直到终止态或超过宽限期（默认 30 秒，`ORVIX_SLURM_KILL_GRACE` 可改，0 表示立即升级）。
4. 仍未结束则兜底 `scancel <作业号>`（控制器取消路径，作业标记为 CANCELLED）。

信号阶段失败而作业仍存活时立即升级为普通 `scancel`；状态查询已查不到作业时视为终止成功。
显式 `orvix kill -s <信号>` 保持单次发送语义，同样附加 `--full`。

## Donau

翻译结果为 `#DSUB` 行。**空值一律跳过**，不产生空旗标（与 SLURM 相反）。

| orvix 指令 | 输出 | 值处理 |
|---|---|---|
| `job-name` | `#DSUB -n <值>` | 原样 |
| `output` | `#DSUB -oo <值>` | 原样 |
| `error` | `#DSUB -eo <值>` | 原样 |
| `nodes` | `#DSUB -nn <值>` | 原样 |
| `ntasks-per-node` | `#DSUB -tpn <值>` | 原样 |
| `queue` | `#DSUB -q <值>` | 原样 |
| `account` | `#DSUB -A <值>` | 原样 |
| `nodelist` | `#DSUB -pn '<值>'` | 单引号包裹 |
| `job-type` | `#DSUB --job_type <值>` | 原样 |
| `time` | `#DSUB -T <秒>` | 纯整数视为秒；`MM:SS`、`HH:MM:SS` 转秒；其他形式（如 `8h`）原样传递 |
| `cpus-per-task` + `memory` | `#DSUB -R "cpu=<X>;mem=<Y>"` | 两者合并为一行，分号连接；只设置其一时只含对应部分 |
| `project` + `application` | `#DSUB -d "<项目>:<应用>"` | 两者合并为一行；只设置其一时 `-d "<其一>"` |
| `exclusive` | `#DSUB --exclusive` 或 `--exclusive <值>` | 裸键 / 带值 |
| `ntasks` / `dependency` / `requeue` | 无对应指令 | 忽略 |

`dsub` 的输出为表格，作业号取 `JOBID` 表头下一行的首字段；解析不到时回退为
"首字段是纯数字的行"。`orvix kill -s` 在 Donau 上直接报错，不支持自定义信号。

## local

不经过调度器，直接执行翻译后的脚本，因此**不产生任何调度器指令行**。

| orvix 指令 | 行为 |
|---|---|
| `output` | 启动子进程时把 stdout 重定向到该路径（创建并截断） |
| `error` | stderr 重定向到该路径；未设置或与 `output` 相同时并入 `output`（与 SLURM 语义一致） |
| 两者都未设置 | 继承父进程标准流 |
| 其余资源键 | 忽略 |

作业号为子进程 PID，子进程被放入独立进程组（pgid == pid）。
状态只能区分 `RUNNING` 与 `FINISHED`（信号 0 探测），**无法获知退出码**：
脚本以非零码结束同样归一化为 `COMPLETED`。
默认 kill 向整个进程组发 SIGTERM，对引入进程组机制之前提交的旧作业回退到单 PID。

## 状态映射

归一化状态共七种，其中 **COMPLETED、FAILED、CANCELLED、TIMEOUT 为终止态**；
`UNKNOWN` 不是终止态（持续 UNKNOWN 时 `watch` 不会结束）。

| 归一化状态 | SLURM 原始状态 | Donau 原始状态 | local |
|---|---|---|---|
| `PENDING` | PENDING、CONFIGURING | PENDING、WAITING、QUEUED、CONFIGURING | — |
| `RUNNING` | RUNNING | RUNNING | RUNNING |
| `COMPLETED` | COMPLETED | COMPLETED、DONE、FINISHED | FINISHED |
| `FAILED` | FAILED、NODE_FAIL、BOOT_FAIL、DEADLINE、PREEMPTED、OUT_OF_MEMORY、OOM | FAILED、FAILURE、ABORTED、BOOT_FAIL、NODE_FAIL、DEADLINE、PREEMPTED、OUT_OF_MEMORY、OOM | — |
| `CANCELLED` | CANCELLED | CANCELLED、CANCELED | — |
| `TIMEOUT` | TIMEOUT | TIMEOUT | — |
| `UNKNOWN` | 其他 | 其他 | 其他 |

归一化前的预处理两后端不同：

- **SLURM**：大写化并去首尾空白，取空白前的首个词，再去掉词尾的 `+` / `~` 后缀——因此
  sacct 的 `CANCELLED by 30255` 与 SLURM 20.11 的 `CANCELLED+` 都归一化为 `CANCELLED`。
- **Donau**：仅大写化，不做去空白与取首词。
- SLURM 的 `squeue` 与 `sacct` 都查不到内容时原始状态为空串，归一化结果为 `UNKNOWN`。

`orvix status` 默认输出**原始状态**，`orvix status --format json` 同时给出
`state`（归一化）与 `raw_state`（原始）；`orvix watch` 输出的是归一化状态。

## 相关页面

- [指令参考](directives.md) — 19 个指令键的语义与后端支持矩阵
- [命令参考](cli.md) — 各命令的选项、输出与退出码
- [环境变量与默认值](configuration.md) — `ORVIX_SLURM_KILL_GRACE` 等

核对：`internal/scheduler/{slurm,donau,local,state}.go`；
可用 `orvix generate --scheduler <后端> <脚本>` 实测指令输出。
