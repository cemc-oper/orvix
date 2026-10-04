# 调度器后端

orvix 目前支持三个调度器后端：

| 后端 | 说明 |
|---|---|
| `slurm` | 转换为 `#SBATCH` 指令，通过 `sbatch` 提交到 SLURM 集群 |
| `donau` | 转换为 `#DSUB` 指令，通过 `dsub` 提交到华为 Donau 调度系统 |
| `local` | 作为本地子进程直接运行，适用于本地测试 |

后端由脚本的 `scheduler=` 指令或命令行 `--scheduler` 选项选择，默认 `local`。

## SLURM 后端

### 指令映射

| orvix 指令 | SLURM 指令 | 说明 |
|---|---|---|
| `job-name` | `--job-name` | 作业名称 |
| `output` | `--output` | 标准输出文件 |
| `error` | `--error` | 标准错误文件 |
| `nodes` | `--nodes` | 节点数量 |
| `ntasks` | `--ntasks` | 任务总数 |
| `ntasks-per-node` | `--ntasks-per-node` | 每节点任务数 |
| `cpus-per-task` | `--cpus-per-task` | 每任务 CPU 数 |
| `time` | `--time` | 时间限制（`HH:MM:SS`） |
| `queue` | `--partition` | 分区 / 队列 |
| `account` | `--account` | 计费账户 |
| `project` | `--wckey` | 项目任务号 |
| `application` | `--comment` | 应用名称 |
| `exclusive` | `--exclusive` | 独占节点访问 |
| `nodelist` | `--nodelist` | 指定节点列表 |
| `memory` | `--mem` | 内存需求 |
| `dependency` | `--dependency` | 作业依赖 |
| `requeue=false` | `--no-requeue` | 禁止自动重排队；`requeue` 为 `true` 或未设置时不生成任何指令 |

含空白的值（如 `job-name`）会自动加双引号。`job-type` 在 SLURM 中无对应指令，会被忽略。

### 提交与状态查询

- 提交使用 `sbatch --parsable`，返回的 `<jobid>;<cluster>` 形式只取作业 ID 部分。
- 状态查询先走 `squeue -j <id> -h -o %T`；查不到时回退到 `sacct -j <id> -n -X -o State`，
  因此刚结束、已离开队列的作业也能查到最终状态。

### 终止行为

默认（不带 `-s`）为分阶段优雅终止，让作业脚本的信号 trap
（如 ecFlow `head.h` 的 `ecflow_client --abort` 上报）有机会执行：

1. `scancel --full --signal=TERM <jobid>` —— TERM 同时送达 batch 脚本及其全部子进程
   （裸 `scancel` 只发给 batch shell 本身，脚本若在前台等待子进程，trap 会被推迟到
   子进程自行结束后才执行，通常来不及）；
2. 轮询作业状态，最多等待宽限期（默认 30 秒，可用 `ORVIX_SLURM_KILL_GRACE`
   环境变量覆盖，单位秒，`0` 表示立即升级）；
3. 超时仍未结束则兜底 `scancel <jobid>`（controller cancel 路径：SIGCONT+SIGTERM、
   KillWait 后 SIGKILL，并把作业标记为 CANCELLED）。

`orvix kill -s <信号>` 显式指定信号时保持单次发送语义，但同样附加 `--full`，
确保信号能到达 batch 脚本的子进程。

## Donau 后端

### 指令映射

| orvix 指令 | Donau 指令 | 说明 |
|---|---|---|
| `job-name` | `-n` | 作业名称 |
| `output` | `-oo` | 标准输出文件 |
| `error` | `-eo` | 标准错误文件 |
| `nodes` | `-nn` | 节点数量 |
| `ntasks-per-node` | `-tpn` | 每节点任务数 |
| `cpus-per-task` | `-R "cpu=X"` | 每任务 CPU 数（与 `memory` 合并为一条 `-R` 指令） |
| `memory` | `-R "mem=Y"` | 内存需求（与 `cpus-per-task` 合并为一条 `-R` 指令） |
| `time` | `-T` | 时间限制（`HH:MM:SS` 自动转换为秒；时长字符串如 `8h` 原样传递） |
| `queue` | `-q` | 队列 |
| `account` | `-A` | 账户 |
| `project` | `-d` | 项目号（与 `application` 合并为 `-d "project:application"`） |
| `application` | `-d` | 应用名称（与 `project` 合并为 `-d "project:application"`） |
| `exclusive` | `--exclusive` | 独占节点访问（可带值或不带值） |
| `nodelist` | `-pn` | 指定节点列表（自动加引号） |
| `job-type` | `--job_type` | 作业类型 |

注意：`ntasks` 和 `dependency` 在 Donau 中无对应指令，会被静默忽略；`requeue` 同样不生效。

### 提交与状态查询

- 提交使用 `dsub -s <脚本>`，从输出表格中提取作业 ID。
- 状态查询使用 `djob -L <jobid>.0`，解析输出中的 `STATE` 字段。

### 终止行为

使用 `djob -T <jobid>`，不接受自定义信号（`orvix kill -s` 会报错）。

## local 后端

`local` 后端不经过任何调度器，直接把翻译后的脚本作为本地子进程运行，适用于本地测试。

- **作业 ID** 为子进程的 PID。子进程被放入独立的进程组（pgid == pid）。
- **输出重定向**：本地没有调度器层来落实 `output` / `error` 指令，因此后端在启动子进程时
  自行把 stdout/stderr 重定向到这两个路径（语义与 SLURM 一致：未设置 `error` 时 stderr
  合并进 `output` 文件）。两个指令都未设置时继承父进程的标准流。
- **状态查询**只能区分 `RUNNING` 与 `FINISHED`（通过对 PID 发信号 0 探测），
  无法获知退出码。
- **终止**：默认向整个进程组发送 SIGTERM，让脚本的 trap / 清理逻辑有机会执行；
  对进程组机制引入之前提交的旧作业回退到单 PID 发信号。`kill -s` 可指定其他信号。

## 归一化状态

`orvix watch` 输出的是归一化状态（`orvix status` 输出调度器原始状态）。
六个归一化状态中，`COMPLETED`、`FAILED`、`CANCELLED`、`TIMEOUT` 为终止状态：

| 归一化状态 | SLURM 原始状态 | Donau 原始状态 | local |
|---|---|---|---|
| `PENDING` | PENDING、CONFIGURING | PENDING、WAITING、QUEUED、CONFIGURING | — |
| `RUNNING` | RUNNING | RUNNING | RUNNING |
| `COMPLETED` | COMPLETED | COMPLETED、DONE、FINISHED | FINISHED |
| `FAILED` | FAILED、NODE_FAIL、BOOT_FAIL、DEADLINE、PREEMPTED、OUT_OF_MEMORY | FAILED、FAILURE、ABORTED、BOOT_FAIL、NODE_FAIL、DEADLINE、PREEMPTED、OUT_OF_MEMORY | — |
| `CANCELLED` | CANCELLED | CANCELLED、CANCELED | — |
| `TIMEOUT` | TIMEOUT | TIMEOUT | — |
| `UNKNOWN` | 其他 | 其他 | 其他 |
