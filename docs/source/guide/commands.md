# 命令参考

orvix 提供六个子命令：`generate`、`submit`、`status`、`watch`、`kill`、`version`。
所有命令支持 `--debug` 输出调试日志（输出到 stderr）。

## orvix generate

```bash
orvix generate [flags] <脚本>
```

解析脚本中的 `#ORVIX` 指令，生成翻译后的脚本并写入 `.submit` 文件，
**不提交**到调度器，也不生成 `.info.yaml`。适用于需要预生成脚本、手动检查后再提交的场景。

```bash
$ orvix generate case/job/serial/orvix_serial.sh
Generated: /abs/path/orvix_serial.sh.submit
```

选项：

| 选项 | 说明 |
|---|---|
| `--scheduler=<名称>` | 强制指定调度器后端，覆盖脚本中的 `scheduler=` 指令 |
| `--output-script=<路径>` | 自定义生成脚本的路径（默认：脚本同目录 `<脚本>.submit`） |

## orvix submit

```bash
orvix submit [flags] <脚本>
```

解析脚本中的 `#ORVIX` 指令，生成翻译后的脚本并提交。成功时向 stdout 打印作业 ID，
并在脚本同目录生成 `.submit`、`.info.yaml` 和 `.submit.log` 文件（见 [生成的文件](output-files.md)）。

```bash
$ orvix submit case/job/serial/orvix_serial.sh
12345678
```

选项：

| 选项 | 说明 |
|---|---|
| `--dry-run` | 仅向 stdout 打印翻译后的脚本，不提交，也不生成任何附属文件 |
| `--scheduler=<名称>` | 强制指定调度器后端，覆盖脚本中的 `scheduler=` 指令 |
| `--output-script=<路径>` | 自定义生成脚本的路径 |
| `--output-info=<路径>` | 自定义 `.info.yaml` 的路径 |
| `--watch` | 提交成功后持续轮询状态，直到作业到达终止状态 |
| `--watch-interval=<时长>` | `--watch` 的轮询间隔（默认 `5s`） |
| `--no-log` | 不生成 `.submit.log`（默认成功/失败均会生成） |

## orvix status

```bash
orvix status <info.yaml>
```

查询一次作业状态，向 stdout 打印调度器返回的原始状态字符串：

```bash
$ orvix status case/job/serial/orvix_serial.info.yaml
RUNNING
```

## orvix watch

```bash
orvix watch [flags] <info.yaml>
```

持续轮询作业状态，每次轮询打印一行带时间戳的**归一化状态**，
直到作业到达终止状态（`COMPLETED`、`FAILED`、`CANCELLED`、`TIMEOUT`）或状态查询失败。

```bash
$ orvix watch case/job/serial/orvix_serial.info.yaml
[2026-05-22T10:00:00Z] PENDING
[2026-05-22T10:00:05Z] RUNNING
[2026-05-22T10:02:00Z] COMPLETED
```

选项：

| 选项 | 说明 |
|---|---|
| `-i, --interval=<时长>` | 轮询间隔（默认 `5s`） |

各后端的原始状态到归一化状态的映射见 [调度器后端](schedulers.md)。

## orvix kill

```bash
orvix kill [flags] <info.yaml>
```

终止作业。对已结束的作业重复执行 kill 视为成功（幂等）。

```bash
$ orvix kill case/job/serial/orvix_serial.info.yaml
```

选项：

| 选项 | 说明 |
|---|---|
| `-s, --signal=<信号>` | 显式指定发送的信号（名称如 `TERM`/`KILL` 或数字如 `15`） |

各后端的默认终止行为不同：

- **SLURM**：分阶段优雅终止——先 `scancel --full --signal=TERM`（TERM 同时送达 batch 脚本
  及其全部子进程，让脚本的信号 trap 有机会执行），轮询等待宽限期
  （默认 30 秒，可用环境变量 `ORVIX_SLURM_KILL_GRACE` 覆盖，单位秒，`0` 表示立即升级），
  超时仍未结束则兜底 `scancel`（controller cancel 路径，作业标记为 CANCELLED）。
- **Donau**：使用调度器默认的 `djob -T`；不支持自定义信号。
- **local**：向作业的整个进程组发送 SIGTERM（对进程组建立之前提交的旧作业回退到单进程）。

详见 [调度器后端](schedulers.md)。

## orvix version

```bash
orvix version
```

打印 orvix 版本。发布构建为 Git 标签（如 `v1.2.3`）；`go install` 安装的二进制为模块版本；
本地开发构建为 Git 提交哈希；均无则为 `dev`。
