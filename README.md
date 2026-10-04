# orvix

![Maturity-Sandbox](https://img.shields.io/badge/Maturity-Sandbox-F9D71C)
[![ci](https://github.com/cemc-oper/orvix/actions/workflows/ci.yml/badge.svg)](https://github.com/cemc-oper/orvix/actions/workflows/ci.yml)

orvix 是一个用于向 HPC 集群提交脚本作业的命令行工具。

只需在脚本头部使用统一的 `#ORVIX key=value` 语法编写资源需求，orvix 会自动将其转换为目标调度器
（SLURM / 华为 Donau / 本地运行）的指令并提交作业。

完整文档见 **[orvix.readthedocs.io](https://orvix.readthedocs.io/)**。

## 安装

### 下载预编译二进制（推荐）

从 [GitHub Releases](https://github.com/cemc-oper/orvix/releases) 下载对应平台的压缩包，
解压后将 `orvix` 放入 `PATH` 即可。所有发布产物均为 CGO 关闭的全静态二进制，
可在老版本 glibc 的 HPC 上直接运行。

```bash
# 例如 Linux AMD64
tar -xzf orvix_<version>_linux_amd64.tar.gz
install -m 755 orvix /path/to/bin/
```

### 使用 go install

```bash
go install github.com/cemc-oper/orvix@v<version>
```

### 从源码构建

```bash
make build              # 生成 bin/orvix
```

在没有网络的 HPC 上，可在有网络的机器上交叉编译后上传（CGO 关闭，全静态二进制）：

```bash
make build-linux-amd64  # Linux AMD64
make build-linux-arm64  # Linux ARM64
scp bin/orvix-linux-amd64 user@hpc:/path/to/orvix
```

## 快速开始

在脚本顶部添加 `#ORVIX` 指令：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=demo
#ORVIX nodes=2
#ORVIX time=01:00:00

echo "Hello from HPC"
```

提交并监视作业：

```bash
$ orvix submit --watch myjob.sh
12345678
[2026-05-22T03:48:10Z] PENDING
[2026-05-22T03:48:15Z] RUNNING
[2026-05-22T03:50:20Z] COMPLETED
```

其他常用命令：

```bash
orvix generate myjob.sh          # 只生成翻译后的 .submit 脚本，不提交
orvix status myjob.sh.info.yaml  # 查询作业状态
orvix kill myjob.sh.info.yaml    # 终止作业
```

## 文档

在线文档：<https://orvix.readthedocs.io/>

本地构建文档：

```bash
pip install -r docs/requirements.txt
cd docs && make html           # 输出到 docs/build/html/
```

文档内容概览：

- **快速上手** — 安装与第一个作业
- **使用指南** — `#ORVIX` 指令参考、命令参考、生成的文件、调度器后端与指令映射
- **参考** — 与 takflow jobspec 契约的关系
- **开发** — 架构说明、新增指令清单

## 许可证

orvix 基于 [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) 授权。
