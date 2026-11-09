# 快速上手

本教程用 **`local` 后端**（直接在本地运行，无需调度器）带你跑通
生成 → 提交 → 查状态 → 监视 → 终止的完整流程，在任何 Linux 机器上都可以照着做。
最后给出切换到真实集群的方法。

前置条件：已安装 orvix，见 [安装](installation.md)。

## 1. 编写作业脚本

```bash
mkdir -p orvix-demo && cd orvix-demo
```

创建 `hello.sh`，头部用 `#ORVIX` 指令描述需求：

```bash
#!/bin/bash
#ORVIX scheduler=local
#ORVIX job-name=hello
#ORVIX output=hello.out

echo "Hello from $(hostname)"
sleep 5
echo "=== done ==="
```

`output=hello.out` 让作业的标准输出写入 `hello.out`（`local` 后端由 orvix 自己重定向）。

## 2. 生成翻译后的脚本（不提交）

```bash
$ orvix generate hello.sh
Generated: /abs/path/orvix-demo/hello.sh.submit
```

`local` 后端没有调度器指令，生成的脚本只是删掉了 `#ORVIX` 行：

```bash
#!/bin/bash
echo "Hello from $(hostname)"
sleep 5
echo "=== done ==="
```

用 `--scheduler` 可以在**没有集群的机器上**预览其他后端的翻译结果，
配合 `--output-script` 避免覆盖已有文件：

```bash
$ orvix generate --scheduler slurm --output-script hello.slurm.submit hello.sh
Generated: hello.slurm.submit
$ cat hello.slurm.submit
#!/bin/bash
#SBATCH --job-name=hello
#SBATCH --output=hello.out

echo "Hello from $(hostname)"
sleep 5
echo "=== done ==="
```

## 3. 提交作业

```bash
$ orvix submit hello.sh
3334563
```

输出的数字是作业号（`local` 后端即进程 PID）。同目录下现在有四类文件：

```
hello.sh              原始脚本（不变）
hello.sh.submit       翻译后的脚本，实际被执行的就是它
hello.sh.info.yaml    作业元数据，后续命令都以它为输入
hello.sh.submit.log   提交日志
hello.out             作业输出（由 output= 指定）
```

## 4. 查询与监视状态

```bash
$ orvix status hello.sh.info.yaml
RUNNING
```

`status` 默认打印调度器原始状态；加 `--format json` 可同时拿到归一化状态：

```bash
$ orvix status --format json hello.sh.info.yaml
{"job_id":"3334563","scheduler":"local","state":"COMPLETED","raw_state":"FINISHED","terminal":true}
```

持续监视直到结束：

```bash
$ orvix watch hello.sh.info.yaml
[2026-11-09T13:56:20+08:00] RUNNING
[2026-11-09T13:56:25+08:00] RUNNING
[2026-11-09T13:56:29+08:00] COMPLETED
```

也可以在提交时一步完成：`orvix submit --watch hello.sh`。

作业结束后查看输出：

```bash
$ cat hello.out
Hello from DESKTOP-7PJNF2G
=== done ===
```

> **注意退出码语义**：`watch` 与 `submit --watch` 在作业失败、被取消、超时时同样返回 0；
> `local` 后端甚至无法区分成功与失败（脚本非零退出也显示 `COMPLETED`）。
> 要在脚本或 CI 里判断作业成败，请读取 `status --format json` 的 `state` 字段，
> 详见 [命令参考](../reference/cli.md)。

## 5. 终止作业

创建一个运行时间较长的脚本 `long.sh`：

```bash
#!/bin/bash
#ORVIX scheduler=local
#ORVIX job-name=long
#ORVIX output=long.out

echo "started"
sleep 300
echo "finished"
```

提交后立即终止：

```bash
$ orvix submit long.sh
3334657
$ orvix kill long.sh.info.yaml
$ orvix status --format json long.sh.info.yaml
{"job_id":"3334657","scheduler":"local","state":"COMPLETED","raw_state":"FINISHED","terminal":true}
$ cat long.out
started
```

`long.out` 只有 `started`，说明作业在 `sleep` 期间被终止。
`kill` 是幂等的，对已结束的作业重复执行不会报错。

## 6. 切换到真实集群

> **HPC-only**：以下步骤需要 SLURM 或 Donau 环境。

把脚本头部的指令换成集群对应的值，其余内容不变：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=hello
#ORVIX output=hello.out
#ORVIX nodes=1
#ORVIX ntasks=1
#ORVIX time=00:05:00
#ORVIX project=105-01-01
#ORVIX application=MCV

echo "Hello from $(hostname)"
```

然后使用同样的命令：

```bash
orvix submit --watch hello.sh
```

`queue`、`project`、`application` 的取值由平台决定（CMA HPC 上后两者通常必填）。
切换到 Donau 只需把 `scheduler` 改为 `donau` 并调整队列名；
各后端的指令映射与状态差异见 [调度器映射](../reference/scheduler-mapping.md)。

## 下一步

- [指令参考](../reference/directives.md) — 全部 19 个指令键与后端支持矩阵
- [命令参考](../reference/cli.md) — 各命令的选项、输出与退出码
- [文件格式](../reference/file-formats.md) — `.submit` / `.info.yaml` / `.submit.log` 的完整格式
