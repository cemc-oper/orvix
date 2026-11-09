# 生成的文件与管理

orvix 默认把附属文件写在原脚本**同一目录**。本页讲这些文件怎么用、怎么管；
字段与格式的精确定义见 [文件格式](../reference/file-formats.md)。

## 产物速览

```
myjob.sh              原始脚本（orvix 不修改）
myjob.sh.submit       翻译后的脚本，实际被调度器/本地执行
myjob.sh.info.yaml    作业元数据，status / watch / kill 的输入
myjob.sh.submit.log   提交日志（成功与失败都写）
```

- `generate` 只写 `.submit`（失败时仍会写 `.submit.log`）。
- `submit` 成功时三个都写；提交失败时 `.submit` 已存在，`.info.yaml` 还没有。
- `--no-log` 关闭 `.submit.log`；`--dry-run` 不落任何文件。
- 各场景的完整对照表见 [文件格式](../reference/file-formats.md)。

## 日常怎么用

**`.submit`：确认翻译结果。** 提交被拒或资源不符预期时先看它，例如检查 `#SBATCH --partition`
是否真的是你想要的队列、有没有空旗标。

**`.info.yaml`：作业的句柄。** 后续所有操作都以它为参数，不要删、不要改名；
它也记录了生效的指令列表，可用于追溯某次提交到底用了什么资源。

**`.submit.log`：排错第一站。** `[submit-cmd]` 是实际执行的调度器命令，`[output]`
是调度器返回的作业号或错误信息。

## 覆盖与并发

- 重复提交同一脚本会**覆盖**同名 `.submit`、`.info.yaml`、`.submit.log`，
  旧作业号随之丢失；旧作业本身仍在集群上，需要时可从集群侧用作业号管理。
- 同一脚本**并发提交多份**时产物会互相覆盖。需要并行多实例时，给每份实例独立的
  输出路径，或把脚本复制到各自目录：

```bash
orvix submit myjob.sh \
  --output-script runs/exp01/myjob.submit \
  --output-info   runs/exp01/myjob.info.yaml
```

- `.submit.log` 路径固定为 `<脚本>.submit.log`，无法按实例区分，并发时以最新一次为准。

## 清理与归档

- 作业结束后，`.submit` 与 `.submit.log` 可以安全删除；**`.info.yaml` 建议保留到
  你确定不再需要查状态或终止该作业为止**。
- 需要审计或复盘时，把 `.info.yaml` 与 `.submit.log` 随作业输出一起归档——它们记录了
  orvix 版本、提交命令、提交主机与生效指令。
- 在 ecFlow 流程中这些文件位于 `%ECF_HOME%` 下的任务工作目录，通常随流程的
  housekeep 一起清理；清理前确认对应任务不再需要 kill。

## 文件丢失的后果

| 丢失文件 | 影响 | 处置 |
|---|---|---|
| `.info.yaml` | `status` / `watch` / `kill` 全部无法使用 | 用调度器原生命令按作业号处理；下次提交会重新生成 |
| `.submit` | 无法核对翻译结果，正在运行的作业不受影响 | 用 `orvix generate` 重新生成 |
| `.submit.log` | 失去提交期错误记录 | 重新提交或用 `--debug` 重放 |

## 相关页面

- [文件格式](../reference/file-formats.md)
- [提交与运行作业](running-jobs.md)
- [故障排查](troubleshooting.md)
