orvix 文档
==========

`orvix` 是一个用于向 HPC 集群提交脚本作业的命令行工具。

只需在脚本头部使用统一的 ``#ORVIX key=value`` 语法编写资源需求，orvix 会自动将其转换为目标调度器
（SLURM / 华为 Donau / 本地运行）的指令并提交作业，同时提供状态查询、监视与终止等完整的作业生命周期命令。

**从哪里读起**：第一次接触 orvix 请看「开始」，先在无调度器的机器上跑通一个作业；
日常使用的操作方法在「指南」；修改 orvix 本身看「开发」；指令、命令、状态与文件格式的
精确定义在「参考」。

.. 文档站分为四区：开始 → 指南 → 开发 → 参考。
   页面归属与写作约定见 develop/documentation。

.. toctree::
   :maxdepth: 2
   :caption: 开始

   start/introduction
   start/installation
   start/quickstart

.. toctree::
   :maxdepth: 2
   :caption: 指南

   guide/directives
   guide/commands
   guide/output-files
   guide/schedulers

.. toctree::
   :maxdepth: 2
   :caption: 开发

   develop/architecture
   develop/adding-directive
   develop/documentation

.. toctree::
   :maxdepth: 2
   :caption: 参考

   reference/directives
   reference/cli
   reference/scheduler-mapping
   reference/file-formats
   reference/configuration
   reference/jobspec
