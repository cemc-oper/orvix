# 选择调度器后端

orvix 支持 `slurm`、`donau`、`local` 三个后端。本页帮你选定后端并处理平台差异；
指令到调度器指令的逐键映射见 [调度器映射](../reference/scheduler-mapping.md)。

## 后端如何确定

优先级从高到低：

1. 命令行 `--scheduler=<名称>`（最高，适合临时切换与翻译检查）；
2. 脚本中第一条无条件的 `#ORVIX scheduler=<名称>`；
3. 默认 `local`。

```bash
orvix submit --scheduler donau myjob.sh     # 不改脚本，直接换后端
```

## 三个后端怎么选

| 后端 | 适用场景 | 主要限制 |
|---|---|---|
| `slurm` | SLURM 集群上的生产作业 | 需要 `sbatch` / `squeue` / `sacct` / `scancel` 可用 |
| `donau` | 华为 Donau 集群上的生产作业 | 不支持 `ntasks`、`dependency`、`requeue`；kill 不能指定信号 |
| `local` | 开发机/WSL 上调试脚本与流程、无调度器环境 | 无资源控制；状态只有运行中/已结束，**拿不到退出码** |

`local` 后端是调试利器：同一份脚本先在本地跑通逻辑，再把 `scheduler` 与队列相关指令
换成集群值即可提交到生产集群。但它**不能用来验证资源映射是否正确**——资源键在 local
下全部被忽略，验证翻译请用 `orvix generate --scheduler <目标后端>`。

## CMA HPC 平台字段

CMA 集群通常要求显式的项目与应用标记：

| 指令 | 说明 |
|---|---|
| `queue` | 队列/分区名，按平台填写 |
| `project` | 项目任务号，由管理员提供 |
| `application` | 应用名称，可用 `modelname` 查看可选值 |

它们在 SLURM / Donau 上分别映射成什么旗标，见
[调度器映射](../reference/scheduler-mapping.md)。

```bash
#ORVIX scheduler=slurm
#ORVIX queue=serial
#ORVIX project=105-04-01
#ORVIX application=GRAPES
```

两个注意点：

- **不要留空值**：SLURM 会生成 `#SBATCH --wckey=` 这类空旗标，可能被集群拒绝；
  Donau 则直接跳过空值行。不需要时把整行删掉。
- 不确定平台是否强制校验时，先看 `.submit.log` 里的调度器报错，再与管理员确认队列与项目号。

## 提交前自检

不提交就能确认最终生效的指令与翻译结果：

```bash
# 看翻译后的调度器指令
orvix generate --scheduler slurm --output-script /tmp/check.submit myjob.sh
cat /tmp/check.submit

# 解析过程有疑问时打开调试日志，能看到被丢弃的未知键与实际后端
orvix generate --debug --scheduler slurm --output-script /tmp/check.submit myjob.sh
```

## 相关页面

- [调度器映射](../reference/scheduler-mapping.md) — 指令输出格式与状态映射
- [编写作业脚本](writing-scripts.md) — 条件指令与常用配方
- [故障排查](troubleshooting.md) — 提交被拒绝、状态查不到等问题
