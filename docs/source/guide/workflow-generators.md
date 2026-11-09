# 被工作流生成器使用

orvix 是工作流栈里的**提交层**：takflow 及其上层应用（mcv-workflow、cma-tym 等）负责生成
作业脚本，orvix 负责提交。两者之间的接缝是 jobspec 契约。本页面向**写生成器的人**；
契约本身见 [jobspec 契约](../reference/jobspec.md)。

## 分工

```
生成器（takflow / 应用）          orvix                     调度器
资源模型 → #ORVIX 指令行   →   解析 + 翻译 + 提交    →   #SBATCH / #DSUB
                              写 .info.yaml sidecar
                              status / watch / kill
```

生成器不需要关心 SLURM 与 Donau 的旗标差异，只需按契约输出指令。

## 输出指令时遵守的规则

- **键名**：契约字段用 snake_case，指令用连字符（`job_name` → `job-name`）。
- **只输出有值的键**：未设置的键不要写成 `key=`——SLURM 会把空值翻成 `--wckey=`
  这样的空旗标，可能被集群拒绝。
- **位置**：指令必须在脚本开头注释区，第一条可执行语句之前。
- **引号**：值含空白时用双引号或单引号包裹。
- **布尔**：`exclusive` 为真时输出裸键 `#ORVIX exclusive`，为假时整行不输出；
  `requeue` 这类布尔键输出 `key=true|false`。
- **不做变量展开**：orvix 不会展开 `$VAR`，需要展开的值由生成器在渲染期完成。
- **用 `queue`，不用 `partition`**：后者已不在契约中，会被静默丢弃。

takflow 的 `takflow.jobspec.render` 已按上述规则实现（含固定的 canonical 输出顺序与
`conditional_for` 条件指令生成），直接用即可，不要另写一套拼接逻辑。

## 一个容易踩的命名坑

契约里的 `job_type` 是**调度器作业类型**（例如 Donau 的 `cosched`），
而 takflow 高层模型 `TaskResource.job_type` 是 `serial` / `parallel` 的**资源选择器**，
只用于推导队列等默认值，**不会也不应**被渲染成 `#ORVIX job-type`。
两者同名但语义不同，接入新模型时注意区分。

## 生成结果怎么验证

不需要集群就能验证生成器与 orvix 是否一致：

```bash
# 对生成的作业脚本做翻译检查
orvix generate --scheduler slurm --output-script /tmp/check.submit path/to/generated.sh
cat /tmp/check.submit
```

更系统的做法是用契约自带的 conformance 资产做回归：
`spec/jobspec/conformance/vectors/` 中的向量脚本经 `orvix generate` 处理后，
应与 `conformance/golden/<后端>/` 逐字节一致。takflow 与 orvix-hpc-test 都已把这道检查
纳入各自的测试，见 [jobspec 契约](../reference/jobspec.md)。

## 契约变更时做什么

契约只能在 orvix 仓库修改。流程：

1. orvix 侧按清单完成变更（schema、解析器、后端、golden、文档）；
2. 生成器仓库升 pin 同步只读副本，并检查是否需要把新键接入自己的配置模型；
3. 跑消费方的 golden / 一致性测试确认两端一致。

orvix 侧的完整清单见 [新增一个指令](../develop/adding-directive.md)。

## 相关页面

- [jobspec 契约](../reference/jobspec.md)
- [指令参考](../reference/directives.md)
- [与 ecFlow 集成](ecflow.md) — 生成器产出的流程如何提交
