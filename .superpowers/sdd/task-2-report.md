# Task 2 报告

## 实现内容

- 新增 SQLite Store、核心数据模型和幂等迁移。
- 新增首次启动初始化：创建 bcrypt 管理员密码，并仅在设置不存在时写入 MeoW API 地址与日志保留天数。
- 新增设置读取和 UPSERT 写入接口。
- 在程序启动时打开数据库、迁移并执行初始化。
- 调整配置校验：仅首次数据库初始化需要 `ADMIN_PASSWORD`；管理员存在后启动不再依赖该环境变量。

## 测试

- `rtk go test ./internal/config ./internal/store`
- `rtk go test ./internal/store ./...`

两条命令均通过。

## 提交

- `2089ee0 添加数据库迁移和启动初始化`

## 说明

`internal/config` 的最小调整用于满足“管理员密码写入数据库后不再依赖 `ADMIN_PASSWORD`”这一任务全局约束；首次启动缺少密码会由 `Store.Bootstrap` 明确失败。

## 评审修复（持久化设置优先）

- `config.Load` 仅保留每次启动都需要的 `JWT_SECRET` 校验，不再在打开数据库前要求 `MEOW_API_BASE_URL`。
- `Store.Bootstrap` 仅在 `admin_users` 或 `meow_api_base_url` 设置缺失时，分别校验 `ADMIN_PASSWORD` 或 `MEOW_API_BASE_URL`；校验在写入前完成，避免半初始化状态。
- 补充覆盖：第二次 Bootstrap 不覆盖 `log_retention_days`；`SetSetting` 的更新语义；`GetSetting` 对缺失值返回 `ErrNotFound`。

### 验证

- `rtk go test ./internal/config ./internal/store ./...`：10 项测试通过（4 个包）。
