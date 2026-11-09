#!/usr/bin/env python3
"""Generate the directive reference tables from the jobspec schema.

The key set, types and defaults come from ``spec/jobspec/jobspec.schema.json``
(the authoritative contract). Chinese descriptions, examples and the per-backend
summary come from ``docs/data/directives.json``. The two must cover exactly the
same keys, so adding a contract key without documenting it fails loudly.

Usage:
    python3 docs/tools/generate_directive_tables.py           # write the tables
    python3 docs/tools/generate_directive_tables.py --check   # diff only, exit 1 on drift

Output: docs/source/reference/_generated/directive-tables.md, included by
reference/directives.md. Never edit the generated file by hand.

Standard library only, so it runs anywhere (including CI) without extra deps.
"""
from __future__ import annotations

import argparse
import difflib
import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SCHEMA_PATH = REPO_ROOT / "spec/jobspec/jobspec.schema.json"
DATA_PATH = REPO_ROOT / "docs/data/directives.json"
OUTPUT_PATH = REPO_ROOT / "docs/source/reference/_generated/directive-tables.md"

BACKENDS = ("slurm", "donau", "local")

HEADER = """\
<!-- 本文件由 docs/tools/generate_directive_tables.py 生成，请勿手工编辑。
     键集合、类型与默认值：改 spec/jobspec/jobspec.schema.json
     中文说明、示例与后端摘要：改 docs/data/directives.json
     重新生成：make docs-directives；CI 用 make docs-check 校验漂移 -->
"""

SCHEMA_TYPES = {
    "string": "字符串",
    "boolean": "布尔",
    "integer": "整数",
    "number": "数值",
}


def hyphenate(snake_key: str) -> str:
    return snake_key.replace("_", "-")


def type_and_default(prop: dict) -> str:
    """Render the schema type/default cell, e.g. '字符串，默认 `local`'."""
    raw_type = prop["type"]
    if isinstance(raw_type, list):
        # e.g. exclusive: ["string", "boolean"] — a bare flag or a valued key.
        types = set(raw_type)
        if types == {"string", "boolean"}:
            text = "裸键或带值"
        else:
            text = " / ".join(SCHEMA_TYPES.get(t, t) for t in sorted(types))
    else:
        text = SCHEMA_TYPES.get(raw_type, raw_type)
    if "default" in prop:
        default = prop["default"]
        if isinstance(default, bool):
            default = "true" if default else "false"
        text += f"，默认 `{default}`"
    return text


def load_inputs() -> tuple[dict, dict]:
    schema = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    data = json.loads(DATA_PATH.read_text(encoding="utf-8"))
    return schema, data


def validate(schema: dict, data: dict) -> list[str]:
    """Return a list of human-readable problems; empty means consistent."""
    problems: list[str] = []
    schema_keys = {hyphenate(k) for k in schema["properties"]}
    doc_keys = set(data["directives"])

    for key in sorted(schema_keys - doc_keys):
        problems.append(f"契约键 `{key}` 在 docs/data/directives.json 中缺少说明与示例")
    for key in sorted(doc_keys - schema_keys):
        problems.append(f"docs/data/directives.json 中的 `{key}` 不在契约 schema 中")

    grouped: list[str] = []
    for group in data["groups"]:
        grouped.extend(group["keys"])
    for key in sorted(set(grouped)):
        if grouped.count(key) > 1:
            problems.append(f"键 `{key}` 在 groups 中出现了 {grouped.count(key)} 次")
    for key in sorted(set(grouped) - schema_keys):
        problems.append(f"groups 中的 `{key}` 不在契约 schema 中")
    for key in sorted(schema_keys - set(grouped)):
        problems.append(f"契约键 `{key}` 未被任何分组收录")

    for key, entry in data["directives"].items():
        for field in ("description", "example", "backends"):
            if field not in entry:
                problems.append(f"`{key}` 缺少字段 `{field}`")
        for backend in BACKENDS:
            if backend not in entry.get("backends", {}):
                problems.append(f"`{key}` 的 backends 缺少 `{backend}` 摘要")
    return problems


def render(schema: dict, data: dict) -> str:
    lines = [HEADER.rstrip(), ""]
    for group in data["groups"]:
        lines += [f"## {group['title']}", "",
                  "| 指令 | 类型/默认 | 说明 | 示例 |",
                  "|---|---|---|---|"]
        for key in group["keys"]:
            prop = schema["properties"][key.replace("-", "_")]
            entry = data["directives"][key]
            lines.append(
                f"| `{key}` | {type_and_default(prop)} | "
                f"{entry['description']} | `{entry['example']}` |"
            )
        lines.append("")

    lines += ["## 后端支持矩阵", "",
              "| 指令 | SLURM | Donau | local |",
              "|---|---|---|---|"]
    for group in data["groups"]:
        for key in group["keys"]:
            backends = data["directives"][key]["backends"]
            cells = " | ".join(backends[b] for b in BACKENDS)
            lines.append(f"| `{key}` | {cells} |")
    lines.append("")
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true",
                        help="只校验漂移，不写文件；不一致时输出 diff 并以 1 退出")
    args = parser.parse_args()

    schema, data = load_inputs()
    problems = validate(schema, data)
    if problems:
        print("指令文档数据与契约不一致：", file=sys.stderr)
        for problem in problems:
            print(f"  - {problem}", file=sys.stderr)
        return 1

    content = render(schema, data)
    if args.check:
        if not OUTPUT_PATH.exists() or OUTPUT_PATH.read_text(encoding="utf-8") != content:
            print(f"{OUTPUT_PATH} 与 schema/数据文件不一致，请运行 make docs-directives",
                  file=sys.stderr)
            if OUTPUT_PATH.exists():
                diff = difflib.unified_diff(
                    OUTPUT_PATH.read_text(encoding="utf-8").splitlines(),
                    content.splitlines(), fromfile="committed", tofile="generated",
                    lineterm="",
                )
                print("\n".join(diff), file=sys.stderr)
            return 1
        print(f"OK: {OUTPUT_PATH} 与契约/数据文件一致")
        return 0

    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT_PATH.write_text(content, encoding="utf-8")
    print(f"Wrote {OUTPUT_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
