# orvix

**English** | [简体中文](README.zh-CN.md)

![Maturity-Sandbox](https://img.shields.io/badge/Maturity-Sandbox-F9D71C)
[![GitHub Release](https://img.shields.io/github/v/release/cemc-oper/orvix)](https://github.com/cemc-oper/orvix/releases/latest)
[![ci](https://github.com/cemc-oper/orvix/actions/workflows/ci.yml/badge.svg)](https://github.com/cemc-oper/orvix/actions/workflows/ci.yml)
[![Documentation Status](https://app.readthedocs.org/projects/orvix/badge/?version=latest)](https://orvix.readthedocs.io/zh-cn/latest/?badge=latest)

orvix is a command-line tool for submitting script-based jobs to HPC clusters.

Define resource requirements with unified `#ORVIX key=value` directives at the top of a script.
orvix translates them into directives for the target scheduler (Slurm, Huawei Donau, or local
execution) and submits the job. It also provides commands for status queries, monitoring, and
termination, covering the complete job lifecycle.

Full documentation is available at **[orvix.readthedocs.io](https://orvix.readthedocs.io/)**.

## Installation

Download a prebuilt binary from [GitHub Releases](https://github.com/cemc-oper/orvix/releases),
extract it, and place it on your `PATH`. The binaries are fully static with CGO disabled, so they
can run directly on HPC systems with older glibc versions:

```bash
tar -xzf orvix_<version>_linux_amd64.tar.gz
install -m 755 orvix /path/to/bin/
```

You can also use `go install github.com/cemc-oper/orvix@v<version>`, or run `make build` to build
from source. See **Start → Installation** in the online documentation for offline cross-compilation,
checksum verification, and upgrade instructions.

## Quick Start

Add `#ORVIX` directives to the top of your script:

```bash
#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX queue=normal
#ORVIX job-name=demo
#ORVIX nodes=2
#ORVIX time=01:00:00

echo "Hello from HPC"
```

Submit and monitor the job:

```bash
$ orvix submit --watch myjob.sh
12345678
[2026-05-22T03:48:10Z] PENDING
[2026-05-22T03:48:15Z] RUNNING
[2026-05-22T03:50:20Z] COMPLETED
```

Other common commands:

```bash
orvix generate myjob.sh          # Generate the translated .submit script without submitting
orvix status myjob.sh.info.yaml  # Query the job status
orvix kill myjob.sh.info.yaml    # Terminate the job
```

No cluster is required to try orvix: set `scheduler` to `local` to run the complete workflow on
your local machine.

## Documentation

Online documentation: <https://orvix.readthedocs.io/>

- **Start** — Introduction, core concepts, installation, and a quickstart using the `local` backend
  without an HPC cluster.
- **Guide** — Writing job scripts, choosing a backend, submitting and scripting jobs, monitoring
  and termination, output-file management, ecFlow integration, workflow-generator usage, and
  troubleshooting.
- **Develop** — Environment setup, architecture, testing, adding directives and scheduler backends,
  releases, and documentation conventions.
- **Reference** — The complete directive set, commands and exit codes, scheduler mappings and
  states, file formats, environment variables, and the jobspec contract.

Build the documentation locally:

```bash
pip install -r docs/requirements.txt
cd docs && make html           # Output: docs/build/html/
```

Before submitting documentation changes, run `make docs-html` from the repository root. This runs
a strict build with warnings treated as errors and verifies that the directive reference tables
match the jobspec contract.

## Development

```bash
make build       # Build bin/orvix
make test        # Run unit tests
make vet         # Run go vet
```

Contributors should start with the **Develop** section of the documentation. When changing
user-visible behavior, update the corresponding documentation pages in the same pull request.

## License

orvix is licensed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0).
