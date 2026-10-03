#!/usr/bin/env python3
# =============================================================================
# generate.py  ——  由 aliases.json 生成可直接 source 的 aliases.sh
#
# 这是「配置 -> 可移植产物」的生成器。Web 管理界面每次保存都会调用它,
# 因此别名.sh 始终保持与 aliases.json 一致, 且天然可被任何 bash 环境 source。
# =============================================================================
import json
import os
import sys

BASE = os.environ.get("APP_DIR", os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(BASE, "aliases.json")
OUT = os.path.join(BASE, "aliases.sh")

HEADER = """#!/usr/bin/env bash
# =============================================================================
# aliases.sh  ——  由 aliases.json 自动生成, 请勿手改(保存时会被覆盖)
# 编辑请走 Web 管理界面, 或改 aliases.json 后运行: python3 generate.py
# 本文件可直接 source, 与 Docker 镜像共用, 保证可移植。
# =============================================================================
"""


def main():
    with open(SRC, encoding="utf-8") as f:
        data = json.load(f)

    lines = [HEADER.rstrip("\n")]
    for s in data.get("settings", []):
        lines.append(s)
    lines.append("")
    lines.append("# ------------------------- 别名与函数(受 Web 管理) -------------------------")
    for e in data.get("entries", []):
        t, name, value = e["type"], e["name"], e.get("value", "")
        if t == "alias":
            lines.append(f"alias {name}='{value}'")
        elif t == "function":
            lines.append(f"{name}() {{")
            for bl in value.split("\n"):
                lines.append(bl)
            lines.append("}")

    out = "\n".join(lines) + "\n"
    with open(OUT, "w", encoding="utf-8") as f:
        f.write(out)
    print(f"已生成 {OUT}（共 {len(data.get('entries', []))} 条）")


if __name__ == "__main__":
    main()
