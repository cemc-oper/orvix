# 生成的文件

orvix 在脚本**同一目录**下生成若干附属文件。重新提交会覆盖同名现有文件。

## orvix submit 生成的文件

```
path/to/script.sh              # 原始脚本（不变）
path/to/script.sh.submit       # 实际执行的翻译后脚本
path/to/script.sh.info.yaml    # 作业元数据（status / watch / kill 的输入）
path/to/script.sh.submit.log   # 提交日志
```

- 提交成功时，生成 `.submit`、`.info.yaml` 和 `.submit.log`。
- 提交失败时（如解析错误、调度器拒绝），仍生成 `.submit.log`，记录错误和时间戳。
- 使用 `--no-log` 可关闭 `.submit.log` 的生成（成功和失败均不生成）。
- `--output-script` / `--output-info` 可自定义前两个文件的路径。

## orvix generate 生成的文件

`orvix generate` 仅生成 `.submit` 文件，**不**生成 `.info.yaml` 和 `.submit.log`：

```
path/to/script.sh              # 原始脚本（不变）
path/to/script.sh.submit       # 翻译后的脚本
```

## .submit 脚本

翻译后的可执行脚本（权限 0755）：

- 保留原脚本的 shebang；
- 紧随其後插入调度器指令块（如 `#SBATCH` / `#DSUB` 行）；
- 删除所有 `#ORVIX` 行；
- 其余内容原样保留。

## .info.yaml

作业元数据 sidecar，`status` / `watch` / `kill` 命令通过它找到作业：

```yaml
version: v1.2.3
scheduler: slurm
job_id: "12345678"
submitted_at: 2026-05-10T11:27:22.107722252Z
script_source: /abs/path/script.sh
script_generated: /abs/path/script.sh.submit
submit_dir: /abs/path
hostname: login01
user: wangdp
submit_command: sbatch --parsable /abs/path/script.sh.submit
directives:
    - key: scheduler
      value: slurm
    - key: queue
      value: normal
    - key: nodes
      value: "2"
```

字段说明：

| 字段 | 说明 |
|---|---|
| `version` | 提交时使用的 orvix 版本 |
| `scheduler` | 实际使用的调度器后端 |
| `job_id` | 调度器返回的作业 ID（local 后端为进程 PID） |
| `submitted_at` | 提交时间（RFC3339） |
| `script_source` | 原始脚本绝对路径 |
| `script_generated` | 翻译后脚本绝对路径 |
| `submit_dir` | 提交时的工作目录 |
| `hostname` / `user` | 提交主机与用户名 |
| `submit_command` | 实际执行的提交命令 |
| `directives` | 解析后的全部 `#ORVIX` 指令（含条件指令的 `condition` 字段），用于追溯 |

## .submit.log

提交日志，成功与失败格式相同，仅状态标签不同：

```
[2026-05-10T11:27:22+08:00] status: SUCCESS

[submit-cmd]
sbatch --parsable /abs/path/script.sh.submit

[output]
12345678
```

失败时标签为 `ERROR: orvix submit failed`，`[output]` 段记录错误信息。
