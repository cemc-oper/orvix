# 测试策略

orvix 的验证分四层：单元测试、契约 golden、CI 门禁、真实 HPC 集成测试。
改动越靠近用户可见行为，越需要靠后几层。

## 1. 单元测试

使用 `stretchr/testify`（`assert` / `require`），测试文件与被测代码同目录。

| 位置 | 覆盖内容 |
|---|---|
| `internal/directive/parser_test.go` | 引号值、裸键、条件过滤与覆盖、畸形输入、未知键丢弃 |
| `internal/scheduler/*_test.go` | 各后端的指令映射与特殊逻辑（Donau 时间转秒、`-R`/`-d` 合并、SLURM kill 分阶段、信号解析、状态归一化） |
| `internal/script/render_test.go` | shebang 提升、`#ORVIX` 剥离、空行处理 |
| `internal/submit/*_test.go` | 提交/生成管线、产物路径推导、日志写入 |
| `internal/jobinfo/jobinfo_test.go` | sidecar 读写 |
| `cmd/status_test.go`、`internal/version/version_test.go` | 命令输出与版本推导 |

```bash
make test
go test -race ./...     # local 后端与信号处理有并发逻辑，CI 也跑 -race
```

注意 `-race` 需要 CGO，与发布构建的 `CGO_ENABLED=0` 不冲突：测试在开发机/CI 上跑，
发布产物仍是静态二进制。

## 2. 契约 golden 测试

`spec/jobspec/conformance_test.go` 的 `TestConformanceGolden` 把
`conformance/vectors/*.sh` 逐个向量、逐个后端（`slurm` / `donau` / `local`）跑真实生成管线，
与 `conformance/golden/<后端>/` 逐字节比对。

**任何改变生成输出的改动都会让它变红**，这正是它的作用。有意改变输出时：

```bash
make regen-golden      # 即 go test ./spec/jobspec/ -run TestConformanceGolden -update
```

纪律：重新生成的 golden diff 必须**随同一个 commit 一起评审**，不要单独提交基线变更。

## 3. CI 门禁

`.github/workflows/ci.yml` 在每次 push 与 PR 上运行：

| Job | 内容 |
|---|---|
| `lint` | `gofmt -l .` 必须为空 + `go vet ./...` |
| `test` | `go test -race -coverprofile=coverage.out ./...`，覆盖率作为 artifact 上传 |
| `jobspec golden` | 单独跑 `TestConformanceGolden`，让契约漂移按名字可见 |
| `docs` | 指令表漂移校验（`make docs-check`）+ 严格文档构建（`make docs-html`，警告即失败） |
| `build` | linux/amd64 与 linux/arm64 交叉编译（CGO_ENABLED=0） |

发布流程见 [发布流程](releasing.md)。

## 4. 真实 HPC 集成测试

> **HPC-only**：以下需要在真实集群上执行；WSL 侧只能跑其中的离线部分。

工作区的 `runtime/orvix-devel/repo/orvix-hpc-test/` 是独立集成测试套件，真实调用
`sbatch` / `dsub` 提交并验证结果。orvix 的行为变更（尤其是 kill、状态、指令映射）
应在合并前用它实跑验证。

常用入口（详见该仓库 README）：

```bash
cd runtime/orvix-devel/repo/orvix-hpc-test

make check                                    # lint + 单元测试 + 构建（离线）
./bin/orvix-hpc-test --lint                   # 离线校验用例定义
make test-quick PLATFORM=hpc2023_sc1          # 快速冒烟（真实提交）
make test PLATFORM=hpc2023_sc1                # 全量
./bin/orvix-hpc-test --platform hpc2023_sc5 --tag smoke
```

与契约相关的两个机制：

- **conformance 联动**：套件 vendor 了 orvix 的向量与 golden，运行时自动合成 `conf_*`
  用例做 golden diff，在无调度器的 `ci_local` 平台也会真跑；
- **漂移检查**：`scripts/ci.sh drift` 克隆构建 orvix@main，比对 vendor 副本并跑 local 端到端，
  用于尽早发现主线与测试套件的漂移。

ecFlow 提交/kill 链路的端到端验证使用该仓库的 `examples/ecflow/` 示例套件，
runbook 见其 `docs/ecflow.md`；集成方式本身见 [与 ecFlow 集成](../guide/ecflow.md)。

## 改动该跑哪一层

| 改动类型 | 最小验证 |
|---|---|
| 仅文档 | `make docs-html`（含指令表漂移校验 + 严格构建），见 [文档编写指南](documentation.md) |
| 内部重构（无行为变化） | `make test` + `go vet` |
| 指令解析、脚本改写、后端映射 | 单元测试 + golden + HPC 实跑 |
| kill / 信号 / 状态 | 单元测试（含 `-race`）+ HPC 实跑（含 ecFlow kill 链路） |
| 新增指令键 / 新增后端 | 上述全部，并通知消费方，见 [新增一个指令](adding-directive.md)、[新增调度器后端](adding-scheduler.md) |
