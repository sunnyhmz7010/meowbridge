# meowbridge 部署与发布设计

## 背景

meowbridge 已具备后端服务、内嵌管理后台和 GitHub 仓库基础设施。下一阶段需要补齐可交付部署链路，让项目可以通过 Docker 运行，并通过 GitHub Actions 在版本标签上自动发布服务器端产物。

本阶段范围限定为部署与发布基础设施，不改变 Webhook 行为、管理 API、前端交互和数据库 schema。

## 目标

- 提供可本地构建的 Docker 镜像。
- 提供单服务 Docker Compose 部署示例。
- 在 `main` 日常提交上运行持续集成验证。
- 在 `v*` tag 推送时自动发布 Linux 服务器二进制产物。
- 在 `v*` tag 推送时自动发布 GHCR 多架构镜像。
- 保持 README 面向用户，避免写入内部协作流程。

## 非目标

- 不引入反向代理、HTTPS 自动签发或域名配置示例。
- 不发布 Windows 或 macOS 二进制。
- 不改变当前 SQLite 存储实现。
- 不引入新的运行时服务依赖。
- 不改变 Webhook 标准接口语义。

## 发布架构

发布链路由两个 GitHub Actions workflow 组成。

`ci.yml` 面向日常验证，在 `push` 到 `main` 和 `pull_request` 时触发。虽然当前开发约定是直接推送 `main`，保留 `pull_request` 触发可以覆盖外部贡献和 GitHub 默认协作入口。

`release.yml` 面向版本发布，只在 `v*` tag 推送时触发。发布流程按顺序执行：checkout、前端依赖安装、前端构建、Go 测试、Linux 二进制构建、checksums 生成、GHCR 镜像构建与推送、GitHub Release 创建。

Release 代表已经通过测试和构建的提交。任一步失败时，后续发布步骤不继续执行。

## 发布产物

版本 tag 触发后生成以下 GitHub Release 资产：

- `meowbridge-linux-amd64.tar.gz`
- `meowbridge-linux-arm64.tar.gz`
- `checksums.txt`

二进制产物只覆盖服务器常用平台：

- `linux/amd64`
- `linux/arm64`

GHCR 镜像发布到：

- `ghcr.io/sunnyhmz7010/meowbridge:<tag>`
- `ghcr.io/sunnyhmz7010/meowbridge:latest`

Release 标题使用纯 tag，例如 `v0.1.0`。Release notes 使用 `## 更新内容` 标题，并由 GitHub CLI 基于提交历史生成内容。

## Docker 镜像设计

`Dockerfile` 使用多阶段构建：

1. 前端构建阶段使用 Node 20，执行 `npm ci` 和 `npm run build`。
2. Go 构建阶段使用 Go 1.23，将内嵌前端资源编译进 `cmd/meowbridge` 单二进制。
3. 运行阶段使用轻量基础镜像，创建非 root 用户运行服务。

容器默认配置：

- `HTTP_ADDR=:8080`
- `DATABASE_PATH=/data/meowbridge.db`
- 暴露端口 `8080`
- `/data` 用于 SQLite 持久化

启动仍要求用户提供：

- `ADMIN_PASSWORD`
- `JWT_SECRET`
- `MEOW_API_BASE_URL`

`JWT_SECRET` 不在镜像或 compose 文件中写死真实值，只提供占位示例。

## Compose 部署设计

`compose.yaml` 提供单服务部署：

- 服务名：`meowbridge`
- 镜像：`ghcr.io/sunnyhmz7010/meowbridge:latest`
- 端口映射：`8080:8080`
- 命名 volume：挂载到 `/data`
- 环境变量：`ADMIN_PASSWORD`、`JWT_SECRET`、`MEOW_API_BASE_URL`、`LOG_RETENTION_DAYS`

Compose 不包含 Caddy、Nginx、HTTPS、数据库外置服务或监控服务。这样保持默认部署足够小，也避免公共示例暗示生产 TLS 方案已经被项目托管。

## GitHub Actions 权限

`ci.yml` 使用只读代码权限即可。

`release.yml` 使用最小必要权限：

- `contents: write`：创建 GitHub Release 和上传资产。
- `packages: write`：向 GHCR 推送容器镜像。

GHCR 发布使用同仓库 `GITHUB_TOKEN`，不要求用户配置 PAT。首次发布后，如果 GitHub Packages 默认可见性为 private，用户需要在 GitHub Packages 页面将镜像改为 public。

## 文件变更范围

预期新增：

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `Dockerfile`
- `.dockerignore`
- `compose.yaml`

预期修改：

- `README.md`：补充 Docker、Compose、Release 下载和镜像使用方式。
- `AGENTS.md`：如有必要，补充开发者发布约定，不写入面向用户的公共文档。

## 验证策略

本地验证命令：

```powershell
go test ./...
```

```powershell
cd web
npm run test
npm run type-check
npm run build
cd ..
```

```powershell
docker build -t meowbridge:test .
```

CI 验证：

- Go 测试通过。
- 前端测试通过。
- TypeScript 类型检查通过。
- 前端生产构建通过。

Release 验证：

- Linux `amd64` 和 `arm64` 二进制成功生成。
- `checksums.txt` 包含发布资产校验值。
- GHCR 镜像成功推送 tag 和 `latest`。
- GitHub Release 包含两个 Linux tarball 和 `checksums.txt`。

## 失败处理

- 任一测试失败时，不创建 Release，不推送 GHCR 镜像。
- Docker 构建失败时，不创建镜像发布结果。
- Release 创建失败时，workflow 失败并保留日志，便于从失败步骤定位。
- 不在 workflow 中吞掉错误或制造伪成功状态。

## 验收标准

- 用户可以通过 Docker Compose 启动服务，并在 `http://localhost:8080/admin/` 访问管理后台。
- 用户可以通过 GHCR 镜像运行服务，并将 `/data` 持久化到宿主机或 volume。
- 推送 `v*` tag 后，GitHub Release 自动出现 Linux 双架构二进制和 checksum。
- 推送 `v*` tag 后，GHCR 自动出现对应 tag 和 `latest` 镜像。
- README 中的运行示例与实际环境变量和端口一致。
