# 开发环境搭建

本页面向修改 orvix 代码的开发者。所有命令默认在 `repo/orvix/` 下执行。

## 环境要求

- Go **1.26**（以 `go.mod` 的 `go` 指令为准）
- `make`、`git`
- 文档构建另需 Python 环境，见 [文档编写指南](documentation.md)

环境准备：

- **HPC**：工作区提供 `runtime/orvix-devel/load_env.sh`，用 `module` 加载 Go 1.26 与 CMake，
  并把 `repo/orvix/bin` 加入 `PATH`：

  ```bash
  source runtime/orvix-devel/load_env.sh
  ```

- **WSL / 本地开发机**：自行安装 Go 1.26 即可，orvix 本体无 C 依赖；
  默认 `CGO_ENABLED=0`，产出全静态二进制。

跑测试与构建**不需要**调度器；只有真实提交才需要 `sbatch` / `dsub` 等命令。

## 构建

```bash
make build               # 生成 bin/orvix
make run ARGS="generate examples/hello.sh"
make run ARGS="submit --dry-run examples/hello.sh"
```

交叉编译（默认关闭 CGO，可在老 glibc 的 HPC 上运行）：

```bash
make build-linux-amd64
make build-linux-arm64
make build-all           # 当前平台 + 上述两个 Linux 目标
```

## 测试与静态检查

```bash
make test                                  # 全部单元测试
make test PKG=./internal/directive         # 单包
go test ./internal/directive -run TestParseBasic
go test -race ./...                        # 并发相关改动建议加 -race
make vet
gofmt -l .                                 # 应为空
make tidy                                  # 同步 go.mod / go.sum
```

`make regen-golden` 用于重新生成契约 golden 基线，使用纪律见 [测试策略](testing.md)。

## 调试

任何子命令加 `--debug` 都会把解析、后端选择、实际执行的调度器命令等写到 stderr：

```bash
go run . submit --debug --dry-run examples/hello.sh
```

## make 目标一览

| 目标 | 作用 |
|---|---|
| `all` | 同 `build`（默认目标） |
| `build` | 编译当前平台的 `bin/orvix` |
| `build-all` | 当前平台 + linux/amd64 + linux/arm64 |
| `build-linux-amd64` / `build-linux-arm64` | 交叉编译 Linux 交付物 |
| `test` | 运行单元测试，可用 `PKG=` 限定包 |
| `regen-golden` | 重新生成 jobspec conformance golden |
| `vet` | `go vet` |
| `tidy` | `go mod tidy` |
| `clean` | 删除 `bin/` |
| `run` | 构建后以 `ARGS` 运行 |
| `release-snapshot` | GoReleaser 本地预演，产物到 `dist/`，不发布 |
| `docs-directives` | 从契约 schema 重新生成指令参考表格 |
| `docs-check` | 校验指令表格是否存在漂移 |
| `docs-html` | 严格构建文档（警告即失败），先跑 `docs-check` |
| `help` | 打印目标与变量说明 |

可用变量：`GO_BUILD_FLAGS`（附加构建/测试参数）、`CGO_ENABLED`（默认 0）、`PKG`、`ARGS`、
`PYTHON`（文档相关目标使用的解释器，默认 `python3`）。

## 下一步

- [架构说明](architecture.md) — 改内部实现前先读
- [测试策略](testing.md)
- [文档编写指南](documentation.md) — 用户可见行为的文档同步要求
