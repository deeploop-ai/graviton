# Orionid TypeScript SDK

`@orionid/sdk` 封装 Orionid **Client API**（用户 JWT）与 **Server API**（API Key + `X-Orionid-Project`），便于集成测试与交互演示。

## 目录

| 路径 | 说明 |
|------|------|
| `typescript/` | SDK 包 `@orionid/sdk` |
| `demo/` | Web 演示站点（注册/登录 + SDK 功能演示） |

## 快速开始

```bash
# 安装依赖并编译 SDK
task sdk-install
task sdk-build

# 启动本地 Orionid（另开终端）
task up
task migrate
go run ./cmd/seed   # 记下输出的 api_key
task dev-server

# 启动 Web 演示（默认 http://localhost:5174）
task sdk-demo
```

复制 `sdk/demo/.env.example` 为 `sdk/demo/.env` 可调整默认 Endpoint / Project ID。

## Web 演示站点

演示站点提供完整的前端体验：

| 页面 | 说明 |
|------|------|
| `/register` `/login` | 用户注册与登录（Client Account SDK） |
| `/app/account` | me / prefs / sessions / refresh |
| `/app/databases` | Server + Client Databases API 全功能验证 |
| `/app/teams` | 建队、刷新 Token、邀请成员 |
| `/app/server` | Health / Projects / Users / Teams / Databases |
| `/app/settings` | Endpoint、Project ID、API Key 配置 |

Server API 相关功能需在设置页填写 `go run ./cmd/seed` 输出的 API Key。

## SDK 用法

```typescript
import { Orionid } from "@orionid/sdk";

// Server API
const admin = Orionid.withApiKey("http://localhost:9080", "default", apiKey);
await admin.server.health.check();

// Client API（注册后自动保存 access token）
const client = Orionid.create({ endpoint: "http://localhost:9080", projectId: "default" });
await client.account.signUp({ email: "u@example.com", password: "Pass@123", name: "User" });
await client.databases.createDocument("app", "notes", { data: { title: "Hi" } });
```

## 已实现 API surface

**Client：** Account（注册/登录/会话/偏好）、Databases（文档 CRUD + count）、Teams 与 Memberships。

**Server：** Health、Projects、Users、Teams、Databases（库/集合/属性/索引/文档/Bulk）、API Keys、Storage（Bucket/File）。
