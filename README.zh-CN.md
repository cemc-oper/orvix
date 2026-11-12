# orvix

[English](README.md) | **简体中文**

![Maturity-Sandbox](https://img.shields.io/badge/Maturity-Sandbox-F9D71C)
[![GitHub Release](https://img.shields.io/github/v/release/cemc-oper/orvix)](https://github.com/cemc-oper/orvix/releases/latest)
[![ci](https://github.com/cemc-oper/orvix/actions/workflows/ci.yml/badge.svg)](https://github.com/cemc-oper/orvix/actions/workflows/ci.yml)
[![Documentation Status](https://app.readthedocs.org/projects/orvix/badge/?version=latest)](https://orvix.readthedocs.io/zh-cn/latest/?badge=latest)

orvix 是一个用于向 HPC 集群提交脚本作业的命令行工具。

只需在脚本头部使用统一的 `#ORVIX key=value` 语法编写资源需求，orvix 会自动将其转换为目标调度器
（SLURM / 华为 Donau / 本地运行）的指令并提交作业，同时提供状态查询、监视与终止等完整的作业生命周期命令。

完整文档见 **[orvix.readthedocs.io](https://orvix.readthedocs.io/)**。

## 安装

从 [GitHub Releases](https://github.com/cemc-oper/orvix/releases) 下载预编译二进制，
解压后放入 `PATH`（CGO 关闭的全静态二进制，可在老版本 glibc 的 HPC 上直接运行）：

```bash
tar -xzf orvix_<version>_linux_amd64.tar.gz
install -m 755 orvix /path/to/bin/
```

也可以使用 `go install github.com/cemc-oper/orvix@v<version>`，或从源码 `make build`。
离线 HPC 的交叉编译、校验与升级说明见在线文档的「开始 → 安装」。

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

没有集群也能试：把 `scheduler` 写成 `local`，即可在本机跑通完整流程。

## 文档

在线文档：<https://orvix.readthedocs.io/>

- **开始** — 简介与核心概念、安装、快速上手（用 `local` 后端，无需 HPC）
- **指南** — 编写作业脚本、选择后端、提交与脚本化、监视与终止、产物管理、
  ecFlow 集成、被工作流生成器使用、故障排查
- **开发** — 环境搭建、架构、测试策略、新增指令、新增后端、发布流程、文档编写
- **参考** — 指令全集、命令与退出码、调度器映射与状态、文件格式、环境变量、jobspec 契约

本地构建文档：

```bash
pip install -r docs/requirements.txt
cd docs && make html           # 输出到 docs/build/html/
```

提交文档改动前可在仓库根目录跑 `make docs-html`：严格构建（警告即失败），
并校验指令参考表是否与 jobspec 契约一致。

## 开发

```bash
make build       # 生成 bin/orvix
make test        # 单元测试
make vet         # go vet
```

贡献者请从文档的「开发」部分读起；修改用户可见行为时需在同一个 PR 更新对应文档页。

## 许可证

orvix 基于 [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) 授权。
