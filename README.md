# orvix

orvix 是一个将脚本任务提交到 HPC 集群上运行的命令行工具。

只需要在脚本头部用统一的 `#ORVIX key=value` 语法写明资源需求，orvix 会自动把它翻译成目标调度系统 (如 SLURM) 的指令并提交作业。

## 安装

```bash
make build
```

二进制文件生成到 `bin/orvix`，可将其加入 `PATH`。

## 快速开始

在脚本头部添加 `#ORVIX` 指令：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=demo
#ORVIX nodes=2
#ORVIX time=01:00:00
#ORVIX exclusive

echo "Hello from HPC"
```

使用 orvix submit 命令提交脚本：

```bash
$ orvix submit myjob.sh
12345678
```

该命令会生成两个文件：

- myjob.sh.submit：用于提交到作业队列系统的脚本
- myjob.sh.info.yaml：记录作业信息，包括作业 ID

使用 orivx status 命令查询作业运行状态：

```bash
$ orvix status myjob.sh.info.yaml
RUNNING
```

使用 orvix kill 命令终止作业：

```bash
$ orvix kill myjob.sh.info.yaml
```

## 支持的调度后端

| 后端 | 说明 |
|---|---|
| `slurm` | 翻译为 `#SBATCH` 指令，通过 `sbatch` 提交到 SLURM 集群 |
| `local` | 在本机直接以子进程方式执行，用于本地测试 |

## 指令语法

`#ORVIX` 指令必须写在脚本最开头的注释段里（第一个非空非注释行之前）：

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

### 写法规则

| 写法 | 含义 |
|---|---|
| `#ORVIX key=value` | 带值的指令 |
| `#ORVIX key` | 无值的指令（如 `exclusive` 等布尔选项） |
| `#ORVIX key="x y"` 或 `'x y'` | 含空格的值用双/单引号包裹 |

注意：

- `#ORVIX` 必须大写且 `#` 后无空格。`# ORVIX`、`#orvix`、`#ORVIXFOO` 都会被忽略
- 解析在第一个非空、非注释行处停止
- 必须写成 `key=value` 形式，`key value`（无等号）会报错

### 常用指令

#### 调度与标识

| 指令 | 说明 | 示例 |
|---|---|---|
| `scheduler` | 选择调度后端（`slurm` 或 `local`），默认 `local` | `scheduler=slurm` |
| `job-name` | 作业名称 | `job-name=myjob` |
| `queue` | 分区/队列 | `queue=normal` |

#### 计算资源

| 指令 | 说明 | 示例 |
|---|---|---|
| `nodes` | 节点数 | `nodes=2` |
| `ntasks` | 总任务数 | `ntasks=4` |
| `ntasks-per-node` | 每节点任务数 | `ntasks-per-node=2` |
| `cpus-per-task` | 每任务 CPU 数 | `cpus-per-task=4` |
| `time` | 运行时限（`HH:MM:SS`） | `time=01:00:00` |
| `memory` | 内存需求 | `memory=16G` |
| `exclusive` | 独占节点 | `exclusive` |
| `nodelist` | 指定节点 | `nodelist=node[01-04]` |

#### 输入输出

| 指令 | 说明 | 示例 |
|---|---|---|
| `output` | 标准输出文件 | `output=job.out` |
| `error` | 标准错误文件 | `error=job.err` |

#### 作业控制

| 指令 | 说明 | 示例 |
|---|---|---|
| `account` | 计费账户 | `account=proj01` |
| `dependency` | 作业依赖 | `dependency=afterok:12345` |

#### 平台必填项（CMA HPC）

| 指令 | 说明 | 示例 |
|---|---|---|
| `project` | 项目编号 | `project=105-01-01` |
| `application` | 应用名称 | `application=GRAPES` |

### 后端条件指令

如果同一脚本需要在不同后端使用不同值，可以用 `[scheduler=<name>]` 条件前缀：

```bash
#ORVIX queue=normal
#ORVIX [scheduler=slurm] account=slurm_proj
#ORVIX [scheduler=donau] account=donau_proj
```

上面例子中，`queue=normal` 对所有后端生效；`account` 的值则根据当前后端选择对应的行。条件行可以覆盖之前无条件设置的同名指令。

## 命令

### `orvix submit <脚本>`

解析脚本中的 `#ORVIX` 指令，生成翻译后的脚本并提交。

```bash
$ orvix submit case/job/serial/orvix_serial.sh
12345678
```

选项：

```bash
orvix submit --dry-run script.sh   # 仅打印翻译后脚本，不实际提交
orvix submit --scheduler=slurm script.sh  # 强制指定调度后端，覆盖脚本中的设置
```

### `orvix status <info.yaml>`

查询作业状态。

```bash
$ orvix status case/job/serial/orvix_serial.info.yaml
RUNNING
```

### `orvix kill <info.yaml>`

终止作业。

```bash
$ orvix kill case/job/serial/orvix_serial.info.yaml
```

## 生成的文件

提交 `path/to/script.sh` 后，会在**同目录**生成两个文件：

```
path/to/script.sh             # 原始脚本（不变）
path/to/script.sh.submit      # 翻译后真正执行的脚本
path/to/script.sh.info.yaml      # 作业元数据（status/kill 需要用到）
```

重复提交会覆盖同名文件。脚本无扩展名时，生成脚本默认补 `.sh`。

`info.yaml` 示例：

```yaml
scheduler: slurm
job_id: "12345678"
submitted_at: 2026-05-10T11:27:22.107722252Z
script_source: /abs/path/script.sh
script_generated: /abs/path/script.sh.submit
submit_dir: /abs/path
hostname: login01
user: wangdp
directives:
    - key: scheduler
      value: slurm
    - key: queue
      value: normal
    - key: nodes
      value: "2"
```

## 许可证

orvix 采用 [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) 开源许可证。
