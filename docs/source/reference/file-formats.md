# 文件格式

本页定义 orvix 生成的附属文件的完整格式。默认都写在与原脚本**同一目录**，重复提交会覆盖同名文件。

| 文件 | 权限 | 说明 |
|---|---|---|
| `<脚本>.submit` | 0755 | 翻译后的可执行脚本，实际交给调度器或本地执行 |
| `<脚本>.info.yaml` | 0644 | 作业元数据 sidecar，`status` / `watch` / `kill` 的输入 |
| `<脚本>.submit.log` | 0644 | 提交日志，成功与失败都会写 |

## 各命令生成哪些文件

| 场景 | `.submit` | `.info.yaml` | `.submit.log` |
|---|---|---|---|
| `generate` 成功 | 是 | 否 | 否 |
| `generate` 失败 | 视失败阶段而定 | 否 | **是**（记录错误） |
| `submit` 成功 | 是 | 是 | 是（`--no-log` 时否） |
| `submit` 提交失败 | 是（提交前已写出） | 否 | 是（`--no-log` 时否） |
| `submit` 解析/读取失败 | 否 | 否 | 是（`--no-log` 时否） |
| `submit --dry-run` | 否（写 stdout） | 否 | 仅失败时写 |

`--output-script` / `--output-info` 可自定义前两个文件的路径；
`.submit.log` 的路径固定为 `<脚本>.submit.log`，不随 `--output-script` 改变。

## .submit 脚本

改写规则：

1. shebang 取第一个非空行中的 `#!` 行并提升到文件最前（容忍原文件的前导空行，
   ecFlow 生成的作业文件常有这种情况）；
2. 紧随其後插入后端指令块（`#SBATCH` / `#DSUB` 行），指令块与正文之间空一行；
3. **删除全文所有 `#ORVIX` 行**——包括解析阶段未生效的、位于首条可执行语句之后的指令行；
4. 去掉正文的前导空行，其余内容原样保留。

`local` 后端没有指令块，生成的脚本就是去掉 `#ORVIX` 行后的原脚本。

```bash
#!/bin/bash
#SBATCH --job-name=demo
#SBATCH --partition=normal

echo "Hello from HPC"
```

## .info.yaml

`status` / `watch` / `kill` 通过它确定后端与作业号；文件被删除或移动后，这三个命令即无法工作。

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
    - key: account
      value: donau_proj
      condition: scheduler=donau
```

| 字段 | 可省略 | 说明 |
|---|---|---|
| `version` | 是 | 提交时使用的 orvix 版本 |
| `scheduler` | 否 | 实际使用的后端 |
| `job_id` | 否 | 调度器返回的作业号（local 后端为进程 PID） |
| `submitted_at` | 否 | 提交时间，RFC3339（含纳秒与时区） |
| `script_source` | 否 | 原始脚本绝对路径 |
| `script_generated` | 否 | 翻译后脚本绝对路径 |
| `submit_dir` | 是 | 提交时的工作目录 |
| `hostname` / `user` | 是 | 提交主机与用户名 |
| `submit_command` | 是 | 实际执行的提交命令；local 后端为 `.submit` 脚本路径本身 |
| `directives` | 是 | **过滤后生效**的指令列表，用于追溯 |

`directives` 列表项：

| 字段 | 可省略 | 说明 |
|---|---|---|
| `key` | 否 | 指令键（连字符形式） |
| `value` | 是 | 指令值；裸键时省略 |
| `condition` | 是 | 条件表达式，如 `scheduler=donau`；无条件指令省略 |

未知键与条件不匹配的指令不会出现在该列表中。

## .submit.log

成功与失败格式相同，仅首行标签不同：

```
[2026-05-10T11:27:22+08:00] status: SUCCESS

[submit-cmd]
sbatch --parsable /abs/path/script.sh.submit

[output]
12345678
```

失败时首行为 `[<时间戳>] ERROR: orvix submit failed`，`[output]` 段记录错误信息；
尚未执行提交命令时 `[submit-cmd]` 为 `(none)`。日志标签统一为 `orvix submit failed`，
即使失败发生在 `orvix generate`。

写日志失败会被静默忽略，不会掩盖原始错误。

## 相关页面

- [命令参考](cli.md) — 各命令的选项与退出码
- [指令参考](directives.md) — `directives` 字段中指令键的含义

核对：`internal/jobinfo/jobinfo.go`、`internal/script/render.go`、`internal/submit/submit.go`。
