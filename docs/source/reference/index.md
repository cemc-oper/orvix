# 参考

面向查找的精确定义：指令键、命令选项、调度器映射、状态与文件格式。

```{toctree}
:maxdepth: 2

directives
cli
scheduler-mapping
file-formats
configuration
jobspec
```

- [指令参考](directives.md) — 19 个指令键的语义、默认值与后端支持矩阵
- [命令参考](cli.md) — 六个子命令的选项、输出与退出码
- [调度器映射](scheduler-mapping.md) — 指令输出格式与原始/归一化状态对照
- [文件格式](file-formats.md) — `.submit` / `.info.yaml` / `.submit.log` 的完整定义
- [环境变量与默认值](configuration.md) — `ORVIX_SLURM_KILL_GRACE` 与全部默认值
- [jobspec 契约](jobspec.md) — 契约的所有权、版本与消费方 vendor 关系
