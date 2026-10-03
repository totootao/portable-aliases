#!/usr/bin/env bash
set -e

# 构建可移植别名镜像(含 Web 管理)
docker build -t myaliases:latest .

echo "构建完成 ✅"
echo "启动服务:    docker compose up -d"
echo "管理面板:    http://localhost:8080"
echo "交互式 shell: docker compose run --rm shell"
echo "本地生成:    python3 generate.py && source ./aliases.sh"
