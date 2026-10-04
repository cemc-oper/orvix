# 指令参考

orvix 通过在脚本头部的 `#ORVIX` 注释指令描述作业的资源需求与调度选项。
本页是 `#ORVIX` 指令的完整参考。

## 语法规则

`#ORVIX` 指令必须出现在脚本的**初始注释块**中（在第一个非空、非注释行之前）：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=demo
#ORVIX nodes=2
#ORVIX time=01:00:00
#ORVIX comment="long running benchmark"
#ORVIX exclusive

set -euo pipefail
echo "running"
```

| 语法 | 含义 |
|---|---|
| `#ORVIX key=value` | 带值的指令 |
| `#ORVIX key` | 无值指令（如布尔标志 `exclusive`） |
| `#ORVIX key="x y"` 或 `'x y'` | 包含空格的值必须用双引号或单引号包裹 |

注意事项：

- `#ORVIX` 必须为大写，`#` 后不能有空格。`# ORVIX`、`#orvix` 和 `#ORVIXFOO` 均会被忽略。
- 解析在第一个非空、非注释行处停止，之后的 `#ORVIX` 行不生效。
- 指令必须为 `key=value` 或裸 `key` 形式；`key value`（无 `=`）是错误。
- 不在指令集合内的未知键会被**静默丢弃**，不会传递给调度器。

## 指令全集

orvix 共识别 19 个指令键。该集合与 takflow 的 jobspec 契约保持一一对应
（见 [jobspec 契约](../reference/jobspec.md)）。

### 调度器与标识

| 指令 | 说明 | 示例 |
|---|---|---|
| `scheduler` | 选择调度器后端（`slurm`、`donau` 或 `local`），默认 `local` | `scheduler=slurm` |
| `job-name` | 作业名称 | `job-name=myjob` |
| `queue` | 分区 / 队列 | `queue=normal` |

### 计算资源

| 指令 | 说明 | 示例 |
|---|---|---|
| `nodes` | 节点数量 | `nodes=2` |
| `ntasks` | 任务总数 | `ntasks=4` |
| `ntasks-per-node` | 每节点任务数 | `ntasks-per-node=2` |
| `cpus-per-task` | 每任务 CPU 数 | `cpus-per-task=4` |
| `time` | 时间限制（`HH:MM:SS`） | `time=01:00:00` |
| `memory` | 内存需求 | `memory=16G` |
| `exclusive` | 独占节点访问 | `exclusive` |
| `nodelist` | 指定节点列表 | `nodelist=node[01-04]` |

### I/O

| 指令 | 说明 | 示例 |
|---|---|---|
| `output` | 标准输出文件 | `output=job.out` |
| `error` | 标准错误文件 | `error=job.err` |

### 作业控制

| 指令 | 说明 | 示例 |
|---|---|---|
| `account` | 计费账户 | `account=proj01` |
| `dependency` | 作业依赖 | `dependency=afterok:12345` |
| `job-type` | 调度器作业类型（目前仅 Donau 使用） | `job-type=cosched` |
| `requeue` | 是否允许调度器自动重排队，默认 `true`；`false` 时 SLURM 生成 `--no-requeue` | `requeue=false` |

### 平台必填字段（CMA HPC）

| 指令 | 说明 | 示例 |
|---|---|---|
| `project` | 项目任务号，由 HPC 管理员提供 | `project=105-01-01` |
| `application` | 应用名称，由 HPC 管理员提供。使用 `modelname` 查看可用选项，如 `GRAPES`、`MCV` 等 | `application=GRAPES` |

## 后端条件指令

如果同一脚本在不同后端需要不同值，可使用 `[scheduler=<name>]` 条件前缀：

```bash
#ORVIX queue=normal
#ORVIX [scheduler=slurm] account=slurm_proj
#ORVIX [scheduler=donau] account=donau_proj
```

在上例中，`queue=normal` 适用于所有后端；`account` 的值根据当前后端选择。
条件行可以覆盖之前同键的无条件指令。

注意事项：

- 目前只支持 `[scheduler=...]` 这一种条件。
- 条件的求值以后端的最终决定为准：脚本中的 `scheduler=` 指令，或命令行 `--scheduler` 覆盖值。
