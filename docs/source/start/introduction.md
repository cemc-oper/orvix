# orvix 简介

orvix 是一个向 HPC 集群提交脚本作业的命令行工具：你在脚本头部用统一的
`#ORVIX key=value` 语法描述资源需求，orvix 把它翻译成目标调度器的指令并提交，
同时提供状态查询、持续监视与终止等完整的作业生命周期命令。

支持的提交目标：**SLURM**、**华为 Donau**、**本地直接运行**（`local`，无需调度器）。

## 为什么需要 orvix

- **一套脚本，多种集群**：资源需求只写一次，切换调度器不需要改写脚本。
- **把工作流系统与调度器解耦**：ecFlow / takflow 只负责任务编排，作业提交细节交给 orvix；
  上层生成器（如 takflow 应用）通过 jobspec 契约产出指令，见 [jobspec 契约](../reference/jobspec.md)。
- **提交之后仍然可操作**：orvix 把作业号与后端写进 sidecar 文件，后续查状态、监视、
  终止都不需要手工记录作业号，也不需要直接调用 `squeue` / `scancel`。

## 工作原理

```{mermaid}
flowchart TD
    A[作业脚本<br/>头部含 #ORVIX 指令] --> B[解析指令<br/>过滤未知键与条件指令]
    B --> C[选择后端<br/>slurm / donau / local]
    C --> D[生成 .submit<br/>插入调度器指令 删除 #ORVIX 行]
    D --> E{generate 还是 submit?}
    E -->|generate| F[只写 .submit<br/>结束]
    E -->|submit| G[提交到调度器<br/>或本地启动子进程]
    G --> H[打印作业号<br/>写 .info.yaml 与 .submit.log]
    H --> I{--watch?}
    I -->|是| J[轮询状态直到终止态]
    I -->|否| K[结束]
    J --> K
    H -.-> L[status / watch / kill<br/>读取 .info.yaml 定位作业]
```

`orvix generate` 只走前半段，便于在**没有调度器的机器**上检查翻译结果。

## 核心概念

| 概念 | 含义 |
|---|---|
| `#ORVIX` 指令 | 写在脚本头部注释块中的资源与调度选项，共 19 个键，见 [指令参考](../reference/directives.md) |
| 后端（scheduler） | 提交目标：`slurm` / `donau` / `local`；由 `scheduler=` 指令或 `--scheduler` 选项决定 |
| `.submit` 脚本 | 翻译后的可执行脚本：shebang 提前、插入 `#SBATCH` / `#DSUB` 行、删除 `#ORVIX` 行 |
| `.info.yaml` sidecar | 提交时写出的作业元数据（后端、作业号、路径、生效指令），是 `status` / `watch` / `kill` 的唯一输入 |
| 归一化状态 | 把各调度器的原始状态统一为 PENDING / RUNNING / COMPLETED / FAILED / CANCELLED / TIMEOUT / UNKNOWN，见 [调度器映射](../reference/scheduler-mapping.md) |

## 适合与不适合

**适合**：

- 脚本化的批处理作业提交，尤其是需要在多个集群/调度器之间复用同一套脚本；
- 作为 ecFlow / takflow 工作流的作业提交层（`ECF_JOB_CMD` / `ECF_KILL_CMD`）；
- 在开发机上用 `local` 后端调试作业脚本与流程，再原样拿到集群上跑。

**不适合**：

- 交互式会话、容器编排或长期服务；
- 任务之间的依赖编排——那是 ecFlow / takflow 的职责，orvix 只处理单个作业；
- 需要精确的本地作业成败判定：`local` 后端只有"运行中 / 已结束"两种状态，
  拿不到退出码，脚本失败也会显示 `COMPLETED`。

## 下一步

- [安装](installation.md)
- [快速上手](quickstart.md) — 用 `local` 后端跑通完整流程，无需 HPC
