# 监视与终止作业

本页讲作业提交之后怎么跟踪状态、怎么安全地终止，以及各后端的行为差异。
状态字符串的完整对照见 [调度器映射](../reference/scheduler-mapping.md)。

## 查询状态

```bash
orvix status myjob.sh.info.yaml                  # 调度器原始状态
orvix status --format json myjob.sh.info.yaml    # 原始 + 归一化状态
```

几点实际有用的行为：

- **SLURM 有回退**：先 `squeue`，查不到或报错时回退 `sacct`，因此刚结束、已离开队列的
  作业也能查到最终状态；两者都失败才报错。
- **原始状态可能为空**：`squeue` 与 `sacct` 都没有内容时原始状态是空串，归一化为
  `UNKNOWN`（常见于作业号很旧、记账已清理的情况）。
- **local 只有两种状态**：`RUNNING` 与 `FINISHED`，归一化后分别是 `RUNNING` 与
  `COMPLETED`，无法反映脚本是否成功。
- 状态字符串里的限定后缀（如 sacct 的 `CANCELLED by 30255`）在 SLURM 归一化时只取首个词。

## 持续监视

```bash
orvix watch myjob.sh.info.yaml            # 默认每 5 秒一轮
orvix watch -i 2s myjob.sh.info.yaml      # 自定义轮询间隔
```

每轮打印一行带时间戳的**归一化状态**，到达终止态（COMPLETED / FAILED / CANCELLED /
TIMEOUT）即退出。

两个边界情况：

- `UNKNOWN` **不是**终止态，状态一直无法识别时 watch 会一直跑；需要限时就用
  `timeout 30m orvix watch ...`。
- 状态查询本身失败时打印 `[时间戳] ERROR: ...` 并以非零码退出；
  而作业失败/取消/超时到达终止态时退出码是 **0**。判断成败请读 `status --format json`，
  示例见 [提交与运行作业](running-jobs.md)。

## 终止作业

```bash
orvix kill myjob.sh.info.yaml              # 后端默认行为
orvix kill -s KILL myjob.sh.info.yaml      # 显式指定信号（Donau 不支持）
```

| 后端 | 默认行为 | 适用建议 |
|---|---|---|
| `slurm` | 分阶段优雅终止：先 TERM 到整个作业，等待宽限期，再兜底 `scancel` | 默认即可，让脚本的清理/trap 有机会执行 |
| `donau` | `djob -T <作业号>` | 无信号可选，指定 `-s` 会报错 |
| `local` | 向整个进程组发 SIGTERM | 默认即可；需要强杀用 `-s KILL` |

对已结束的作业重复执行 `kill` 视为成功，可以放心在清理脚本里调用。

### 让脚本有机会清理

SLURM 的默认 kill 用 `scancel --full --signal=TERM`，TERM 会同时送达 batch 脚本及其子进程，
所以脚本里的 trap 能正常执行（裸 `scancel` 只信号 batch shell，脚本前台等待子进程时
trap 会被推迟到来不及执行）：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal

trap 'echo "caught TERM, cleaning up"; rm -f "$TMPDIR/lock"; exit 143' TERM

echo "running"
sleep 3600
```

宽限期默认 30 秒，清理逻辑较慢时可以延长：

```bash
ORVIX_SLURM_KILL_GRACE=120 orvix kill myjob.sh.info.yaml
```

PENDING / CONFIGURING 的作业直接走 `scancel`（尚未运行，无需优雅终止）；
信号阶段失败而作业仍存活时也会立即升级为 `scancel`。完整过程见
[调度器映射](../reference/scheduler-mapping.md)。

### 与 ecFlow 的配合

ecFlow 任务被杀时正是靠这条链路：`ECF_KILL_CMD` 调用 `orvix kill`，
SLURM 的 TERM 触发 `head.h` 里的 trap，上报 `ecflow_client --abort`。
配置方法见 [与 ecFlow 集成](ecflow.md)。

## 终止之后检查什么

```bash
# 看 orvix 侧状态
orvix status --format json myjob.sh.info.yaml

# 集群侧复核（确认没有孤儿作业），把 <作业号> 换成实际值
squeue -j <作业号>
sacct -j <作业号> --format=JobID,JobName,State
```

如果 `.info.yaml` 已经丢失，orvix 无法定位作业，只能直接用调度器原生命令
（`scancel <作业号>` / `djob -T <作业号>`）处理，见 [故障排查](troubleshooting.md)。

## 相关页面

- [命令参考](../reference/cli.md) — status / watch / kill 的选项与退出码
- [环境变量与默认值](../reference/configuration.md)
- [故障排查](troubleshooting.md)
