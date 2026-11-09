# 指令参考

本页列出 orvix 识别的全部 19 个 `#ORVIX` 指令键，以及它们在三个后端的支持情况。
指令的权威定义是 `spec/jobspec/jobspec.schema.json`（见 [jobspec 契约](jobspec.md)），
本页是它的命令行投影。

## 语法速查

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name="my demo job"
#ORVIX [scheduler=donau] queue=donau_queue
#ORVIX exclusive

echo "running"
```

| 规则 | 说明 |
|---|---|
| 书写形式 | `#ORVIX key=value` 或裸键 `#ORVIX key`（布尔标志） |
| 标记 | 恰好大写 `#ORVIX`，`#` 后无空格；`ORVIX` 后须是空白或行尾，故 `#ORVIXFOO` 无效 |
| 行首空白 | 允许缩进 |
| 生效范围 | 从文件开头扫描，遇到第一个非空、非注释行即停止；其后的指令不生效 |
| 含空白的值 | 建议用双引号或单引号包裹（不加引号的 `key=a b` 也会拼接为 `a b`，但裸键后有多余词会报错） |
| 条件指令 | `[scheduler=<name>]` 前缀，只在该后端生效；可覆盖同键的无条件指令 |
| 覆盖 | 同键后出现者覆盖先出现者 |
| 未知键 | **静默丢弃**，不透传给调度器；加 `--debug` 可看到丢弃日志 |
| 其他 | 无转义处理、无环境变量展开；引号未闭合、`[` 未闭合、裸键后有多余词均为解析错误 |

后端的选择顺序：命令行 `--scheduler` > 脚本中第一条无条件 `scheduler=` > 默认 `local`。

## 指令全集

下面的分组表与支持矩阵由脚本从契约 schema 生成，与 `spec/jobspec/jobspec.schema.json`
保持一致；维护方式见 [文档编写指南](../develop/documentation.md)。

```{include} _generated/directive-tables.md
```

**不要留空值**：SLURM 后端会把 `#ORVIX project=` 翻译成空的 `#SBATCH --wckey=`，
可能导致提交被集群拒绝（Donau 后端相反，空值直接跳过该行）。不需要这两个字段时，
整行删掉或用条件指令限定后端。

支持矩阵中的"忽略"表示该键会被正常解析并记录在 `.info.yaml` 中，但不产生任何调度器指令。
生成行的完整格式、引号与合并规则见 [调度器映射](scheduler-mapping.md)。

## 相关页面

- [调度器映射](scheduler-mapping.md) — 各后端的指令输出格式与状态映射
- [文件格式](file-formats.md) — 解析后的指令如何记录在 `.info.yaml` 中
- [jobspec 契约](jobspec.md) — 指令集的权威定义与版本

核对：键集合、类型与默认值来自 `spec/jobspec/jobspec.schema.json`，说明与后端摘要来自
`docs/data/directives.json`；修改后运行 `make docs-directives` 重新生成，`make docs-check`
校验漂移。语法行为见 `internal/directive/parser.go`。
