# Fleet TypeScript SDK

`@fleet/sdk` 封装 Fleet **Client API**（用户 JWT）与 **Server API**（API Key + `X-Fleet-Project`），便于集成测试与脚本调用。

## 目录

| 路径 | 说明 |
|------|------|
| `typescript/` | SDK 包 `@fleet/sdk` |
| `demo/` | 端到端演示（Server + Client 流程） |

## 快速开始

```bash
# 安装依赖并编译 SDK
task sdk-install
task sdk-build

# 启动本地 Fleet（另开终端）
task up
task migrate
go run ./cmd/seed   # 记下输出的 api_key
task dev-server

# 运行演示
FLEET_API_KEY=<seed 输出的 key> task sdk-demo
```

也可分步运行：

```bash
FLEET_API_KEY=... npm run demo:server --prefix sdk/demo
FLEET_DEMO_DB_ID=... FLEET_DEMO_COLL_ID=posts npm run demo:client --prefix sdk/demo
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `FLEET_ENDPOINT` | `http://localhost:8088` | Fleet HTTP 入口 |
| `FLEET_PROJECT_ID` | `default` | 项目 ID |
| `FLEET_API_KEY` | （必填） | Server API 密钥 |
| `FLEET_DEMO_EMAIL` | `sdk.demo.<随机>@fleet.local` | Client 演示账号 |
| `FLEET_DEMO_PASSWORD` | `Sdk@123456` | Client 演示密码 |
| `FLEET_DEMO_INVITE_EMAIL` | `invitee@fleet.local` | 团队邀请邮箱 |

## SDK 用法

```typescript
import { Fleet } from "@fleet/sdk";

// Server API
const admin = Fleet.withApiKey("http://localhost:8088", "default", process.env.FLEET_API_KEY!);
await admin.server.health.check();
await admin.server.databases.createDatabase({ id: "app", name: "App DB" });

// Client API（注册后自动保存 access token）
const client = Fleet.create({ endpoint: "http://localhost:8088", projectId: "default" });
await client.account.signUp({ email: "u@example.com", password: "Pass@123", name: "User" });
await client.databases.createDocument("app", "notes", { data: { title: "Hi" } });
```

## 已实现 API  surface

**Client：** Account（注册/登录/会话/偏好）、Databases 文档 CRUD、Teams 与 Memberships。

**Server：** Health、Projects、Users、Teams、Databases（库/集合/属性/文档）、API Keys、Storage（Bucket/File）。
