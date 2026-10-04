orvix 文档
==========

`orvix` 是一个用于向 HPC 集群提交脚本作业的命令行工具。

只需在脚本头部使用统一的 ``#ORVIX key=value`` 语法编写资源需求，orvix 会自动将其转换为目标调度器
（SLURM / 华为 Donau / 本地运行）的指令并提交作业，同时提供状态查询、监视与终止等完整的作业生命周期命令。

.. toctree::
   :maxdepth: 2
   :caption: 快速上手

   quickstart

.. toctree::
   :maxdepth: 2
   :caption: 使用指南

   guide/directives
   guide/commands
   guide/output-files
   guide/schedulers

.. toctree::
   :maxdepth: 2
   :caption: 参考

   reference/jobspec

.. toctree::
   :maxdepth: 2
   :caption: 开发

   develop/architecture
   develop/adding-directive
