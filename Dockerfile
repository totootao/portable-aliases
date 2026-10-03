# 可移植别名镜像 (+ Web 管理, Go 实现)
#
# 多阶段构建: 用 golang 镜像把 main.go 编译成静态二进制, 再塞进精简运行镜像。
# 一份镜像两种用法:
#   - 默认 CMD=bash           → 交互式带全套别名的 shell
#   - command /app/app        → 启动 Go 写的 Web 管理界面(配合 compose 的 web 服务)
# 可移植性不变: aliases.sh 由 aliases.json 生成, 仍可直接 source、与任意 bash 环境共用。
FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod main.go ./
RUN go build -o /out/app .

FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      bash curl iproute2 net-tools procps coreutils \
 && rm -rf /var/lib/apt/lists/*

# Web 管理二进制与前端的静态资源
COPY --from=build /out/app /app/app
COPY aliases.json /app/aliases.json
COPY index.html /app/index.html

# 别名文件(初始版, compose 会用宿主机文件覆盖挂载)
COPY aliases.sh /etc/aliases.sh

# 让所有用户的【交互式】bash 加载别名
RUN echo '[ -f /etc/aliases.sh ] && source /etc/aliases.sh' >> /etc/bash.bashrc

# 让【非交互式】bash 也加载(配合 aliases.sh 里的 shopt -s expand_aliases)
ENV BASH_ENV=/etc/aliases.sh
ENV ENV=/etc/aliases.sh
ENV APP_DIR=/app

# 默认进入交互式 bash(web 服务由 compose 用 command 覆盖)
CMD ["bash"]
