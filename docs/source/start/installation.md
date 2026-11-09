# 安装

orvix 是单个静态二进制（发布构建关闭 CGO），不依赖 Python 或运行时环境，
可直接放在老版本 glibc 的 HPC 登录节点上运行。

## 下载预编译二进制（推荐）

从 [GitHub Releases](https://github.com/cemc-oper/orvix/releases) 下载对应平台的压缩包，
解压后将 `orvix` 放入 `PATH`：

```bash
# 以 Linux AMD64 为例
tar -xzf orvix_<version>_linux_amd64.tar.gz
install -m 755 orvix /path/to/bin/
```

发布页同时提供 `checksums.txt`（SHA-256），下载后可校验压缩包完整性：

```bash
sha256sum -c checksums.txt --ignore-missing
```

发布产物目前只有 `linux/amd64` 与 `linux/arm64` 两个目标。

## 使用 go install

```bash
go install github.com/cemc-oper/orvix@v<version>
```

需要有 Go 工具链与网络；二进制默认写入 `$GOBIN` 或 `$GOPATH/bin`。

## 从源码构建

```bash
git clone https://github.com/cemc-oper/orvix.git
cd orvix
make build              # 生成 bin/orvix（默认 CGO_ENABLED=0，全静态）
```

## 离线 HPC：交叉编译后上传

HPC 没有外网或没有 Go 工具链时，在有网络的机器上交叉编译，再上传二进制：

```bash
make build-linux-amd64  # 生成 bin/orvix-linux-amd64
make build-linux-arm64  # 生成 bin/orvix-linux-arm64
make build-all          # 当前平台 + 上述两个 Linux 目标
```

```bash
scp bin/orvix-linux-amd64 user@hpc:/path/to/bin/orvix
```

## 验证安装

```bash
$ orvix version
v1.2.3
```

版本字符串的来源：

| 构建方式 | 显示内容 |
|---|---|
| 正式发布（GoReleaser 注入） | Git 标签，如 `v1.2.3` |
| `go install ...@vX.Y.Z` | 模块版本 |
| 本地 `make build` / `go build` | 模块伪版本（如 `v0.2.1-0.2026...`）或 VCS 短哈希（12 位，工作区有改动时带 `+dirty`） |
| 无法获取任何版本信息 | `dev` |

## 运行期依赖

orvix 自身无依赖，但提交与查询要调用目标调度器的原生命令：

| 后端 | 需要的命令 |
|---|---|
| `slurm` | `sbatch`、`squeue`、`sacct`、`scancel` |
| `donau` | `dsub`、`djob` |
| `local` | 无 |

在 HPC 上通常先 `module load` 相应调度器环境，并确认这些命令对 orvix 的调用方
（例如 ecFlow server 进程）可见。

## 升级与降级

直接替换二进制即可。已有作业不受影响：`status` / `watch` / `kill` 只读取既有
`.info.yaml` 中的后端名与作业号。

需要留意的是**指令词汇随版本演进**：旧版本中的 `partition=` 已被 `queue=` 取代，
未知键会被静默丢弃。升级后若发现分区相关指令没有生效，先用
`orvix generate` 检查翻译结果，再对照 [指令参考](../reference/directives.md)。

## 下一步

[快速上手](quickstart.md) — 用 `local` 后端提交第一个作业。
