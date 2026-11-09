# 故障排查

按"现象 → 原因 → 处置"组织。先记住两个通用手段：

```bash
# 看解析与后端选择过程，包括被丢弃的未知键、实际执行的调度器命令
orvix generate --debug --output-script /tmp/check.submit myjob.sh
orvix submit  --debug myjob.sh

# 不提交就能看到翻译结果
cat /tmp/check.submit
```

提交类问题的第一手证据始终是 `<脚本>.submit.log` 与 `<脚本>.submit`。

## 指令与解析

| 现象 | 原因 | 处置 |
|---|---|---|
| 分区/队列没有生效，`.submit` 里没有 `--partition` | 用了旧键名 `partition=`，已被静默丢弃 | 改为 `queue=` |
| 某些指令完全没生效 | 写在第一条可执行语句之后，解析已停止 | 把指令移到开头注释区 |
| 同上，且指令行在 `.submit` 里消失了 | 正常行为：orvix 删除全文所有 `#ORVIX` 行，包括未生效的 | 同上 |
| 提交报 `expected key=value or bare key` | 裸键后面跟了多余的词，或 `key value` 缺少 `=` | 改成 `key=value`；含空格的值加引号 |
| 报 `unterminated "-quote` / `unclosed [` | 引号或条件括号未闭合 | 修正该行 |
| 输出路径里出现字面的 `$HOME` / `${VAR}` | 指令不做变量展开 | 用相对路径，或在生成脚本的外层展开 |
| 加了自定义 `#SBATCH` 用的键却没效果 | 未知键一律丢弃，不透传 | 只能使用契约内的 19 个键 |
| 条件指令没生效 | 条件后端与实际后端不一致；或条件键不是 `scheduler` | 用 `--debug` 确认最终后端；只支持 `[scheduler=...]` |

## 提交

| 现象 | 原因 | 处置 |
|---|---|---|
| `sbatch failed: ...` | 队列/账户/项目号不被集群接受，或资源超限制 | 看 `.submit.log` 的 `[output]`；核对 `queue`/`project`/`application` |
| `.submit` 里有 `--wckey=`、`--comment=` 空旗标 | 写了 `#ORVIX project=` 这类空值指令 | 删除整行，或用条件指令限定后端 |
| `unknown scheduler "xxx"` | `scheduler=` 拼写错误 | 只支持 `slurm` / `donau` / `local` |
| 提交失败但 `.submit` 已存在 | 正常顺序：先写脚本再提交 | 直接用该脚本排查，无需重新生成 |
| 没有 `.submit.log` | 用了 `--no-log`，或处于 `--dry-run` | 去掉 `--no-log` 重试 |
| 并发提交后日志/info 内容对不上 | 同一脚本的产物互相覆盖 | 每个实例用独立 `--output-script` / `--output-info` 或独立目录 |
| `dsub` 提交后作业号解析异常 | dsub 输出表格格式变化 | 用 `--debug` 看原始输出，按表格核对 |

## 状态与监视

| 现象 | 原因 | 处置 |
|---|---|---|
| `status` 输出为空或 json 中 `state=UNKNOWN` | 作业已离开队列且 `sacct` 也无记录（作业号很旧、记账未开启/已清理） | 用 `sacct -j <作业号>` 手工确认；以集群侧为准 |
| `watch` 一直不结束 | 状态持续 `UNKNOWN`，而它不是终止态；或作业确实还在运行 | 用 `timeout 30m orvix watch ...` 限时；查集群状态 |
| `watch` 返回 0 但作业失败了 | **设计如此**：到达任何终止态都返回 0 | 用 `status --format json` 的 `state` 判断成败 |
| local 作业显示 `COMPLETED`，但脚本明明失败 | local 只有运行中/已结束两种状态，拿不到退出码 | 检查脚本输出与产物；本地验证成败需脚本自行断言 |
| `status` 报 `missing job_id` / 读取失败 | `.info.yaml` 被删、被移动或损坏 | 用调度器原生命令按作业号查询；重新提交生成新 sidecar |

## 终止

| 现象 | 原因 | 处置 |
|---|---|---|
| `kill` 报错但作业其实已经没了 | 一般被幂等逻辑覆盖；若仍报错，看具体错误文本 | 用 `squeue -j <作业号>` / `sacct` 复核 |
| `kill` 之后脚本的 trap 没执行 | 显式 `-s` 用了不可捕获的信号（如 KILL），或清理超过宽限期 | 用默认 TERM 路径；调大 `ORVIX_SLURM_KILL_GRACE` |
| SLURM 上 `kill` 很快返回，作业却还在 | 处于 PENDING 时直接 `scancel`，或宽限期后兜底 `scancel` 需要时间 | 稍等并用 `squeue` 复核 |
| Donau 上 `kill -s TERM` 报错 | Donau 不支持自定义信号 | 去掉 `-s`，使用默认 `djob -T` |
| ecFlow 里任务已 aborted，集群作业仍在跑 | `.info.yaml` 缺失或 `job_id` 不对，kill 链路断在 orvix 之前 | 检查 sidecar；必要时手工 `scancel <作业号>` |

## 相关页面

- [编写作业脚本](writing-scripts.md) — 常见写法错误对照
- [监视与终止作业](monitoring-and-killing.md)
- [与 ecFlow 集成](ecflow.md) — 集成特有的检查点
- [文件格式](../reference/file-formats.md)
