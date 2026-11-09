# 命令参考

orvix 提供六个功能子命令：`generate`、`submit`、`status`、`watch`、`kill`、`version`，
另有 cobra 自带的 `help` 与 `completion`（生成 shell 补全脚本）。

所有命令支持 `--debug`，向 stderr 输出调试日志（解析、后端选择、实际执行的调度器命令等）。
任何命令出错时，向 stderr 打印 `orvix: <错误>` 并以退出码 1 结束；成功为 0。

## orvix generate

```bash
orvix generate [flags] <脚本>
```

解析 `#ORVIX` 指令并生成翻译后的脚本，**不提交**，也不生成 `.info.yaml`。

| 选项 | 默认 | 说明 |
|---|---|---|
| `--scheduler=<名称>` | 空 | 强制指定后端，覆盖脚本中的 `scheduler=` |
| `--output-script=<路径>` | `<脚本>.submit` | 自定义生成脚本路径 |

输出与退出码：

- 成功：stdout 打印 `Generated: <绝对路径>`，退出 0。
- 失败：退出 1，**仍写出 `<脚本>.submit.log`** 记录错误（首行标签为 `ERROR: orvix submit failed`）。

```bash
$ orvix generate myjob.sh
Generated: /abs/path/myjob.sh.submit
```

## orvix submit

```bash
orvix submit [flags] <脚本>
```

生成翻译后的脚本并提交，成功时 stdout 只打印一行作业号。

| 选项 | 默认 | 说明 |
|---|---|---|
| `--dry-run` | `false` | 只把翻译后的脚本写到 stdout，不提交，不写 `.submit` / `.info.yaml` |
| `--scheduler=<名称>` | 空 | 强制指定后端 |
| `--output-script=<路径>` | `<脚本>.submit` | 自定义生成脚本路径 |
| `--output-info=<路径>` | `<脚本>.info.yaml` | 自定义元数据 sidecar 路径 |
| `--watch` | `false` | 提交成功后轮询状态直到终止态 |
| `--watch-interval=<时长>` | `5s` | `--watch` 的轮询间隔 |
| `--no-log` | `false` | 不写 `.submit.log`（成功与失败都不写） |

执行顺序：解析 → 选后端 → 渲染 → 写 `.submit` → 提交 → 写 `.info.yaml` → 打印作业号 →
写 `.submit.log` →（可选）进入 watch。因此**提交失败时 `.submit` 已经存在**，
`.submit.log` 也会记录失败原因；`.submit.log` 的路径恒为 `<脚本>.submit.log`，
不受 `--output-script` 影响。

```bash
$ orvix submit myjob.sh
12345678
```

## orvix status

```bash
orvix status [--format text|json] <info.yaml>
```

查询一次作业状态。输入是 `submit` 生成的 `.info.yaml`，orvix 从中读取后端名与作业号。

| 选项 | 默认 | 说明 |
|---|---|---|
| `--format=text\|json` | `text` | `text` 打印调度器原始状态；`json` 输出机器可读对象 |

`--format json` 的字段：

| 字段 | 说明 |
|---|---|
| `job_id` | 作业号 |
| `scheduler` | 后端名 |
| `state` | 归一化状态，见 [调度器映射](scheduler-mapping.md) |
| `raw_state` | 调度器原始状态字符串 |
| `terminal` | 是否为终止状态 |

```bash
$ orvix status --format json myjob.sh.info.yaml
{"job_id":"12345","scheduler":"slurm","state":"COMPLETED","raw_state":"COMPLETED","terminal":true}
```

退出码：查询成功为 0（**与作业本身成败无关**）；info 文件缺失、缺 `job_id`、后端未知、
`--format` 非法或调度器查询失败时为 1。

## orvix watch

```bash
orvix watch [flags] <info.yaml>
```

持续轮询，每次打印一行 `[RFC3339 时间戳] <归一化状态>`，直到作业到达终止态或状态查询失败。

| 选项 | 默认 | 说明 |
|---|---|---|
| `-i, --interval=<时长>` | `5s` | 轮询间隔 |

```bash
$ orvix watch myjob.sh.info.yaml
[2026-05-22T10:00:00Z] PENDING
[2026-05-22T10:00:05Z] RUNNING
[2026-05-22T10:02:00Z] COMPLETED
```

**退出码警示**：到达任何终止态——包括 `FAILED`、`CANCELLED`、`TIMEOUT`——都返回成功（0）。
只有状态查询本身失败时才打印 `[时间戳] ERROR: ...` 并非零退出。
另外 `UNKNOWN` 不是终止态，状态一直无法识别时 watch 不会自行结束。
`submit --watch` 的退出码语义与此相同。

因此**不要用 `orvix watch` / `orvix submit --watch` 的退出码判断作业成败**；
在脚本或 CI 中请在 watch 结束后再读 `orvix status --format json` 的 `state` 字段：

```bash
orvix watch myjob.sh.info.yaml
state=$(orvix status --format json myjob.sh.info.yaml | python3 -c 'import json,sys; print(json.load(sys.stdin)["state"])')
if [ "$state" != "COMPLETED" ]; then
    echo "job finished with state=$state" >&2
    exit 1
fi
```

注意 `local` 后端只能区分运行中与已结束，作业失败同样归一化为 `COMPLETED`，
不能用于判断脚本是否成功。

## orvix kill

```bash
orvix kill [flags] <info.yaml>
```

终止作业。对已结束的作业重复执行视为成功（幂等）。

| 选项 | 默认 | 说明 |
|---|---|---|
| `-s, --signal=<信号>` | 空 | 显式指定信号：名称（`TERM`/`KILL`/`SIGUSR1`，大小写不敏感、可省 `SIG` 前缀）或数字 1–64 |

默认行为按后端区分：

| 后端 | 默认终止方式 | `-s` 支持 |
|---|---|---|
| `slurm` | 分阶段优雅终止：`scancel --full --signal=TERM` → 等待宽限期 → 兜底 `scancel`；PENDING/CONFIGURING 作业直接 `scancel` | 单次 `scancel --full --signal=<数字>` |
| `donau` | `djob -T <作业号>` | **不支持**，指定即报错 |
| `local` | 向整个进程组发 SIGTERM；无进程组时回退单进程 | 支持 |

宽限期默认 30 秒，可用 `ORVIX_SLURM_KILL_GRACE` 调整，见 [环境变量与默认值](configuration.md)。

## orvix version

```bash
orvix version
```

打印版本。发布构建为 Git 标签（如 `v1.2.3`，由 GoReleaser 在链接期注入）；
`go install` 安装的二进制为模块版本；本地开发构建为模块伪版本或 VCS 短哈希
（12 位，工作区有改动时加 `+dirty`）；均无信息时为 `dev`。

## 相关页面

- [调度器映射](scheduler-mapping.md) — 原始状态与归一化状态的对照
- [文件格式](file-formats.md) — 各命令生成文件的完整格式
- [环境变量与默认值](configuration.md) — 环境变量与全部默认值汇总

核对：`cmd/*.go`、`main.go`、`internal/submit/`、`internal/watch/watch.go`。
