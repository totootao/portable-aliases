#!/usr/bin/env python3
# =============================================================================
# app.py  ——  别名 Web 管理界面 (Flask)
#
# 数据模型: aliases.json 是权威来源; 每次增删改都回写 json 并重新生成 aliases.sh。
# 注意: 本服务【无鉴权】, 仅限本机或可信内网访问, 切勿直接暴露公网。
# =============================================================================
import json
import os
import subprocess

from flask import Flask, request, jsonify, send_from_directory, Response

BASE = os.environ.get("APP_DIR", os.path.dirname(os.path.abspath(__file__)))
JSON_PATH = os.path.join(BASE, "aliases.json")
SH_PATH = os.path.join(BASE, "aliases.sh")

app = Flask(__name__)


def load():
    with open(JSON_PATH, encoding="utf-8") as f:
        return json.load(f)


def save(data):
    with open(JSON_PATH, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    # 保存后立即重新生成 aliases.sh, 保持可移植文件与配置一致
    subprocess.run(["python3", os.path.join(BASE, "generate.py")], check=True)


@app.route("/")
def index():
    return send_from_directory(BASE, "index.html")


@app.route("/api/aliases", methods=["GET"])
def get_aliases():
    return jsonify(load())


@app.route("/api/aliases", methods=["POST"])
def add_alias():
    e = request.get_json(force=True)
    if not e.get("name") or not e.get("type"):
        return jsonify({"ok": False, "error": "name 与 type 必填"}), 400
    data = load()
    data.setdefault("entries", [])
    # 同名则更新, 否则追加
    for i, x in enumerate(data["entries"]):
        if x["name"] == e["name"]:
            data["entries"][i] = e
            break
    else:
        data["entries"].append(e)
    save(data)
    return jsonify({"ok": True})


@app.route("/api/aliases/<name>", methods=["DELETE"])
def del_alias(name):
    data = load()
    before = len(data.get("entries", []))
    data["entries"] = [x for x in data.get("entries", []) if x["name"] != name]
    if len(data["entries"]) == before:
        return jsonify({"ok": False, "error": "未找到 " + name}), 404
    save(data)
    return jsonify({"ok": True})


@app.route("/api/aliases.sh", methods=["GET"])
def view_sh():
    try:
        with open(SH_PATH, encoding="utf-8") as f:
            return Response(f.read(), mimetype="text/plain")
    except FileNotFoundError:
        return Response("# aliases.sh 尚未生成, 请先保存一次", mimetype="text/plain")


if __name__ == "__main__":
    # 首次启动确保产物存在
    if not os.path.exists(SH_PATH):
        subprocess.run(["python3", os.path.join(BASE, "generate.py")], check=False)
    app.run(host="0.0.0.0", port=8080)
