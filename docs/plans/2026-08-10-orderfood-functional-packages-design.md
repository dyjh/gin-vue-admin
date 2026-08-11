# OrderFood 功能域包重构设计

## 目标

把 `server/router`、`server/api/v1`、`server/service` 和 Front 对应三层的平铺业务文件迁入功能子包，并严格沿用既有 `system` 模块的组装方式。重构不提供旧扁平字段、类型别名、转发构造器或兼容入口。

外部 HTTP 路径、Swagger 契约、权限码、数据库模型和业务行为保持不变。

## 后台功能域

| 功能域 | 职责 |
| --- | --- |
| `user` | 小程序用户、微信配置 |
| `engagement` | 积分、订阅场景、通知记录 |
| `content` | 用户内容、媒体、审核、违规治理 |
| `dish` | 分类标签单位、官方菜品、推荐、建议目录 |
| `meal` | 饭局与采购清单管理；域内按资源分别提供 Router、API 和 Service |
| `ai` | AI 平台、模型、提示词、调用记录 |
| `dashboard` | 管理端运营概览 |
| `audit` | 管理操作审计 |
| `common` | 权限、幂等、访问审计和跨域小型工具 |

`governance` 属于内容治理，因此保留在 `content`。原 `operations.go` 混合的仪表盘和 AI 调用记录分别进入 `dashboard` 与 `ai`。

## Front 功能域

Front API 和 Router 按对外能力分为：

- `auth`
- `system`
- `profile`
- `content`
- `engagement`
- `meal`
- `assist`
- API 共享绑定辅助位于 `api/common`

Front Service 按真实依赖边界分为：

- `user`：登录、个人资料、微信订阅消息投递。
- `experience`：运行时、内容、上传、互动、饭局、增强能力、偏好。这里的服务存在事务内双向协作，放在同一包内避免人为接口和导入环。
- `system`：显式测试服务。

## 统一组装规则

调用关系完全参考现有 `system`：

1. 每个 `service/<domain>/enter.go` 定义本域 `ServiceGroup`；根 `service/enter.go` 只聚合嵌套域组。
2. 每个 `api/v1/<domain>/enter.go` 从根 `service.ServiceGroupApp.<DomainServiceGroup>` 绑定包级服务变量；处理器直接使用这些变量。
3. 每个 `router/<domain>/enter.go` 从根 `api.ApiGroupApp.<DomainApiGroup>` 绑定包级 API 变量；路由方法不接收 API 参数。
4. 初始化入口先取得域 RouterGroup，再调用域内注册方法。
5. Front 三层使用同一规则。

例如：

```go
// api/v1/content/enter.go
var contentService = service.ServiceGroupApp.ContentServiceGroup.Content

// router/content/enter.go
var contentApi = api.ApiGroupApp.ContentApiGroup.ContentApi
```

根聚合器不再暴露 `ServiceGroupApp.Content`、`ApiGroupApp.ContentApi`、`RouterGroupApp.InitContentRouter` 等旧扁平入口。

## 架构决策：域内资源边界

- 状态：已采用。
- 背景：`meal` 是一个完整功能域，但饭局和采购清单是两个顶级 HTTP 资源，具有不同的路由、权限和返回模型。
- 决策：域包内继续按外部资源保持清晰边界。`meal` 三层分别暴露 `Meal` 与 `ShoppingList` 组件，不使用一个 `MealAdmin` 类型承载两个顶级 HTTP 资源，也不提供旧名称的兼容别名或转发入口。
- 权衡：文件和聚合字段略有增加，但调用链、权限归属和后续演进边界更直观；饭局详情仍可通过包内共享函数读取采购清单摘要，不为复用引入跨包接口。

后台业务路由的 `/orderfood` 根分组由初始化入口统一创建，再传给各功能域 Router。域 Router 只注册自身的资源相对路径；需要操作审计中间件时创建空路径子组，避免把中间件扩散到共享根组。

后台 API 从共享辅助包读取 `CurrentAdminActor`。该名称只表达当前管理员上下文，不携带 `content` 等具体领域名称。API 不通过每次请求执行 `available()` 来兜底包级 Service 绑定；聚合入口必须保证依赖组装有效，处理器沿用 `system` 风格直接调用已绑定 Service。

后台业务 Service 与 `system` 保持相同的数据库生命周期：方法执行时通过 `database()` 取库，测试显式注入的 `Service.DB` 优先，否则读取 `global.GVA_DB`。包初始化阶段不绑定数据库，也不在数据库初始化或重载时遍历修改 Service 对象。

## 架构决策：Model 功能域

- 状态：已采用。
- 背景：原业务实体集中在 `server/model/*.go`，管理端 DTO 集中在 `model/request` 和 `model/response`。一个文件同时承载多个无关领域模型，导致 API、Service 和 Front 即使已经按域拆包，仍然依赖同一个扁平模型包。
- 决策：持久化实体按 `ai`、`audit`、`common`、`content`、`dish`、`engagement`、`meal`、`user` 分包；每个业务域按需包含自己的 `request`、`response`。运营看板没有持久化实体，只保留 `dashboard/request` 和 `dashboard/response`。分页、管理员、用户引用、目录引用和幂等请求头放在 `model/common` 的对应子包。
- 文件规则：一个实体文件只承载一个持久化聚合；仅约束该实体字段的枚举类型和常量与实体放在同一文件，只有被多个独立实体共同拥有的领域枚举才单独成文件。饭局与采购清单、看板与 AI 调用记录分别拆文件，避免重新形成按历史接口命名的混合文件。
- 依赖规则：业务代码直接导入目标域包，旧 `model`、`model/request`、`model/response` 不提供别名或转发兼容。数据库表名、GORM 标签、JSON 字段和 HTTP 契约保持不变。
- 权衡：跨域聚合需要显式导入多个模型包，但依赖方向和数据所有权可见，修改一个领域时不再被迫经过全局模型包。

## 根目录约束

- `server/service`：只保留 `enter.go`。
- `server/api/v1`：只保留 `enter.go`。
- `server/router`：只保留 `enter.go` 和根级路由测试。
- Front 的 `service`、`api`、`router` 根目录各只保留 `enter.go`。
- `server/model`：只保留跨域约束测试；业务实体进入功能域目录，业务 DTO 进入各域的 `request`、`response`。
- 既有 `system`、`example` 保持原结构。

## 验证标准

1. `go test ./... -run '^$'` 全仓编译通过。
2. `go test ./...` 全量测试通过，或明确记录环境型失败。
3. `git diff --check` 通过。
4. 搜索不到旧扁平聚合字段和为兼容而新增的别名/转发构造器。
5. 不改动用户现有的 `miniApp` 工作区内容。
