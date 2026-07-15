# meowbridge Deployment Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Docker, Docker Compose, CI, and tag-based GitHub Release/GHCR publishing for meowbridge.

**Architecture:** Keep deployment infrastructure outside application runtime code. CI validates existing Go and Vue test/build commands; release builds the embedded admin UI first, then compiles Linux binaries and Docker images from the same checked-out source.

**Tech Stack:** Go 1.23, Node.js 20, npm, Vue/Vite, Docker Buildx, GitHub Actions, GHCR, GitHub CLI.

## Global Constraints

- Do not change Webhook behavior, management API behavior, frontend interactions, or SQLite schema.
- Do not add runtime service dependencies.
- Release binaries target only `linux/amd64` and `linux/arm64`.
- Docker default database path is `/data/meowbridge.db`.
- Docker default HTTP address is `:8080`.
- Required runtime environment variables remain `ADMIN_PASSWORD`, `JWT_SECRET`, and `MEOW_API_BASE_URL`.
- GHCR image names are `ghcr.io/sunnyhmz7010/meowbridge:<tag>` and `ghcr.io/sunnyhmz7010/meowbridge:latest`.
- Release titles use the plain tag name, for example `v0.1.0`.
- Release notes use the `## 更新内容` heading.
- README stays user-facing and must not include internal AI workflow or handoff notes.
- Commit messages are in Chinese.

---

## File Structure

- Create `.github/workflows/ci.yml`: daily validation workflow for `main` and pull requests.
- Create `.github/workflows/release.yml`: tag-based release workflow for binaries, checksums, GHCR, and GitHub Release.
- Create `Dockerfile`: multi-stage frontend and Go build, non-root runtime, `/data` persistence default.
- Create `.dockerignore`: exclude local/generated/sensitive files from Docker context.
- Create `compose.yaml`: single-service deployment using the GHCR image and a named data volume.
- Modify `README.md`: add Docker/Compose and release download usage.
- Modify `AGENTS.md`: add developer-only release convention for `v*` tag publishing.

---

### Task 1: Add CI Workflow

**Files:**
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: existing Go tests via `go test ./...`; existing frontend scripts in `web/package.json`.
- Produces: a GitHub Actions workflow named `CI` that verifies pushes to `main` and pull requests.

- [ ] **Step 1: Create CI workflow**

Create `.github/workflows/ci.yml` with this content:

```yaml
name: CI

on:
  push:
    branches:
      - main
  pull_request:

permissions:
  contents: read

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.23.x'
          cache: true

      - name: Run Go tests
        run: go test ./...

      - name: Set up Node.js
        uses: actions/setup-node@v7
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: web/package-lock.json

      - name: Install frontend dependencies
        working-directory: web
        run: npm ci

      - name: Run frontend tests
        working-directory: web
        run: npm run test

      - name: Run frontend type check
        working-directory: web
        run: npm run type-check

      - name: Build frontend
        working-directory: web
        run: npm run build
```

- [ ] **Step 2: Validate workflow syntax locally**

Run:

```powershell
Get-Content -Raw .github\workflows\ci.yml
```

Expected: file contains `name: CI`, `go test ./...`, `npm run test`, `npm run type-check`, and `npm run build`.

- [ ] **Step 3: Run the same local verification commands**

Run:

```powershell
go test ./...
```

Expected: all Go packages pass.

Run:

```powershell
cd web
npm run test
npm run type-check
npm run build
cd ..
```

Expected: frontend tests, type check, and production build pass.

- [ ] **Step 4: Commit CI workflow**

Run:

```powershell
git add .github/workflows/ci.yml
git commit -m "添加持续集成工作流"
```

Expected: commit succeeds and contains only `.github/workflows/ci.yml`.

---

### Task 2: Add Docker Image Build Files

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

**Interfaces:**
- Consumes: `web/package-lock.json`, `web/package.json`, `web/src`, existing Go module files, and `cmd/meowbridge`.
- Produces: a runnable container image exposing port `8080`, running as a non-root user, with `DATABASE_PATH=/data/meowbridge.db`.

- [ ] **Step 1: Create Dockerfile**

Create `Dockerfile` with this content:

```dockerfile
# syntax=docker/dockerfile:1

FROM node:20-bookworm-slim AS web-build
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.23-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN rm -rf internal/webui/dist
COPY --from=web-build /src/web/dist ./internal/webui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/meowbridge ./cmd/meowbridge

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
ENV HTTP_ADDR=:8080
ENV DATABASE_PATH=/data/meowbridge.db
COPY --from=go-build /out/meowbridge /app/meowbridge
EXPOSE 8080
VOLUME ["/data"]
USER nonroot:nonroot
ENTRYPOINT ["/app/meowbridge"]
```

- [ ] **Step 2: Create Docker ignore file**

Create `.dockerignore` with this content:

```gitignore
.git
.github
docs/superpowers
web/node_modules
web/dist
node_modules
*.db
*.db-*
dist
tmp
coverage
.DS_Store
*.log
.env
.env.*
```

- [ ] **Step 3: Verify frontend build output path**

Run:

```powershell
cd web
npm run build
cd ..
```

Expected: command passes.

Run:

```powershell
Test-Path web\dist\index.html
```

Expected: `True`.

Run:

```powershell
Select-String -Path Dockerfile -Pattern "COPY --from=web-build /src/web/dist ./internal/webui/dist"
```

Expected: the Dockerfile copies the fresh Vite output into `internal/webui/dist` before compiling the Go binary.

- [ ] **Step 4: Build Docker image locally**

Run:

```powershell
docker build -t meowbridge:test .
```

Expected: image builds successfully.

- [ ] **Step 5: Commit Docker build files**

Run:

```powershell
git add Dockerfile .dockerignore
git commit -m "添加 Docker 镜像构建配置"
```

Expected: commit succeeds and contains only `Dockerfile` and `.dockerignore`.

---

### Task 3: Add Docker Compose Deployment

**Files:**
- Create: `compose.yaml`

**Interfaces:**
- Consumes: the Docker runtime contract from Task 2.
- Produces: a single-service Compose deployment with a persistent `/data` volume and example values for required environment variables.

- [ ] **Step 1: Create compose file**

Create `compose.yaml` with this content:

```yaml
services:
  meowbridge:
    image: ghcr.io/sunnyhmz7010/meowbridge:latest
    container_name: meowbridge
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      HTTP_ADDR: ":8080"
      DATABASE_PATH: /data/meowbridge.db
      ADMIN_PASSWORD: change-me
      JWT_SECRET: replace-with-long-random-secret
      MEOW_API_BASE_URL: https://api.chuckfang.com
      LOG_RETENTION_DAYS: "14"
    volumes:
      - meowbridge-data:/data

volumes:
  meowbridge-data:
```

- [ ] **Step 2: Validate compose config**

Run:

```powershell
docker compose config
```

Expected: compose renders one `meowbridge` service and one `meowbridge-data` volume without errors.

- [ ] **Step 3: Commit compose file**

Run:

```powershell
git add compose.yaml
git commit -m "添加 Docker Compose 部署示例"
```

Expected: commit succeeds and contains only `compose.yaml`.

---

### Task 4: Add Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: Dockerfile from Task 2 and existing project build commands.
- Produces: a tag-triggered workflow that publishes Linux binary archives, checksums, GHCR images, and a GitHub Release.

- [ ] **Step 1: Create release workflow**

Create `.github/workflows/release.yml` with this content:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

env:
  IMAGE_NAME: ghcr.io/${{ github.repository }}

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.23.x'
          cache: true

      - name: Set up Node.js
        uses: actions/setup-node@v7
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: web/package-lock.json

      - name: Install frontend dependencies
        working-directory: web
        run: npm ci

      - name: Run frontend tests
        working-directory: web
        run: npm run test

      - name: Run frontend type check
        working-directory: web
        run: npm run type-check

      - name: Build frontend
        working-directory: web
        run: npm run build

      - name: Copy frontend assets for Go embed
        run: |
          rm -rf internal/webui/dist
          mkdir -p internal/webui/dist
          cp -R web/dist/. internal/webui/dist/

      - name: Run Go tests
        run: go test ./...

      - name: Build Linux binaries
        shell: bash
        run: |
          set -euo pipefail
          mkdir -p dist
          for arch in amd64 arm64; do
            output="dist/meowbridge-linux-${arch}"
            GOOS=linux GOARCH="${arch}" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${output}/meowbridge" ./cmd/meowbridge
            tar -C "${output}" -czf "dist/meowbridge-linux-${arch}.tar.gz" meowbridge
          done

      - name: Generate checksums
        working-directory: dist
        run: sha256sum *.tar.gz > checksums.txt

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v4

      - name: Log in to GHCR
        uses: docker/login-action@v4
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract Docker metadata
        id: meta
        uses: docker/metadata-action@v6
        with:
          images: ${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=tag
            type=raw,value=latest
          labels: |
            org.opencontainers.image.source=https://github.com/${{ github.repository }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v7
        with:
          context: .
          push: true
          platforms: linux/amd64,linux/arm64
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}

      - name: Generate release notes
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          TAG_NAME: ${{ github.ref_name }}
        run: |
          printf "## 更新内容\n\n" > release-notes.md
          gh api "repos/${GITHUB_REPOSITORY}/releases/generate-notes" \
            -f tag_name="$TAG_NAME" \
            --jq .body >> release-notes.md

      - name: Create GitHub Release
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          TAG_NAME: ${{ github.ref_name }}
        run: |
          gh release create "$TAG_NAME" \
            dist/meowbridge-linux-amd64.tar.gz \
            dist/meowbridge-linux-arm64.tar.gz \
            dist/checksums.txt \
            --title "$TAG_NAME" \
            --notes-file release-notes.md
```

- [ ] **Step 2: Inspect release notes command**

Run:

```powershell
Get-Content -Raw .github\workflows\release.yml
```

Expected: the workflow writes `## 更新内容` to `release-notes.md`, calls `releases/generate-notes`, and creates the release with `--title "$TAG_NAME"` plus `--notes-file release-notes.md`.

- [ ] **Step 3: Validate YAML can be parsed**

Run:

```powershell
Get-Content .github\workflows\release.yml | Select-String -Pattern "docker/build-push-action@v7|packages: write|contents: write|platforms: linux/amd64,linux/arm64"
```

Expected: all four strings are found.

- [ ] **Step 4: Commit release workflow**

Run:

```powershell
git add .github/workflows/release.yml
git commit -m "添加自动发布工作流"
```

Expected: commit succeeds and contains only `.github/workflows/release.yml`.

---

### Task 5: Update User and Developer Documentation

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`

**Interfaces:**
- Consumes: Dockerfile, compose file, and workflow behavior from Tasks 1-4.
- Produces: user-facing Docker/Release instructions in README and developer-only release conventions in AGENTS.

- [ ] **Step 1: Update README quick start**

In `README.md`, after the existing local Go run example, add:

````markdown
### Docker Compose

```bash
docker compose up -d
```

默认 Compose 会监听 `8080`，并将 SQLite 数据持久化到 `meowbridge-data` volume。首次部署前请修改 `compose.yaml` 中的 `ADMIN_PASSWORD`、`JWT_SECRET` 和 `MEOW_API_BASE_URL`。

### Docker 镜像

```bash
docker run -d \
  --name meowbridge \
  -p 8080:8080 \
  -v meowbridge-data:/data \
  -e ADMIN_PASSWORD=change-me \
  -e JWT_SECRET=replace-with-long-random-secret \
  -e MEOW_API_BASE_URL=https://api.chuckfang.com \
  ghcr.io/sunnyhmz7010/meowbridge:latest
```
````

Expected: the README keeps these instructions user-facing and does not mention internal agent workflows.

- [ ] **Step 2: Add release download section**

In `README.md`, add a concise section near usage or installation:

```markdown
### Release 下载

版本发布页提供 Linux `amd64` 和 `arm64` 二进制包，以及 `checksums.txt` 校验文件。下载后解压并按环境变量说明启动即可。
```

- [ ] **Step 3: Update AGENTS release convention**

In `AGENTS.md`, add this under the release or development rules:

```markdown
- 发布稳定版本时推送 `v*` tag；Release workflow 会生成 Linux 二进制、checksum 和 GHCR 镜像。
```

- [ ] **Step 4: Validate documentation references**

Run:

```powershell
Select-String -Path README.md -Pattern "docker compose up -d","ghcr.io/sunnyhmz7010/meowbridge:latest","checksums.txt"
```

Expected: all three strings are found.

- [ ] **Step 5: Commit documentation updates**

Run:

```powershell
git add README.md AGENTS.md
git commit -m "补充部署发布文档"
```

Expected: commit succeeds and contains `README.md` plus `AGENTS.md`.

---

### Task 6: Final Verification and Push

**Files:**
- No new files expected.
- Verify all files touched by Tasks 1-5.

**Interfaces:**
- Consumes: all outputs from previous tasks.
- Produces: verified `main` pushed to `origin/main`.

- [ ] **Step 1: Check changed files and commit history**

Run:

```powershell
git status --short --branch
git log --oneline -8
```

Expected: branch is `main`, only intentional changes are present, and recent commits match the task commits.

- [ ] **Step 2: Run Go verification**

Run:

```powershell
go test ./...
```

Expected: all Go packages pass.

- [ ] **Step 3: Run frontend verification**

Run:

```powershell
cd web
npm run test
npm run type-check
npm run build
cd ..
```

Expected: frontend tests, type check, and production build pass.

- [ ] **Step 4: Run Docker verification**

Run:

```powershell
docker build -t meowbridge:test .
```

Expected: Docker image builds successfully.

Run:

```powershell
docker compose config
```

Expected: compose config renders successfully.

- [ ] **Step 5: Inspect final status**

Run:

```powershell
git status --short --branch
```

Expected: worktree is clean and branch is ahead of `origin/main` by the new commits.

- [ ] **Step 6: Push main**

Run:

```powershell
git push origin main
```

Expected: push succeeds and `origin/main` includes the deployment release infrastructure commits.

---

## Self-Review Notes

- Spec coverage: CI, release workflow, Dockerfile, `.dockerignore`, Compose, README, AGENTS release convention, local verification, GHCR permissions, release assets, and push are covered.
- Scope check: the plan only changes deployment/release infrastructure and documentation; it does not alter runtime behavior or database schema.
- Placeholder scan: the plan contains no deferred implementation choices or unfinished task language.
- Risk note: Dockerfile copies Vite output from `web/dist` into `internal/webui/dist` because Go embeds `internal/webui/dist`.
