# 指南

日常使用 orvix 的操作指引，按任务组织。指令、命令、状态的精确表格在「参考」区。

```{toctree}
:maxdepth: 2

writing-scripts
choosing-scheduler
running-jobs
monitoring-and-killing
output-files
ecflow
workflow-generators
troubleshooting
```

- [编写作业脚本](writing-scripts.md) — 指令写法、常用配方与常见错误
- [选择调度器后端](choosing-scheduler.md) — 后端优先级、适用场景与 CMA 平台字段
- [提交与运行作业](running-jobs.md) — 四种用法、产物路径、脚本化与 CI 写法
- [监视与终止作业](monitoring-and-killing.md) — 状态查询、watch 边界、kill 与 trap 配合
- [生成的文件与管理](output-files.md) — 覆盖语义、并发实例、清理与归档
- [与 ecFlow 集成](ecflow.md) — `ECF_JOB_CMD` / `ECF_KILL_CMD` 与 kill 链路排查
- [被工作流生成器使用](workflow-generators.md) — 生成器输出指令的规则与验证方法
- [故障排查](troubleshooting.md) — 按"现象 → 原因 → 处置"组织的症状表
