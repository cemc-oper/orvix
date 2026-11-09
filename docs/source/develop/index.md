# 开发

修改 orvix 本身所需的材料。改内部实现前先读架构说明，改用户可见行为时请同步更新文档。

```{toctree}
:hidden:
:maxdepth: 2

setup
architecture
testing
adding-directive
adding-scheduler
releasing
documentation
```

- [开发环境搭建](setup.md) — 工具链、构建、测试与 make 目标
- [架构说明](architecture.md) — 提交管线、Scheduler 接口与 Generator 模式
- [测试策略](testing.md) — 单元测试、契约 golden、CI 门禁与真实 HPC 实跑
- [新增一个指令](adding-directive.md) — 契约到后端再到文档的完整清单
- [新增调度器后端](adding-scheduler.md) — 接口实现、注册、契约与测试要求
- [发布流程](releasing.md) — 版本注入、标签发布与发布后事项
- [文档编写指南](documentation.md) — 四区归属、单一事实来源与生成的指令表格
