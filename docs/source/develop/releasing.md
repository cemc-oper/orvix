# 发布流程

orvix 通过 Git 标签触发 GoReleaser 发布，产物是静态 Linux 二进制压缩包。
本页面向维护者。

## 版本号从哪里来

`orvix version` 的取值链路（`internal/version/version.go`）：

1. 发布构建由 GoReleaser 在链接期注入：
   `-X github.com/cemc-oper/orvix/internal/version.Version=<tag 去掉 v 后的版本>`；
2. 未注入时读 Go build info：`go install ...@vX.Y.Z` 安装的二进制取模块版本；
3. 本地开发构建退化为 VCS 短哈希（12 位，脏工作区加 `+dirty`）或模块伪版本；
4. 都没有则为 `dev`。

## 发布步骤

1. 确认主线就绪：`go test -race ./...`、`go vet ./...`、`gofmt -l .` 全部通过；
2. 确认文档与行为一致，特别是指令表、命令选项与状态映射（见
   [文档编写指南](documentation.md)）；
3. 若本次变更涉及指令集或后端，确认契约 `spec/jobspec/VERSION` 已按语义化版本更新，
   golden 基线已随变更评审；
4. 打标签并推送：

   ```bash
   git tag v1.3.0
   git push origin v1.3.0
   ```

5. `release.yml` 自动执行：`go vet` + `go test` → GoReleaser 构建并发布到 GitHub Releases。

## 发布产物

`.goreleaser.yaml` 定义：

| 产物 | 说明 |
|---|---|
| `orvix_<version>_linux_amd64.tar.gz` / `..._arm64.tar.gz` | 静态二进制压缩包（CGO_ENABLED=0，`-s -w` 精简符号） |
| `checksums.txt` | 全部产物的 SHA-256 校验和 |
| Release 说明 | 按 Git 提交生成的 changelog 与版本间对比链接 |

## 本地预演

不发布、只验证打包流程（需要本机安装 `goreleaser`）：

```bash
make release-snapshot     # goreleaser release --snapshot --clean，产物在 dist/
```

`dist/` 已在 `.gitignore` 中，不会误提交。

## 发布之后

- 通知契约消费方升 pin：takflow 与 orvix-hpc-test（仅在契约或行为有变化时必须），
  见 [jobspec 契约](../reference/jobspec.md)；
- 涉及提交/kill 行为变化时，在真实集群上用 orvix-hpc-test 复跑相关用例，
  见 [测试策略](testing.md)；
- 确认 Read the Docs 文档站已更新到对应版本。

## 相关页面

- [开发环境搭建](setup.md) — 构建与 make 目标
- [测试策略](testing.md)
- [新增一个指令](adding-directive.md) — 契约版本规则
