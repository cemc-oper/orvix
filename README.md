# orvix

orvix 是一个把脚本任务提交到 HPC 上运行的命令行工具。它解析脚本头部的
`#ORVIX` 预处理指令，将其翻译为目标作业调度系统的指令（如 SLURM 的
`#SBATCH`），生成可执行脚本并提交。

## 特性

- 用户用统一的 `#ORVIX key=value` 语法写脚本，不用关心后端
- 当前支持的调度后端：
  - **slurm** —— 翻译为 `#SBATCH --key=value` 后通过 `sbatch` 提交
  - **local** —— 直接在本机以子进程方式执行
- 每次提交都在原脚本同目录留下生成脚本和 YAML 元数据，便于审计与重放
- 单二进制，运行期只依赖调度系统自身的工具（`sbatch` / `squeue` / `scancel`）

## 编译

```bash
make           # 构建到 bin/orvix
make test      # 运行单元测试
make help      # 列出全部 Makefile 目标
```

需要 Go 1.26+。依赖通过 `go mod tidy` 自动拉取（cobra、yaml.v3）。

## 命令

### `orvix submit <script>`

读取脚本，解析 `#ORVIX` 指令，在原脚本同目录生成翻译后的脚本并提交，
打印 jobID。

```bash
$ orvix submit case/job/serial/orvix_serial.sh
12345678
```

可选 flag：

```bash
orvix submit --dry-run script.sh   # 仅打印翻译后脚本，不落盘也不提交
```

### `orvix status <info.yaml>`

读取 submit 时生成的 `info.yaml`（包含 scheduler 与 job_id），
查询作业状态。

```bash
$ orvix status case/job/serial/orvix_serial.info.yaml
RUNNING

$ orvix status case/job/local/orvix_local.info.yaml
FINISHED
```

SLURM 后端先查 `squeue`，作业已结束时回退到 `sacct`。

### `orvix kill <info.yaml>`

```bash
orvix kill case/job/serial/orvix_serial.info.yaml
orvix kill case/job/local/orvix_local.info.yaml
```

## 指令语法

每行一个 `#ORVIX` 指令，必须出现在脚本注释段（即第一个非空非注释行之前）：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX partition=normal
#ORVIX job-name=demo
#ORVIX nodes=2
#ORVIX time=01:00:00
#ORVIX comment="long running benchmark"
#ORVIX exclusive

set -euo pipefail
echo "running"
```

| 写法 | 含义 |
|---|---|
| `#ORVIX key=value` | 带值 |
| `#ORVIX key` | 裸 key（无值，对应 SLURM 布尔 flag 比如 `--exclusive`） |
| `#ORVIX key="x y"` 或 `'x y'` | 含空格的值需用双/单引号包裹 |

注意：

- marker 严格匹配 `#ORVIX`（大写、`#` 后无空格）。`# ORVIX` / `#orvix` /
  `#ORVIXFOO` 都会被忽略
- 解析在第一个非注释、非空行处停止
- `key value`（无 `=`）会报错，请用 `key=value`

特殊指令：

- `#ORVIX scheduler=slurm` 或 `scheduler=local`，决定使用哪个调度后端。
  缺省为 `local`

## 生成的文件

提交 `path/to/script.sh` 后，在**同目录**生成两个 sidecar：

```
path/to/script.sh             # 原始（不动）
path/to/script.submit.sh      # 翻译后真正执行的脚本
path/to/script.info.yaml      # 元数据；status/kill 直接读这个文件
```

重复提交会覆盖同名 sidecar（不再保留历史时间戳）。脚本无扩展名时，
生成脚本默认补 `.sh`。

YAML 内容示例：

```yaml
scheduler: slurm
job_id: "12345678"
submitted_at: 2026-05-10T11:27:22.107722252Z
script_source: /abs/path/script.sh
script_generated: /abs/path/script.submit.sh
submit_dir: /abs/path
hostname: login01
user: wangdp
directives:
    - key: scheduler
      value: slurm
    - key: partition
      value: normal
    - key: nodes
      value: "2"
```

字段含义：

- `scheduler` —— `slurm` 或 `local`
- `job_id` —— SLURM 是 sbatch 返回的作业号；local 是子进程 PID
- `submit_dir` —— orvix 调用时的工作目录，等同于 `SLURM_SUBMIT_DIR`
- `directives` —— 解析到的所有 `#ORVIX key=value`，按出现顺序保留

## 项目结构

```
orvix/
├── main.go                          # 入口
├── cmd/                             # cobra 子命令
│   ├── root.go
│   ├── submit.go
│   ├── status.go
│   └── kill.go
├── internal/
│   ├── directive/                   # #ORVIX 指令解析
│   ├── scheduler/                   # 调度后端抽象
│   │   ├── scheduler.go             # Scheduler 接口
│   │   ├── slurm.go                 # sbatch / squeue / scancel
│   │   └── local.go                 # fork+exec
│   ├── script/                      # 脚本翻译生成
│   └── jobinfo/                     # YAML 元数据
├── examples/hello.sh
├── Makefile
└── go.mod
```

## 调度后端

### slurm

每个 `#ORVIX key=value` 翻译为 `#SBATCH --key=value`（含空格的值自动加双
引号）。`#ORVIX scheduler=slurm` 本身不会被翻译。生成脚本通过
`sbatch --parsable` 提交，返回 jobID；`status` 先查 `squeue`，作业已结束时
回退到 `sacct`；`kill` 调用 `scancel`。

### local

不生成 preamble，翻译后的脚本即原脚本去掉 `#ORVIX` 行。通过 `exec` 在
后台启动子进程，返回 PID；`status` 通过 `kill -0` 探测进程存活；`kill`
调用 `Process.Kill()`。

## 添加新调度后端

实现 `internal/scheduler/scheduler.go` 中的 `Scheduler` 接口：

```go
type Scheduler interface {
    Name() string
    PreambleFor(d *directive.Set) ([]string, error)
    Submit(scriptPath string) (string, error)
    Status(jobID string) (string, error)
    Kill(jobID string) error
}
```

然后在同文件的 `ByName()` 里注册名字到构造器的映射即可。
