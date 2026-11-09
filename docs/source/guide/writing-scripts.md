# 编写作业脚本

本页讲怎么把资源需求写进脚本。19 个指令键的完整定义见
[指令参考](../reference/directives.md)，这里只讲写法与常见配方。

## 最小骨架

`#ORVIX` 指令写在脚本**开头的注释区**里，在第一条可执行语句之前：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=fcst
#ORVIX nodes=2
#ORVIX ntasks-per-node=4
#ORVIX time=01:00:00
#ORVIX output=fcst.out
#ORVIX error=fcst.err

set -euo pipefail
echo "running"
```

要点：

- 标记必须是大写 `#ORVIX`，`#` 后不能有空格；`# ORVIX`、`#orvix` 都会被忽略。
- 允许缩进，但**必须在第一条非注释、非空行之前**；写在 `echo` 之后的指令不生效。
- 含空格的值建议加引号：`#ORVIX job-name="my demo job"`。
- 布尔标志写成裸键：`#ORVIX exclusive`。
- 指令**不做变量展开**：`#ORVIX output=$HOME/job.out` 里的 `$HOME` 会原样传给调度器。
  需要按环境变化时，在生成脚本的外层（工作流生成器、ecFlow 的 `%VAR%` 预处理）展开，
  或直接使用相对路径。

写完后先翻译检查一遍，不必真的提交：

```bash
orvix generate --scheduler slurm --output-script /tmp/check.submit myjob.sh
cat /tmp/check.submit
```

## 常用配方

**串行作业（SLURM）**

```bash
#ORVIX scheduler=slurm
#ORVIX queue=serial
#ORVIX nodes=1
#ORVIX ntasks=1
#ORVIX cpus-per-task=1
#ORVIX time=00:10:00
#ORVIX output=serial.%j.out
```

**MPI 并行作业（SLURM）**

```bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX nodes=4
#ORVIX ntasks-per-node=32
#ORVIX time=02:00:00
#ORVIX exclusive
```

**大内存作业**

```bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX nodes=1
#ORVIX ntasks=1
#ORVIX memory=64G
#ORVIX requeue=false
```

`requeue=false` 让 SLURM 生成 `--no-requeue`，避免内存不足导致的失败作业被自动重跑。
Donau 上内存与 CPU 会合并成一条 `-R "cpu=...;mem=..."`，写法相同。

**同一脚本适配多个后端**

```bash
#ORVIX queue=normal
#ORVIX [scheduler=slurm] account=slurm_proj
#ORVIX [scheduler=donau] account=donau_proj
#ORVIX [scheduler=donau] job-type=cosched
```

无条件行对所有后端生效，条件行只在指定后端生效，并可覆盖同键的无条件行。
目前只支持 `[scheduler=...]` 这一种条件。

## 常见写法错误

| 写法 | 后果 | 正确做法 |
|---|---|---|
| `#ORVIX partition=normal` | 旧键名，静默丢弃，不生成 `--partition` | 用 `queue=normal` |
| 指令写在 `echo` / `set -e` 之后 | 不生效，且该行会从 `.submit` 中被删掉 | 全部放在开头注释区 |
| `#ORVIX project=`（空值） | SLURM 生成空的 `--wckey=`，可能被集群拒绝 | 不需要就整行删除，或用条件指令限定后端 |
| `#ORVIX output=$HOME/x.out` | 变量不展开，路径字面化 | 用相对路径，或在外层展开 |
| `#ORVIX nodes 2`（缺少 `=`） | 解析报错，提交失败 | `nodes=2` |
| 期望未知键被透传给调度器 | 未知键一律丢弃 | 只用契约内的 19 个键 |

更多"现象 → 原因 → 处置"见 [故障排查](troubleshooting.md)。

## 相关页面

- [指令参考](../reference/directives.md) — 全部键的语义、默认值与后端支持矩阵
- [选择调度器后端](choosing-scheduler.md) — 后端差异与平台必填字段
- [故障排查](troubleshooting.md)
