# 提交与运行作业

本页讲日常怎么提交作业，以及如何在脚本与 CI 中使用 orvix。
命令的完整选项与退出码定义见 [命令参考](../reference/cli.md)。

## 四种用法怎么选

| 用法 | 何时用 | 产物 |
|---|---|---|
| `orvix generate 脚本` | 只想要翻译后的脚本：提交前自检、交付前预生成、在无集群机器上验证映射 | `.submit` |
| `orvix submit --dry-run 脚本` | 想在终端直接看翻译结果，不落任何文件 | 无（写 stdout） |
| `orvix submit 脚本` | 正常提交，马上拿到作业号 | `.submit`、`.info.yaml`、`.submit.log` |
| `orvix submit --watch 脚本` | 交互式等待结果，或串行流程中的简单等待 | 同上，并持续打印状态 |

```bash
orvix submit myjob.sh                 # 打印作业号
orvix submit --watch myjob.sh         # 提交并监视到结束
```

## 自定义产物路径

默认产物写在原脚本旁边，重复提交会覆盖。同一脚本要并发提交多份、或想把产物集中管理时，
显式指定路径：

```bash
orvix submit myjob.sh \
  --output-script runs/exp01/myjob.submit \
  --output-info   runs/exp01/myjob.info.yaml
```

注意 `.submit.log` 的路径固定为 `<脚本>.submit.log`，不随 `--output-script` 改变；
并发提交同一脚本时它会互相覆盖，排错时以最新一次为准。

## 在脚本与 CI 中使用

**取作业号**：`submit` 成功时 stdout 只有一行作业号。

```bash
job_id=$(orvix submit myjob.sh)
```

**判断作业成败**：不要用 `watch` / `submit --watch` 的退出码——作业 FAILED、CANCELLED、
TIMEOUT 时它们同样返回 0。正确做法是等待结束后读归一化状态：

```bash
#!/bin/bash
set -euo pipefail

orvix submit myjob.sh
orvix watch myjob.sh.info.yaml

state=$(orvix status --format json myjob.sh.info.yaml \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["state"])')

if [ "$state" != "COMPLETED" ]; then
    echo "job failed with state=$state" >&2
    exit 1
fi
```

**加整体超时**：`watch` 没有内置超时，且状态持续为 `UNKNOWN` 时不会自行结束。
需要限时等待时用外部超时控制：

```bash
timeout 30m orvix watch myjob.sh.info.yaml
```

**local 后端的限制**：local 只有"运行中/已结束"两种状态，脚本非零退出也归一化为
`COMPLETED`，上面的判断对 local 作业无效。需要验证本地脚本逻辑时，直接检查脚本输出，
或在脚本内部自行 `exit` 并断言产物。

## 提交失败时看什么

1. `.submit.log`：`[submit-cmd]` 是实际执行的调度器命令，`[output]` 是错误信息；
2. `.submit`：确认翻译结果是否符合预期（提交失败时它通常已经写出）；
3. `orvix generate --debug ...`：重放解析与后端选择过程，能看到被丢弃的未知键。

更多症状见 [故障排查](troubleshooting.md)。

## 相关页面

- [监视与终止作业](monitoring-and-killing.md)
- [生成的文件与管理](output-files.md)
- [命令参考](../reference/cli.md)
