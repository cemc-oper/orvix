# 快速上手

本页介绍如何安装 orvix 并提交第一个作业。

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

在 Linux 上，可直接使用 Makefile 构建：

```bash
make build
```

构建成功后，将生成二进制文件 `bin/orvix`。

在没有网络的 HPC 上，可在有网络的机器上交叉编译 Linux 版本，再上传到目标 HPC 运行
（CGO 关闭，全静态二进制，兼容老版本 glibc）：

```bash
make build-linux-amd64    # Linux AMD64，生成 bin/orvix-linux-amd64
make build-linux-arm64    # Linux ARM64，生成 bin/orvix-linux-arm64
make build-all            # 当前平台 + 上述两个 Linux 目标
```

```bash
# 例如上传 Linux AMD64 版本
scp bin/orvix-linux-amd64 user@hpc:/path/to/orvix
```

验证安装：

```bash
$ orvix version
v1.2.3
```

## 第一个作业

编写一个脚本，在顶部用 `#ORVIX` 指令描述资源需求：

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=demo
#ORVIX nodes=2
#ORVIX time=01:00:00
#ORVIX exclusive

echo "Hello from HPC"
```

使用 `orvix generate` 仅生成翻译后的提交脚本（不提交）：

```bash
$ orvix generate myjob.sh
Generated: /abs/path/myjob.sh.submit
```

使用 `orvix submit` 提交脚本：

```bash
$ orvix submit myjob.sh
12345678
```

该命令会在脚本同目录生成两个附属文件：

- `myjob.sh.submit`：翻译后的脚本，实际提交给调度器执行
- `myjob.sh.info.yaml`：作业元数据，包含作业 ID，`status` / `watch` / `kill` 命令都以它为输入

使用 `orvix status` 查看作业状态：

```bash
$ orvix status myjob.sh.info.yaml
RUNNING
```

使用 `orvix watch` 持续监视作业直到结束：

```bash
$ orvix watch myjob.sh.info.yaml
[2026-05-22T03:48:10Z] PENDING
[2026-05-22T03:48:15Z] RUNNING
[2026-05-22T03:50:20Z] COMPLETED
```

使用 `orvix kill` 终止作业：

```bash
$ orvix kill myjob.sh.info.yaml
```

也可以一步完成提交和监视：

```bash
$ orvix submit --watch myjob.sh
12345678
[2026-05-22T03:48:10Z] PENDING
[2026-05-22T03:48:15Z] RUNNING
[2026-05-22T03:50:20Z] COMPLETED
```

## 工作流程

```{mermaid}
flowchart TD
    A[用户脚本<br/>#ORVIX 指令] --> B{orvix submit<br/>or generate?}
    B -->|generate| C[解析 #ORVIX 指令]
    C --> D[选择调度器后端]
    D --> E[生成翻译后的脚本]
    E --> G[写入 .submit 文件]
    G --> N[完成]
    B -->|submit| C
    E --> F{dry-run?}
    F -->|是| P[打印脚本]
    P --> N
    F -->|否| H[提交到调度器]
    H --> I[输出作业 ID]
    I --> J[生成 .info.yaml]
    J --> K{--watch?}
    K -->|是| L[轮询状态直到结束]
    K -->|否| N
    L --> N
```

## 下一步

- 指令的完整语法与全部可用键见 [指令参考](guide/directives.md)。
- 各命令的全部选项见 [命令参考](guide/commands.md)。
- 各调度器后端的指令映射与行为差异见 [调度器后端](guide/schedulers.md)。
