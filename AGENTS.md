# meowbridge — 需求文档

## 一、背景

Meow 是优秀的手机系统级推送服务，仅支持一个昵称（nickname）作为唯一标识。

**痛点**：
- **消息混淆**：多场景共用一个昵称，无法区分
- **安全隔离**：昵称是公开链接一部分，一旦泄露牵连全部服务
- **协议不通用**：大量服务只支持推送 Telegram，无法直接对接 Meow

**方案**：meowbridge 作为中转层
1. 多入口 token → 映射到 Meow 昵称
2. Webhook 格式自动转换
3. **二期**：Telegram 消息劫持推送到 Meow

---

## 二、技术栈

| 组件 | 选型 | 说明 |
|------|------|------|
| 后端语言 | **Go 1.23+** | 单二进制、低内存、并发模型适合代理 |
| HTTP 路由 | **chi** | 极轻量 |
| 数据库 | **modernc.org/sqlite** | 纯 Go、零 CGO 依赖 |
| 鉴权 | **JWT** | 管理后台登录令牌 |
| 前端 | **Vue 3 + Vite + Tailwind CSS** | 嵌入二进制（go:embed） |

---

## 三、阶段划分

### 一期（核心功能）
- ✅ 单用户密码认证
- ✅ 推送接口管理（创建 / 编辑 / 删除 / 启停）
- ✅ Webhook 推送接收 `POST /webhook/{token}`
- ✅ JSONPath 字段映射 + 预设模板
- ✅ 推送日志

### 二期（Telegram 劫持）
- ⏳ 劫持 Telegram Bot API 写操作，提取消息推送到 Meow
- ⏳ 伪造 TG 成功响应返回

---

## 四、核心需求

### 4.1 推送接口

用户在管理后台创建推送接口，配置如下参数：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | 文本 | ✅ | 接口名称 |
| meow_nickname | 文本 | ✅ | 底层绑定的 Meow 昵称 |
| title | 文本 | ❌ | 默认标题（可被请求参数 / webhook 映射覆盖） |
| msg_type | 下拉 | ❌ | 默认消息类型：text / html / markdown |
| html_height | 数字 | ❌ | HTML 显示高度（px），仅 msgType=html 时生效 |
| default_url | 文本 | ❌ | 默认跳转链接（可被请求参数 / webhook 映射覆盖） |
| default_img_url | 文本 | ❌ | 默认通知图标 URL（216×216 PNG） |
| active | 开关 | ✅ | 启用 / 禁用（默认启用） |

> ⚠️ meow_nickname 一旦创建不可修改，但 token 可重新生成

### 4.2 Webhook 字段映射

外部服务通过以下入口推送：

```
POST /webhook/{token}
Body: { "action": "opened", "head_commit": { "message": "修复 bug" } }
```

通过 JSONPath 从 payload 映射到 Meow 字段：

| 字段 | 映射规则 | 说明 |
|------|---------|------|
| title | `$` 开头 → JSONPath 提取；无 `$` → 固定字面量；空 `""` → 跳过（降级 P2） | 消息标题 |
| msg | 同上 | 消息内容（必须提供） |
| url | 同上 | 跳转链接 |
| imgUrl | 同上 | 通知图标 URL |
| msg_type | 同上 | 消息类型 |

**预设模板（10 套）**：

| source_type | 显示名称 | 推荐 msg_type | title 映射 | msg 映射 |
|---|---|---|---|---|
| `github` | GitHub Webhook | text | `$.action` | `$.repository.full_name` |
| `github_action` | GitHub Actions | markdown | `$.workflow_run.event` | `$.workflow_run.head_commit.message` |
| `github_pr` | GitHub PullRequest | markdown | `$.pull_request.title` | `$.pull_request.body` |
| `jenkins` | Jenkins | text | `$.project.name` | `$.build.full_display_url` |
| `grafana` | Grafana 告警 | markdown | `$.alerts[0].labels.alertname` | `$.alerts[0].annotations.message` |
| `prometheus` | Prometheus AlertManager | markdown | `$.alerts[0].labels.alertname` | `$.alerts[0].annotations.description` |
| `zabbix` | Zabbix | markdown | `$.trigger.description` | `$.event.description` |
| `gotify` | Gotify | markdown | `$.message` | 整个 payload JSON 序列化 |
| `emby` | Emby | text | `$.Title` | `$.Description` |
| `generic` | 通用 | — | 空（手动配置） | 空（手动配置） |

### 4.3 参数优先级（核心）

```
P1 — 请求参数 / webhook 映射（最高）
  • query: msgType / url / imgUrl
  • body: title / msg / url / imgUrl / msg_type
  • webhook JSONPath 提取
  • webhook 固定字面量

P2 — 接口预设值（endpoints 表）
  meow_nickname / msg_type / default_title / html_height / default_url / default_img_url

P3 — Meow 默认值（最低）
  title="Meow" / msgType="text" / htmlHeight="200"
```

| 参数 | P1 请求/映射 | P2 接口预设 | P3 Meow 默认 | 说明 |
|------|------------|------------|------------|------|
| meow_nickname | — | endpoints.meow_nickname | — | 不可修改 |
| title | query/body/webhook | endpoints.default_title | "Meow" | 都空则用 P3 |
| msg | query/body/webhook | — | —（必须） | 不允许默认值 |
| msgType | query/body/webhook | endpoints.msg_type | "text" | 优先级最高 |
| htmlHeight | query | endpoints.html_height | "200" | 仅 msgType=html 时生效 |
| url | query/body/webhook | endpoints.default_url | 不传 | 都空则不传 |
| imgUrl | query/body/webhook | endpoints.default_img_url | 不传 | 都空则不传 |

### 4.4 管理功能

- **推送接口管理**：列表 / 创建 / 编辑 / 删除 / 启停 / 复制 token
- **推送日志**：列表 / 按 token/时间/状态筛选 / 清理过期日志
- **全局设置**：Meow API 地址 / 日志保留天数 / 频率限制 / 修改密码

---

## 五、API 设计

### 5.1 管理后台（JWT 鉴权）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/admin/login` | 验证密码 → 返回 JWT |
| `GET` | `/api/admin/endpoints` | 推送接口列表（分页/搜索/状态筛选） |
| `POST` | `/api/admin/endpoints` | 创建推送接口 |
| `GET` | `/api/admin/endpoints/:id` | 接口详情 |
| `PUT` | `/api/admin/endpoints/:id` | 更新推送接口 |
| `DELETE` | `/api/admin/endpoints/:id` | 删除推送接口 |
| `GET` | `/api/admin/push-logs` | 推送日志列表 |
| `DELETE` | `/api/admin/push-logs` | 清理过期日志 |
| `GET` | `/api/admin/webhook/presets` | 预设模板列表 |
| `GET` | `/api/admin/settings` | 全局设置 |
| `PUT` | `/api/admin/settings` | 更新全局设置 |
| `POST` | `/api/admin/change-password` | 修改管理员密码 |

### 5.2 推送入口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/webhook/{token}` | webhook 推送 |
| `GET` | `/verify/{token}` | 验证 token 有效性，返回接口信息 |

`POST /webhook/{token}` 支持的 query 参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `msgType` | text/html/markdown | 覆盖默认消息类型 |
| `url` | URL | 覆盖默认跳转链接 |
| `imgUrl` | URL | 覆盖默认通知图标 URL |

### 5.3 TG 劫持（二期）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET/POST` | `/tg-proxy/bot{token}/{method_name}?query` | 拦截写操作推 Meow，伪造 TG 响应返回 |

**Token 绑定机制**：

外部服务在 TG 推送配置中填入 **meowbridge 的 token**（而非真实 TG Bot Token），并将 `TG_API_URL` 指向 `https://meowbridge.xxx/tg-proxy`。

```yaml
# 外部服务 TG 推送配置示例
telebot:
  token: "meowbridge-user-token-xxxx"   # ← 填入 meowbridge 的 token，非真实 TG Bot Token
  chat_id: "随便填"                       # ← 无意义，TG 服务器不校验
  tg_api_url: "https://meowbridge.xxx"  # ← 指向 meowbridge
```

**处理逻辑**：

1. 提取 URL 路径中的 `bot{token}` → 查 endpoints 表找到对应接口的 meow_nickname
2. 提取 TG 请求 body 中的 text/caption → 提取为 Meow msg
3. 推送到 Meow API
4. 伪造 TG API 成功响应 `{"ok": true, "result": {}}` 返回给外部服务
5. 不转发到真实 Telegram（消息已通过 Meow 推送）

**劫持范围**：
- 写操作：sendMessage / sendPhoto / sendDocument 等（提取 text/caption 推 Meow）
- 读操作：getMe / getUpdates 等（伪造响应返回，不代理）

---

## 参考

- 推送 API 文档：https://www.chuckfang.com/MeoW/api_doc.html
