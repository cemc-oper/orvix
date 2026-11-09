# 指令参考

本页列出 orvix 识别的全部 19 个 `#ORVIX` 指令键，以及它们在三个后端的支持情况。
指令的权威定义是 `spec/jobspec/jobspec.schema.json`（见 [jobspec 契约](jobspec.md)），
本页是它的命令行投影。

## 语法速查

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name="my demo job"
#ORVIX [scheduler=donau] queue=donau_queue
#ORVIX exclusive

echo "running"
```

| 规则 | 说明 |
|---|---|
| 书写形式 | `#ORVIX key=value` 或裸键 `#ORVIX key`（布尔标志） |
| 标记 | 恰好大写 `#ORVIX`，`#` 后无空格；`ORVIX` 后须是空白或行尾，故 `#ORVIXFOO` 无效 |
| 行首空白 | 允许缩进 |
| 生效范围 | 从文件开头扫描，遇到第一个非空、非注释行即停止；其后的指令不生效 |
| 含空白的值 | 建议用双引号或单引号包裹（不加引号的 `key=a b` 也会拼接为 `a b`，但裸键后有多余词会报错） |
| 条件指令 | `[scheduler=<name>]` 前缀，只在该后端生效；可覆盖同键的无条件指令 |
| 覆盖 | 同键后出现者覆盖先出现者 |
| 未知键 | **静默丢弃**，不透传给调度器；加 `--debug` 可看到丢弃日志 |
| 其他 | 无转义处理、无环境变量展开；引号未闭合、`[` 未闭合、裸键后有多余词均为解析错误 |

后端的选择顺序：命令行 `--scheduler` > 脚本中第一条无条件 `scheduler=` > 默认 `local`。

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

**不要留空值**：SLURM 后端会把 `#ORVIX project=` 翻译成空的 `#SBATCH --wckey=`，
可能导致提交被集群拒绝（Donau 后端相反，空值直接跳过该行）。不需要这两个字段时，
整行删掉或用条件指令限定后端。

## 后端支持矩阵

| 指令 | SLURM | Donau | local |
|---|---|---|---|
| `scheduler` | 选择后端 | 选择后端 | 选择后端 |
| `job-name` | `--job-name` | `-n` | 忽略 |
| `output` | `--output` | `-oo` | 由后端重定向 stdout |
| `error` | `--error` | `-eo` | 由后端重定向 stderr |
| `nodes` | `--nodes` | `-nn` | 忽略 |
| `ntasks` | `--ntasks` | **忽略** | 忽略 |
| `ntasks-per-node` | `--ntasks-per-node` | `-tpn` | 忽略 |
| `cpus-per-task` | `--cpus-per-task` | 合并进 `-R` | 忽略 |
| `time` | `--time` | `-T`（可转秒） | 忽略 |
| `queue` | `--partition` | `-q` | 忽略 |
| `account` | `--account` | `-A` | 忽略 |
| `project` | `--wckey` | 合并进 `-d` | 忽略 |
| `application` | `--comment` | 合并进 `-d` | 忽略 |
| `exclusive` | `--exclusive` | `--exclusive` | 忽略 |
| `nodelist` | `--nodelist` | `-pn` | 忽略 |
| `job-type` | **忽略** | `--job_type` | 忽略 |
| `memory` | `--mem` | 合并进 `-R` | 忽略 |
| `dependency` | `--dependency` | **忽略** | 忽略 |
| `requeue` | `false` → `--no-requeue` | **忽略** | 忽略 |

"忽略"表示该键会被正常解析并记录在 `.info.yaml` 中，但不产生任何调度器指令。
生成行的完整格式、引号与合并规则见 [调度器映射](scheduler-mapping.md)。

## 相关页面

- [调度器映射](scheduler-mapping.md) — 各后端的指令输出格式与状态映射
- [文件格式](file-formats.md) — 解析后的指令如何记录在 `.info.yaml` 中
- [jobspec 契约](jobspec.md) — 指令集的权威定义与版本

核对：`spec/jobspec/jobspec.schema.json`、`internal/directive/parser.go`、`internal/scheduler/`。
