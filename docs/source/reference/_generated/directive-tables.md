<!-- 本文件由 docs/tools/generate_directive_tables.py 生成，请勿手工编辑。
     键集合、类型与默认值：改 spec/jobspec/jobspec.schema.json
     中文说明、示例与后端摘要：改 docs/data/directives.json
     重新生成：make docs-directives；CI 用 make docs-check 校验漂移 -->

## 调度与标识

| 指令 | 类型/默认 | 说明 | 示例 |
|---|---|---|---|
| `scheduler` | 字符串，默认 `local` | 目标后端：`slurm` / `donau` / `local` | `scheduler=slurm` |
| `job-name` | 字符串 | 作业名称 | `job-name=fcst` |
| `queue` | 字符串 | 分区 / 队列（规范键名是 `queue`，旧的 `partition` 已不存在，会被静默丢弃） | `queue=normal` |

## 计算资源

| 指令 | 类型/默认 | 说明 | 示例 |
|---|---|---|---|
| `nodes` | 字符串 | 节点数量 | `nodes=2` |
| `ntasks` | 字符串 | 任务总数 | `ntasks=8` |
| `ntasks-per-node` | 字符串 | 每节点任务数 | `ntasks-per-node=4` |
| `cpus-per-task` | 字符串 | 每任务 CPU 数 | `cpus-per-task=4` |
| `time` | 字符串 | 运行时限；`HH:MM:SS`，Donau 另接受 `MM:SS` 与纯秒 | `time=01:00:00` |
| `memory` | 字符串 | 内存需求 | `memory=16G` |
| `exclusive` | 裸键或带值 | 独占节点访问 | `exclusive` |
| `nodelist` | 字符串 | 指定节点列表 | `nodelist=node[01-04]` |

## 输入输出

| 指令 | 类型/默认 | 说明 | 示例 |
|---|---|---|---|
| `output` | 字符串 | 标准输出文件 | `output=job.out` |
| `error` | 字符串 | 标准错误文件；未设置时 stderr 并入 `output` | `error=job.err` |

## 作业控制

| 指令 | 类型/默认 | 说明 | 示例 |
|---|---|---|---|
| `account` | 字符串 | 计费账户 | `account=proj01` |
| `dependency` | 字符串 | 作业依赖表达式 | `dependency=afterok:12345` |
| `job-type` | 字符串 | 调度器作业类型（仅 Donau 使用） | `job-type=cosched` |
| `requeue` | 布尔，默认 `true` | 是否允许调度器自动重排队；仅 `false` 时 SLURM 生成 `--no-requeue` | `requeue=false` |

## 平台必填字段（CMA HPC）

| 指令 | 类型/默认 | 说明 | 示例 |
|---|---|---|---|
| `project` | 字符串 | 项目任务号，由 HPC 管理员提供 | `project=105-01-01` |
| `application` | 字符串 | 应用名称，由 HPC 管理员提供；可用 `modelname` 查看可选值 | `application=MCV` |

## 后端支持矩阵

| 指令 | SLURM | Donau | local |
|---|---|---|---|
| `scheduler` | 选择后端 | 选择后端 | 选择后端 |
| `job-name` | `--job-name` | `-n` | 忽略 |
| `queue` | `--partition` | `-q` | 忽略 |
| `nodes` | `--nodes` | `-nn` | 忽略 |
| `ntasks` | `--ntasks` | **忽略** | 忽略 |
| `ntasks-per-node` | `--ntasks-per-node` | `-tpn` | 忽略 |
| `cpus-per-task` | `--cpus-per-task` | 合并进 `-R` | 忽略 |
| `time` | `--time` | `-T`（可转秒） | 忽略 |
| `memory` | `--mem` | 合并进 `-R` | 忽略 |
| `exclusive` | `--exclusive` | `--exclusive` | 忽略 |
| `nodelist` | `--nodelist` | `-pn` | 忽略 |
| `output` | `--output` | `-oo` | 由后端重定向 stdout |
| `error` | `--error` | `-eo` | 由后端重定向 stderr |
| `account` | `--account` | `-A` | 忽略 |
| `dependency` | `--dependency` | **忽略** | 忽略 |
| `job-type` | **忽略** | `--job_type` | 忽略 |
| `requeue` | `false` → `--no-requeue` | **忽略** | 忽略 |
| `project` | `--wckey` | 合并进 `-d` | 忽略 |
| `application` | `--comment` | 合并进 `-d` | 忽略 |
