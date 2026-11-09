orvix 文档
==========

`orvix` 是一个用于向 HPC 集群提交脚本作业的命令行工具。

只需在脚本头部使用统一的 ``#ORVIX key=value`` 语法编写资源需求，orvix 会自动将其转换为目标调度器
（SLURM / 华为 Donau / 本地运行）的指令并提交作业，同时提供状态查询、监视与终止等完整的作业生命周期命令。

**从哪里读起**：

- :doc:`开始 <start/index>` — 第一次接触 orvix：简介、安装，以及在无调度器的机器上跑通一个作业。
- :doc:`指南 <guide/index>` — 日常使用的操作方法，按任务组织。
- :doc:`开发 <develop/index>` — 修改 orvix 本身所需的架构说明与贡献者清单。
- :doc:`参考 <reference/index>` — 指令、命令、状态与文件格式的精确定义。

.. 文档站分为四区：开始 → 指南 → 开发 → 参考。
   顶层只列四个区的落地页，各区子页面在该区 index 的 toctree 中，
   避免侧边栏把全部页面平铺展开。
   所有 toctree 都带 :hidden: —— 它只负责导航，正文入口由上面的导览列表承担，
   否则 Sphinx 会在正文里再渲染一遍整棵目录树。
   页面归属与写作约定见 develop/documentation。

.. toctree::
   :hidden:
   :maxdepth: 2

   start/index
   guide/index
   develop/index
   reference/index
