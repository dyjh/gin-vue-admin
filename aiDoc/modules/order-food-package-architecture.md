# `orderfood` 业务包架构与 GVA MCP 生成规范

## 1. 文档目的

本文定义“来干饭”后端业务包、管理后台目录、路由边界、领域拆分和 GVA MCP 使用方式，作为后端与后台 Web 实现前的技术基线。

关联文档：

- 产品口径：`../prd/family-menu-miniapp-prd.md`
- 后台产品：`../prd/order-food-admin-web-prd.md`
- 接口契约：`../frontend-backend/order-food-contract-baseline.md`
- 小程序接口：`../miniApp/api.json`
- 通用分层：`backend-layer-rules.md`

## 2. 架构决策

### 2.1 使用独立核心业务包

业务包统一命名为：

```text
orderfood
```

选择独立 `package`，不放入 `system`、`example`，也不做 GVA 插件。

原因：

- “来干饭”是本仓库的核心业务，不是可插拔的附加功能。
- 业务会同时服务微信小程序和后台管理端，需要共享领域模型和 Service。
- 放入 `system` 会污染框架基础能力，放入 `example` 会把正式业务和示例混合。
- 插件目录适合可独立安装、卸载的模块，本业务与当前数据库、用户、配置和管理后台深度集成。

### 2.2 分层不变

严格遵守：

```text
Router -> API -> Service -> Model
```

- Router：路由、分组和中间件。
- API：HTTP 参数绑定、校验、鉴权上下文转换和统一响应。
- Service：领域逻辑、事务、状态机和外部服务编排，不依赖 `gin.Context`。
- Model：持久化模型、请求模型和响应模型。
- `enter.go`：各层分组注册和对外聚合入口。

### 2.3 双客户端、共享领域

微信小程序和后台 Web：

- 共享数据库模型与核心 Service。
- 在 API、Router、request、response 层分开。
- 使用不同认证中间件和权限体系。
- 不让后台接口直接复用小程序 Handler。
- 不让小程序令牌进入 GVA 管理员权限域。

## 3. 代码目录

### 3.1 后端目录

```text
server/
  model/
    orderfood/
      enter.go
      user.go
      media.go
      catalog.go
      dish.go
      recipe.go
      recommendation.go
      checkin.go
      preference.go
      meal.go
      shopping.go
      points.go
      capability.go
      ai.go
      notification.go
      governance.go
      request/
        enter.go
        admin_*.go
        miniapp_*.go
      response/
        enter.go
        admin_*.go
        miniapp_*.go
  service/
    orderfood/
      enter.go
      user.go
      media.go
      catalog.go
      dish.go
      recipe.go
      recommendation.go
      checkin.go
      preference.go
      meal.go
      shopping.go
      points.go
      capability.go
      ai.go
      notification.go
      governance.go
  api/
    v1/
      orderfood/
        enter.go
        admin/
          enter.go
          *.go
        miniapp/
          enter.go
          *.go
  router/
    orderfood/
      enter.go
      admin/
        enter.go
        *.go
      miniapp/
        enter.go
        *.go
  initialize/
    gorm_biz.go
    router_biz.go
```

说明：

- 模型按领域聚合，不按每张表机械拆文件。
- request/response 按客户端分离，避免后台字段泄漏到小程序。
- `gorm_biz.go` 注册业务表迁移。
- `router_biz.go` 注册后台和小程序业务路由。
- 根层 `enter.go` 继续沿用仓库现有聚合方式，不引入新的全局容器。

### 3.2 管理后台目录

```text
web/src/
  api/
    orderfood/
      user.js
      points.js
      catalog.js
      dish.js
      recommendation.js
      media.js
      governance.js
      meal.js
      shopping.js
      capability.js
      ai.js
      notification.js
      dashboard.js
  view/
    orderFood/
      dashboard/
        index.vue
      user/
        index.vue
        detail.vue
      points/
        index.vue
      category/
        index.vue
      tag/
        index.vue
      unit/
        index.vue
      discoverableDish/
        index.vue
        detail.vue
      recommendation/
        index.vue
      officialDish/
        index.vue
        form.vue
      media/
        index.vue
      moderation/
        index.vue
      governance/
        index.vue
      meal/
        index.vue
        detail.vue
      shopping/
        index.vue
        detail.vue
      capabilityPolicy/
        index.vue
      aiProvider/
        index.vue
      aiModel/
        index.vue
      aiCapability/
        index.vue
      aiUsage/
        index.vue
        detail.vue
      notification/
        index.vue
      subscribeTemplate/
        index.vue
      subscribeLog/
        index.vue
```

GVA 菜单组件路径使用：

```text
view/orderFood/<domain>/index.vue
```

后台根菜单 path 建议为 `orderFood`，组件目录也统一使用 `orderFood`，避免 `order-food`、`order_food` 和 `orderfood` 混用。

## 4. 数据库命名

### 4.1 表前缀

业务表统一使用：

```text
of_
```

示例：

```text
of_users
of_dishes
of_dish_ingredients
of_dish_steps
of_meals
of_meal_candidates
of_meal_votes
of_meal_snapshots
of_shopping_lists
of_point_entries
of_capability_policies
of_user_capability_overrides
of_client_feature_labels
of_ai_usages
```

### 4.2 主键与公开 ID

- 数据库内部主键沿用 `global.GVA_MODEL` 的数值 ID。
- 面向小程序的业务对象增加稳定公开 ID，例如 `PublicID`，建议使用 ULID。
- 小程序 API 只返回公开 ID 字符串，不暴露内部自增主键。
- 后台生成式 CRUD 可以使用内部 ID；跨小程序对象查询时仍优先展示和支持公开 ID。
- 公开 ID 创建后不可修改，建立唯一索引。

这样既兼容 GVA 生成器和关联模型，也保持现有小程序 `api.json` 的字符串 ID 契约。

### 4.3 通用字段

除 GVA 基础字段外，按需要统一使用：

- `public_id`
- `created_by`
- `updated_by`
- `deleted_reason`
- `deleted_by`
- `deleted_at`
- `status`
- `sort_order`
- `version`

高并发状态对象使用乐观锁字段或条件更新，不依赖前端最后一次展示值。

## 5. 领域拆分

| 领域 | 主要职责 | 典型事务或约束 |
| --- | --- | --- |
| user | 微信用户、资料、禁用状态 | 登录时校验禁用 |
| media | 图片资源、对象存储、审核绑定 | 审核通过后才能绑定业务 |
| catalog | 分类、标签、单位 | 引用后只停用不物理删除 |
| dish | 菜品、配料、步骤、来源锁 | 封面必选、复制品不可发现 |
| recipe | 个人菜谱和菜品项 | 只引用当前用户可用菜品 |
| recommendation | 候选池、官方菜、推荐精选、复制 | 源变化自动下线推荐 |
| checkin | 做菜打卡、每日奖励 | 能力开启时每日首次奖励幂等 |
| preference | 用户口味画像 | 异步聚合，不阻塞主流程 |
| meal | 饭局、参与者、候选菜、点选、快照 | 确认时生成稳定快照 |
| shopping | 采购清单、清单项、只读分享 | 分享令牌仅可读 |
| points | 当前积分和积分流水 | 事务锁、幂等扣减与退款 |
| capability | 平台默认策略、紧急停用、用户覆盖、客户端动态文案 | 统一计算有效状态和配置版本 |
| ai | 供应商、模型、能力、任务、调用 | 成本、限额、失败退款 |
| notification | 站内通知、订阅消息 | 站内通知始终保留 |
| governance | 内容处理、复制链任务、审计 | 高风险操作可追踪 |

领域之间通过 Service 组合，不允许 API 层直接操作其他领域 Model。

## 6. 路由与认证

### 6.1 路由前缀

在全局 `/api` 之后：

```text
后台管理：/api/orderfood/...
微信小程序：/api/miniapp/v1/...
只读分享：/api/miniapp/v1/public/...
```

小程序 `api.json` 的 `basePath` 保持 `/api/miniapp/v1`。

### 6.2 后台认证

- 沿用 GVA 管理员 JWT。
- 路由挂载管理员鉴权和 Casbin API 权限。
- 写接口继续使用仓库已有操作记录中间件。
- 菜单权限、按钮权限和 API 权限同步注册。

### 6.3 小程序认证

- `App.onLaunch` 自动调用 `wx.login`，所有页面等待同一认证 Promise。
- 登录 API 使用一次性 `code` 调用微信 `jscode2session`，按 AppID 范围内 `openid` 查找或创建用户。
- `openid`、`unionid`、`session_key` 只存在后端，不进入小程序响应、日志或埋点。
- `wechat_open_id` 建立唯一索引，并发首次登录依靠唯一约束和事务避免重复用户。
- 登录成功后计算当前用户有效能力状态，业务访问令牌、个人资料和 `RuntimeConfig` 一次返回。
- 自定义中间件解析小程序用户身份并校验禁用状态。
- API 层把用户 ID 转为普通参数传入 Service。
- Service 不接收 `gin.Context`，也不自行解析令牌。
- 令牌失效后客户端以单飞锁重新登录；后端不支持客户端直接提交身份 ID 恢复会话。

### 6.4 小程序能力门禁

- 小程序 API 使用 `capability` Service 计算平台紧急停用、用户覆盖和平台默认值。
- 受控 Service 在业务执行前调用统一门禁，不能在各 Handler 中复制判断。
- 有效状态关闭时，能力接口返回 `46001`，积分接口返回 `46002`。
- 打卡 Service 在关闭状态下仍保存打卡，但跳过奖励流水和偏好分析任务。
- 登录与 `/bootstrap` 返回同版本 `RuntimeConfig`；策略变更后通过版本号使缓存失效。
- 小程序响应只使用中性能力编码和动态文案，不序列化供应商、模型、内部 AI 能力名或内部表字段。

### 6.5 公共只读接口

采购清单分享接口只接受高熵分享令牌：

- 不返回创建者敏感信息。
- 只允许读取。
- 令牌可失效、可过期。
- 错误响应不泄漏资源是否存在的额外信息。

## 7. Model、Request 与 Response 规则

### 7.1 Model

- 持久化字段使用明确 `json` 和 `gorm` 标签。
- 业务枚举使用具名字符串常量，禁止散落魔法字符串。
- 数据库模型不直接作为小程序完整响应。
- 外部供应商密钥只存安全引用或加密值，不在 JSON 中序列化。

### 7.2 Request

- 列表查询内嵌 GVA 通用 `request.PageInfo`。
- 管理端和小程序请求分别定义。
- ID 参数类型和路由来源保持一致。
- 写请求进行结构校验后，业务状态校验放在 Service。
- 幂等键从 `X-Idempotency-Key` 或明确 `requestId` 读取并统一传给 Service。

### 7.3 Response

- 列表只返回摘要字段。
- 详情按权限返回完整字段。
- 小程序对象返回公开 ID。
- 后台敏感字段默认脱敏。
- 时间统一序列化为 ISO 8601。
- 可空值使用 `null`，不使用空字符串冒充缺失值。

## 8. Service 与事务边界

以下流程必须手写 Service，不交给通用 CRUD：

### 8.1 图片上传与审核

1. 上传临时对象。
2. 调用同步内容审核。
3. 审核通过后创建可用图片资源。
4. 失败时删除临时对象或进入回收任务。
5. 业务保存时校验资源归属和状态并绑定。

### 8.2 推荐复制

1. 锁定或校验推荐内容可用状态。
2. 复制菜品、配料、步骤和图片引用策略。
3. 写入来源链和 `source_locked`。
4. 保存为个人可用、未公开菜品。
5. 记录复制次数，整个过程幂等。

### 8.3 能力策略与打卡联动

1. 读取已发布的平台策略版本。
2. 检查全局紧急停用。
3. 读取用户三态覆盖，未覆盖时使用平台默认值。
4. 生成 `RuntimeConfig`，未开启能力不下发文案。
5. 能力关闭时，积分 Service 拒绝用户查询和消耗。
6. 打卡 Service 仍保存记录，但不写奖励流水、不投递偏好分析任务。
7. 平台策略或用户覆盖修改后，使对应运行时配置缓存失效并写审计日志。

### 8.4 AI 扣分与退款

1. 使用请求幂等号查询已有结果。
2. 通过统一能力门禁校验当前用户有效状态。
3. 事务锁定用户积分行。
4. 校验余额并写消耗流水。
5. 提交扣分后执行外部模型调用。
6. 失败时使用关联流水幂等退款。
7. 写 AI 调用和站内通知。

外部 HTTP 调用不能长期占用数据库行锁。

### 8.5 饭局确认

1. 条件更新饭局状态，避免重复确认。
2. 读取候选和点选统计。
3. 校验最终制作选择和份数。
4. 生成独立饭局菜品快照。
5. 生成采购清单和清单项。
6. 写状态时间线和通知。

快照和采购清单生成必须在一个事务内保持一致。

### 8.6 内容治理

- 普通处理使用事务更新源内容、发现状态和推荐记录。
- 严重违规复制链处理使用任务表分批执行。
- 每一批可重试，且不能重复删除或重复通知。

## 9. 外部能力适配

外部依赖通过接口封装，不在业务 Service 中散落 SDK 调用：

```text
MediaStorage
ImageModeration
AIProvider
WechatAuth
WechatSubscribeMessage
Clock
IDGenerator
```

适配层负责供应商请求格式和错误转换；领域 Service 只依赖项目内接口。

外部错误至少区分：

- 参数或能力不支持。
- 身份或密钥错误。
- 限流。
- 超时。
- 服务不可用。
- 内容安全拒绝。

### 9.1 小程序与服务端命名边界

后端和管理后台内部可以使用准确的 AI 技术术语；小程序接口和发布包必须使用中性命名。

小程序侧统一使用：

```text
assist
feature
featureUsage
mealSuggestion
prepPlan
runtimeConfig
```

禁止把以下内部对象直接暴露给小程序：

```text
AIProvider
AIModel
AICapability
AIUsage
providerName
modelName
modelId
```

小程序运行路径需迁移为 `pages/ideas/*` 和 `pages/meal/prep-guide`。`/ai/*` 只允许存在于后台内部路由，不得成为小程序请求路径。

发布前对微信实际上传包扫描，而不是只扫描源码目录。扫描脚本放在 `miniApp/` 包之外；`.preview`、工具、文档和生产不需要的 Mock 必须通过打包忽略规则排除。

## 10. GVA MCP 实测基线

2026-07-24 已通过 `mcp:order_food_mini_app` 执行只读连接验证。

实测结果：

- `list_all_apis` 成功，数据库中有 190 条 API 记录。
- 当前 Gin 运行时 API 数量为 0，不能用运行时列表证明业务路由已注册。
- `list_all_menus` 成功，发现 10 个顶层菜单、43 个展开菜单节点。
- 菜单组件路径采用 `view/...` 格式。
- `gva_analyze` 识别到完整标准包 `system`、`example`。
- 识别到 `announcement`、`auto`、`email`、`plugin-tool` 等插件，其中部分插件缺少标准部件，不应复制其不完整结构作为模板。
- `query_dictionaries` 在不传字典 ID 时返回“字典ID不能为空”，与工具名称暗示的全量查询不一致；生成前不能依赖该调用自动获取全部字典。
- `gva_review` 是只读生成后检查工具。

结论：

- MCP 连接可用。
- 当前数据库里尚无“来干饭”业务 API。
- 运行时路由可见性需要在后端启动并完成注册后重新验证。
- 字典查询缺陷必须作为生成前检查项记录，不能默默跳过。

## 11. MCP 工具使用边界

### 11.1 适合自动生成

仅对字段稳定、逻辑简单的后台 CRUD 使用 `gva_execute`：

- 菜品分类。
- 菜品标签。
- 配料单位。
- 小程序能力文案基础记录。
- AI 供应商基础记录。
- AI 模型基础记录。
- AI 能力基础配置。
- 订阅消息模板。

即使使用生成器，也必须补充唯一索引、引用保护、密钥脱敏和业务校验。

### 11.2 必须手写

- 微信 `code` 自动登录、小程序令牌和并发首次登录唯一性。
- 平台默认策略、全局紧急停用、用户三态覆盖和运行时配置计算。
- 小程序发布包中性命名迁移、打包忽略和扫描门禁。
- 图片上传、对象存储和同步审核。
- 菜品保存、推荐复制和来源锁。
- 菜谱菜品归属校验。
- 能力开启条件下的打卡奖励和偏好画像。
- 积分扣减、退款和人工调整。
- 不知道吃什么召回与 AI 回退。
- 饭局加入、点选、关闭和确认。
- 快照、采购生成和只读分享。
- 内容治理和复制链。
- 通知编排。

### 11.3 禁止重复注册

`gva_execute` 如果启用：

- `autoCreateApiToSql`
- `autoCreateMenuToSql`
- `autoCreateBtnAuth`

则不得随后再用 `create_api`、`create_menu` 或手工 SQL 重复创建同一记录。

生成前必须先决定由谁负责：

| 资源 | 建议责任方 |
| --- | --- |
| 稳定 CRUD API | `gva_execute` 自动生成 |
| 自定义业务 API | 手写代码后用受控方式注册 |
| 后台根菜单和业务菜单层级 | 手工规划后统一创建 |
| 按钮权限 | 页面动作确定后批量创建 |
| 小程序 API | 不创建后台菜单，按业务路由手写 |

## 12. MCP 标准执行流程

### 12.1 生成前

1. 调用 `list_all_apis`、`list_all_menus`，保存当前基线。
2. 调用 `gva_analyze` 检查目标 package 和相似标准模块。
3. 确认模块字段、索引、枚举、权限和菜单路径。
4. 明确自动创建 API、菜单和按钮的开关。
5. 对工具无法查询的字典手工确认，不能猜测字典 ID。

### 12.2 生成

1. 先创建 `orderfood` package 骨架。
2. 每次只生成一个稳定 CRUD 模块。
3. 保持 `generateServer`、`generateWeb` 与实际交付范围一致。
4. 只在字段已冻结时启用 `autoMigrate`。
5. 保存 MCP 返回的模块、API、菜单和文件清单。

### 12.3 生成后

1. 调用 `gva_review`。
2. 检查 Model、Service、API、Router、`enter.go` 和 initialize 注册。
3. 检查 API 数据库记录和菜单是否重复。
4. 检查 Swagger 路径、参数来源和响应。
5. 检查管理端 API 封装和页面组件路径。
6. 运行 Go 格式化、编译和相关测试。
7. 启动后再次核对 Gin 运行时路由。

任何一步失败都先修复当前模块，不连续批量生成更多模块。

## 13. API 与菜单命名

### 13.1 后台 API

示例：

```text
GET    /api/orderfood/categories
POST   /api/orderfood/category
PUT    /api/orderfood/category
DELETE /api/orderfood/category/:id

GET    /api/orderfood/discoverable-dishes
POST   /api/orderfood/recommendations
POST   /api/orderfood/recommendations/:id/publish
POST   /api/orderfood/recommendations/:id/offline
```

避免在路径中使用 `getXxxList`、`editXxx` 等 RPC 式重复语义，除非必须与仓库既有生成模式保持一致。最终风格在同一业务包内保持统一。

### 13.2 小程序 API

继续以领域和资源组织：

```text
/api/miniapp/v1/dishes
/api/miniapp/v1/recipes
/api/miniapp/v1/meals
/api/miniapp/v1/shopping-lists
/api/miniapp/v1/notifications
```

动作型状态变化使用清晰子资源或动作：

```text
POST /meals/:id/close
POST /meals/:id/confirm
POST /recommendations/:id/copy
```

### 13.3 菜单

- 菜单 name 使用稳定英文标识。
- title 使用中文业务名称。
- component 与真实 Vue 文件路径一致。
- 路由 path 不携带 `/api`。
- 页面按钮权限与后端具体写 API 对应。

## 14. 测试策略

### 14.1 单元测试

- 状态机。
- 份数和采购汇总。
- 推荐来源锁。
- AI 推荐结果校验。
- 积分计算和每日首次打卡。

### 14.2 数据库集成测试

- 幂等扣分和退款。
- 并发饭局确认。
- 推荐源下线联动。
- 候选菜唯一约束。
- 分享令牌只读权限。

### 14.3 API 契约测试

- 与 `../miniApp/api.json` 对照。
- 响应 envelope 和分页。
- ID、时间、null 和枚举。
- 401、403、404、409、422、429 和 AI 错误。

### 14.4 后台权限测试

- 菜单不可见不代表 API 可访问。
- 只读角色不能调用写 API。
- 内容治理和积分调整需要独立权限。
- 管理员令牌不能调用小程序用户写接口。

## 15. 实施分批

### 批次 A：骨架与稳定配置

- `orderfood` package。
- 分类、标签、单位。
- 路由、迁移、权限和后台菜单。

### 批次 B：用户、媒体和菜品

- 微信用户与令牌。
- 图片资源和审核。
- 菜品、配料、步骤。
- 后台用户与候选池。

### 批次 C：推荐、菜谱、打卡和积分

- 推荐复制与来源锁。
- 菜谱。
- 打卡和画像。
- 积分事务。

### 批次 D：饭局和采购

- 饭局状态机。
- 点选统计、快照和确认。
- 采购清单和只读分享。
- 后台只读诊断。

### 批次 E：AI、通知和治理

- AI 适配与能力配置。
- 调用记录、扣分和退款。
- 通知和订阅消息。
- 内容治理、审计和运营概览。

每个批次都必须完成后端、后台页面、小程序联调和测试后再进入下一批。
