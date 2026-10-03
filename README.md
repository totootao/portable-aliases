# 可移植别名 · Docker 部署 + Web 管理 (Go 实现)

把 Linux 别名做成「一套定义、到处可用」的环境，并配一个 **Web 管理面板** 来增删改。
核心是维护 **`aliases.json`** 作为权威数据源，Web 界面编辑它，并实时重新生成可直接 `source`
的 **`aliases.sh`**（保持可移植性不变）。后端用 **Go 标准库**（`net/http`）实现，零外部依赖。

## 目录结构

```
docker-aliases/
├── aliases.json    # 权威数据源(Web 界面 / 手动编辑这里)
├── aliases.sh      # 由 aliases.json 自动生成, 可直接 source(可移植产物)
├── main.go         # Go 实现: Web 管理后端 + 生成器(纯标准库, 无第三方依赖)
├── go.mod          # Go 模块定义
├── index.html      # Web 管理前端(原生 JS, 无需打包)
├── generate.py     # (可选)Python 版生成器, 作为无 Go 环境时的备用
├── Dockerfile      # 多阶段构建: golang 编译二进制 → 精简运行镜像
├── docker-compose.yml  # web(8080) + shell 两服务, 共享别名文件
├── build.sh
└── README.md
```

## 快速开始（Docker）

```bash
# 1. 构建镜像(多阶段: 编译 Go 二进制后塞入 ubuntu 运行镜像)
bash build.sh

# 2. 启动 Web 管理 + shell 两个服务
docker compose up -d

# 3. 浏览器打开管理面板
#    http://localhost:8080
#    - 表格查看全部别名/函数, 支持编辑、删除
#    - 右侧表单新增/更新(选 alias 或 function)
#    - 底部实时预览生成的 aliases.sh, 可一键复制
#    每次保存都会改写宿主机 ./aliases.json 并重新生成 ./aliases.sh

# 4. 进入带别名的交互式 shell(实时 source 同一个文件)
docker compose run --rm shell
```

## 本地（不装 Docker）

```bash
# 需要 Go 1.21+
go run main.go generate     # 由 json 生成 aliases.sh
source ./aliases.sh
ll

# 启动 Web 管理面板
go run main.go              # 然后访问 http://localhost:8080
# 或直接编译成二进制: go build -o app . && ./app
```

## 关键设计点

| 关注点 | 做法 | 原因 |
|--------|------|------|
| 单一数据源 | 只维护 `aliases.json` | Web 与本地共用同一份，杜绝漂移 |
| 可移植产物 | 生成 `aliases.sh` | 仍是纯 bash，可直接 source，与镜像共用 |
| Go 零依赖 | 仅用 `net/http` 标准库 | 无需 go mod 拉包，`go run` 即跑，镜像更小 |
| 复杂命令用函数 | `mkcd` / `extract` / `bigfiles` | 函数能进脚本、能带参数，比 alias 更稳 |
| 非交互式展开 | `shopt -s expand_aliases` + `BASH_ENV` | 否则 `docker run img bash -c 'll'` 会报 not found |
| 系统级加载 | `/etc/bash.bashrc` | 对所有用户、交互式 shell 生效 |
| Web 与 shell 同步 | compose 挂载同一 `aliases.sh` | 网页改完，新 shell 立即拿到 |

## 加新别名的正确姿势

- **Web 方式**：打开面板 → 右侧填类型/名称/定义 → 保存。推荐，零门槛。
- **手动方式**：编辑 `aliases.json` 的 `entries` → `go run main.go generate`（或等 Web 保存自动生成）。
- 团队共享：把 `aliases.json` 丢进 git / 内网，其他人 `docker compose pull` 即可。

## 安全提示

Web 管理界面**无鉴权**，仅限本机或可信内网访问。切勿直接暴露到公网。
如需公网，请在前方加反向代理 + 认证（如 Nginx basic auth / OAuth）。

## 常见坑

- **`docker run img ll` 报错 command not found**：`ll` 不是二进制，要 `docker run img bash -c 'll /'`。
- **alias 在脚本里不生效**：bash 默认不在非交互式展开 alias，改用函数（本方案已大量用函数）。
- **网页改了 shell 没变**：shell 容器在 `source` 时读取文件，重开一个 shell（或新容器）即生效。
