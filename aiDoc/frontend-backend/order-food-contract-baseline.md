# “来干饭”前后台接口契约基线

## 1. 文档目的

本文把小程序定稿交互、产品规则、gin-vue-admin 工程约束和后台 Web 需求转换为可直接实施的跨端契约。

适用范围：

- 微信小程序 `miniApp/`
- Go 后端 `server/`
- 管理后台 `web/`
- 小程序接口清单 `../miniApp/api.json`
- 微信身份、能力开关与发布合规专项基线 `miniapp-auth-capability-compliance.md`
- Swagger 和后续契约测试

本文不替代完整 PRD，而是负责消除实现层歧义。

## 2. 信息优先级

发生冲突时按以下顺序处理：

1. `../prd/family-menu-miniapp-prd.md` 中明确标记的现行实施口径。
2. `miniapp-auth-capability-compliance.md` 中的身份、开关、积分联动和发布包规则。
3. 本文中的状态、字段、权限和幂等规则。
4. 完成归一化后的 `../miniApp/api.json`。
5. 小程序已定稿运行页面体现的真实交互。
6. `../prd/order-food-admin-web-prd.md` 的后台页面需求。
7. 设计图和历史任务记录。

如果当前 `api.json` 与本文冲突，必须先修订 `api.json` 再开始对应接口编码，不能让前端和后端各自解释。

## 3. 已冻结的产品口径

### 3.1 菜品

- 菜品封面必选。
- 简介选填。
- 做法步骤必须有文字，步骤图选填。
- 用户端菜品状态只有 `draft`、`usable`。
- 删除是软删除和审计行为，不作为第三个用户可见状态。
- 只有 `usable` 菜品可加入菜谱、饭局候选或开启允许被发现。
- “允许被发现”是布尔值，不存在独立审核状态或审核队列。
- 管理员精选进入推荐即为平台内容把关行为。
- 用户编辑已允许被发现的菜品后，系统自动关闭发现并下线相关推荐。
- 来自平台推荐或 AI 推荐的个人副本默认 `usable`、未公开，可编辑，但永久不能再次开启允许被发现。

### 3.2 菜谱

- 菜谱只引用当前用户自己的 `usable` 菜品。
- 列表封面默认使用菜谱中第一道菜的封面。
- 菜谱简介选填，列表超长时单行省略。
- 菜谱删除不删除个人菜品库中的菜。

### 3.3 饭局

- 状态为 `collecting`、`closed`、`confirmed`、`cancelled`。
- 不使用独立 `expired` 终态。
- 截止时间到达时，`collecting` 转为 `closed`，并记录 `closeReason=deadline`。
- 创建者提前关闭时记录 `closeReason=manual`。
- 已关闭饭局不能重新开启。
- 同一微信用户不能重复加入同一饭局。
- 同一饭局、参与者和候选菜只保留一条点选记录。
- 确认最终菜单时才生成完整饭局菜品快照和采购清单。

### 3.4 采购清单

- 清单项不分类。
- 清单项只有待购买和已购买两种完成状态。
- 创建者可新增、编辑、删除和勾选。
- 微信分享进入独立只读页面。
- 分享访问者不能修改原清单。

### 3.5 AI 与推荐

- 数据库权限和硬条件过滤由 SQL 完成，AI 不参与。
- 推荐库样本不足时才允许 AI 补充。
- AI 推荐必须是现实存在的菜品、常见材料和可验证做法。
- AI 失败或超时必须幂等退款并通知用户。
- 打卡偏好分析异步执行，不阻塞打卡保存。
- 平台增强能力默认关闭。
- 用户覆盖状态为 `inherit|enabled|disabled`，全局紧急停用优先级最高。
- 当前用户增强能力关闭时，积分、积分流水、打卡奖励和打卡偏好分析同时关闭。
- 小程序发布包和小程序 API 使用中性能力命名；可见文案全部由后端动态下发。

### 3.6 通知

- 通知默认未读。
- 默认进入“未读”页签。
- 列表每页默认 20 条，界面无限滚动。
- “全部已读”只更新阅读状态，不删除历史记录。

## 4. 通用传输约定

### 4.1 基础路径

```text
微信小程序：/api/miniapp/v1
后台管理：/api/orderfood
公共只读分享：/api/miniapp/v1/public
```

### 4.2 请求头

```http
Authorization: Bearer <accessToken>
X-Request-Id: <traceId>
X-Idempotency-Key: <idempotencyKey>
Content-Type: application/json
```

- `X-Request-Id` 可由客户端传入，也可由服务端生成。
- 所有写操作响应应回传追踪 ID，便于跨端定位。
- 关键写操作必须接受幂等键。
- 文件上传使用 `multipart/form-data`，不能强行包装为 JSON。

### 4.3 成功响应

```json
{
  "code": 0,
  "data": {},
  "msg": "ok"
}
```

### 4.4 失败响应

```json
{
  "code": 40901,
  "data": null,
  "msg": "饭局已关闭，不能继续点菜"
}
```

- HTTP 状态和业务 `code` 均应表达真实语义。
- `msg` 面向用户或管理员，可直接展示但不包含密钥、SQL 或堆栈。
- 外部供应商原始错误仅写安全日志，响应使用项目错误码。

## 5. 分页与排序

### 5.1 请求

```json
{
  "page": 1,
  "pageSize": 20
}
```

- `page` 从 1 开始。
- 小程序默认 20，后台默认 20。
- 后台可选 20、50、100。
- 服务端限制最大 `pageSize`。

### 5.2 响应

```json
{
  "page": 1,
  "pageSize": 20,
  "total": 42,
  "list": []
}
```

### 5.3 稳定排序

- 通知、积分流水：`createdAt DESC, id DESC`。
- 一般后台配置：`sortOrder ASC, id ASC`。
- 无限滚动追加时客户端按公开 ID 去重。
- 前端切换筛选、下拉刷新或搜索时重置为第 1 页。

第一版继续使用页码式服务端分页；界面不显示页码不代表接口一次返回全部数据。

## 6. ID、时间和空值

### 6.1 ID

- 小程序业务对象 ID 均为字符串公开 ID。
- 数据库内部 GVA 数值 ID 不在小程序响应中暴露。
- 后台配置 CRUD 可以使用数值 ID。
- 后台业务诊断页面同时支持按公开 ID 查询。

### 6.2 时间

- 接口统一返回 ISO 8601，例如 `2026-07-24T10:30:00+08:00`。
- 数据库存储和服务端比较使用统一时区策略。
- 前端负责本地化显示，不从中文日期反解析时间。

### 6.3 空值

- 选填且不存在的对象、字符串或时间返回 `null`。
- 空列表返回 `[]`。
- 禁止用 `undefined`、`"null"`、`0` 或空字符串混淆缺失值。

## 7. 认证与身份边界

### 7.1 小程序用户

- 小程序冷启动由 `App.onLaunch` 自动调用 `wx.login`。
- 所有页面请求等待同一个启动认证 Promise，禁止页面各自重复登录。
- 微信 `code` 只用于登录换取业务令牌，不能持久化、写日志或进入埋点。
- 后端通过 `jscode2session` 取得 `openid`，以 AppID 范围内的 `openid` 唯一识别用户。
- `openid`、`unionid`、`session_key` 不返回小程序。
- 登录接口不接受客户端提交的 `openid`、`unionid` 或业务 `userId`。
- 后续接口使用业务访问令牌。
- 后端从令牌取得用户身份，不接受客户端传入的 `userId` 作为授权依据。
- 令牌失效时客户端通过单飞锁重新执行一次微信登录；只重放允许安全重试的请求。
- 被禁用用户登录和调用核心接口时返回稳定禁用错误及可展示原因。

登录成功响应必须同时返回当前用户的 `RuntimeConfig`。页面在认证和运行时配置都就绪前不能渲染受控入口。

### 7.2 后台管理员

- 沿用 GVA 管理员 JWT、角色、Casbin API 权限和按钮权限。
- 高风险操作需要独立权限点。
- 管理员接口不能使用小程序用户令牌。

### 7.3 分享访问者

- 只通过高熵分享令牌读取采购清单。
- 分享令牌不等同于登录令牌。
- 分享接口不能返回后台字段、创建者隐私或可写操作地址。

## 8. 核心枚举

### 8.1 菜品

```text
DishStatus:
  draft
  usable
```

删除通过 `deletedAt`、`deletedBy`、`deletedReason` 表达。

```text
RecommendationSourceType:
  official
  creator
```

```text
RecommendationStatus:
  draft
  published
  offline
```

### 8.2 饭局

```text
MealStatus:
  collecting
  closed
  confirmed
  cancelled
```

```text
MealCloseReason:
  manual
  deadline
  null
```

### 8.3 积分

```text
PointEntryType:
  earned
  spent
  refund
  adjustment
```

### 8.4 能力策略

```text
CapabilityOverride:
  inherit
  enabled
  disabled
```

```text
FeatureUsageStatus:
  pending
  running
  succeeded
  failed
  refunded
```

### 8.5 图片审核

```text
MediaReviewStatus:
  pending
  passed
  rejected
  error
  not_required
```

## 9. 小程序核心模型

本节补齐 `api.json` 中实现前必须定义的类型。

### 9.1 运行时配置

`RuntimeConfig`：

```json
{
  "policyVersion": 1,
  "enhancedFeaturesEnabled": false,
  "pointsEnabled": false,
  "features": []
}
```

`ClientFeature`：

```json
{
  "code": "cover_create",
  "title": "生成封面",
  "actionText": "生成封面",
  "description": "根据菜品信息生成封面图",
  "enabled": true,
  "pointCost": 8,
  "freeQuota": 0,
  "sortOrder": 20
}
```

规则：

- 能力编码仅允许 `dish_extract`、`cover_create`、`meal_suggest`、`prep_sequence`、`taste_profile`。
- 未对当前用户开启的能力不返回文案。
- `enhancedFeaturesEnabled=false` 时 `features=[]` 且 `pointsEnabled=false`。
- 客户端配置缺失时隐藏入口，不设置包含 AI 相关字样的静态兜底。

### 9.2 配料

```json
{
  "id": "string",
  "name": "鸡蛋",
  "amount": "2",
  "unit": "个",
  "note": null,
  "sortOrder": 1
}
```

`IngredientInput`：

```json
{
  "name": "string|required|max:30",
  "amount": "string|required|max:20",
  "unit": "string|required|max:10",
  "note": "string|optional|max:80",
  "sortOrder": "number|required|min:1"
}
```

### 9.3 做法步骤

```json
{
  "id": "string",
  "text": "鸡蛋打散，加入少量盐。",
  "imageUrl": null,
  "sortOrder": 1
}
```

`DishStepInput`：

```json
{
  "text": "string|required|max:500",
  "imageFileId": "string|null",
  "sortOrder": "number|required|min:1"
}
```

### 9.4 作者

`AuthorSummary`：

```json
{
  "id": "string|null",
  "name": "来干饭官方",
  "avatarUrl": null,
  "type": "official"
}
```

- `type` 为 `official|creator`。
- 官方内容不虚构个人作者。

### 9.5 分类和标签

`DishCategory`：

```json
{
  "id": "string",
  "name": "家常菜",
  "sortOrder": 1,
  "enabled": true
}
```

`DishTag`：

```json
{
  "id": "string",
  "name": "快手",
  "sortOrder": 1,
  "enabled": true
}
```

### 9.6 饭局候选菜

`MealCandidate`：

```json
{
  "id": "string",
  "dishId": "string",
  "name": "番茄炒蛋",
  "coverUrl": "https://...",
  "category": "家常菜",
  "tags": ["快手"],
  "voteCount": 2,
  "selectedByMe": true,
  "available": true,
  "sortOrder": 1
}
```

- 点餐操作使用候选菜 ID，不直接使用菜品 ID。
- 候选菜保留稳定展示摘要。
- 原菜品删除或不可用时，已有候选项可展示不可用标记，是否移除由创建者在确认前处理。

### 9.7 备菜步骤

`PrepStep`：

```json
{
  "id": "string",
  "order": 1,
  "title": "鸡腿肉先焯水",
  "instruction": "鸡腿肉冷水下锅，加姜片煮至浮沫出现。",
  "parallelActions": [
    "等待水开时切西兰花和葱姜"
  ],
  "waitMinutes": 6,
  "completionHint": "撇净浮沫后捞出冲洗"
}
```

备菜计划描述实际操作顺序和可并行事项，不使用“启动任务”等生硬系统文案。

## 10. 菜品契约

### 10.1 保存

- `coverFileId` 必填。
- `description` 可为 `null`。
- 标签最多 3 个。
- 每个步骤必须有文字。
- 保存为 `usable` 时必须有至少一项配料和一步做法。
- 保存草稿可以放宽业务完整性，但封面是否允许暂缺必须由同一接口规则明确；第一版 UI 提交正式菜品时封面必选。

### 10.2 允许被发现

单独动作接口优于通用更新：

```text
POST /dishes/:id/discoverability
```

请求：

```json
{
  "discoverable": true
}
```

校验：

- 仅本人。
- 状态必须为 `usable`。
- `sourceLocked=false`。
- 内容未被治理删除。

不返回“审核中”状态。

### 10.3 删除

- 用户删除为软删除。
- 同步清理当前菜品专属图片资源和对象存储文件。
- 已确认饭局快照和采购文字数据继续保留。
- 普通用户自删不影响其他用户已复制副本。

## 11. 推荐契约

- 推荐列表和详情为只读。
- 详情返回作者、来源和是否已复制。
- 加入菜品库使用幂等键。
- 同一用户重复提交返回同一副本或明确已复制结果，不创建多份。
- 复制后设置 `sourceLocked=true`、`discoverable=false`。
- 源推荐普通下线不删除已有副本。
- 严重违规复制链处理由治理任务执行。

## 12. 饭局状态机

允许流转：

```text
collecting -> closed
collecting -> cancelled
closed -> confirmed
closed -> cancelled
```

截止任务：

```text
collecting + deadlineAt <= now
  => closed + closeReason=deadline
```

禁止：

- `closed` 返回 `collecting`。
- `confirmed` 修改候选、点选或最终份数。
- `cancelled` 继续参与。
- 重复确认生成第二份快照或采购清单。

确认接口必须使用条件更新和事务，状态冲突返回 `40901`。

## 13. 能力开关、积分与调用幂等

### 13.1 有效能力状态

有效状态按以下顺序计算：

```text
platform.suspended
  => false
user.override == enabled
  => true
user.override == disabled
  => false
otherwise
  => platform.defaultEnabled
```

- 平台默认值初始化为 `false`。
- `pointsEnabled` 必须等于当前用户的有效能力状态。
- 有效状态关闭时，受控接口在 Service 层返回 `46001`。
- 积分汇总和流水接口返回 `46002`。
- 打卡仍保存，但 `rewardGranted=false`、`rewardAmount=0`，不写积分流水、不触发偏好分析。
- 已有积分和流水只隐藏，不清零、不删除。

### 13.2 幂等范围

以下操作必须幂等：

- AI 能力调用。
- AI 失败退款。
- 推荐复制。
- 打卡每日首次奖励。
- 饭局确认。
- 采购清单生成。
- 后台人工积分调整。

### 13.3 外部能力请求

- `requestId` 作为业务幂等号。
- 同一用户、能力和 `requestId` 只能有一个最终结果。
- 重试时返回已有执行状态或结果。
- 失败退款通过原消耗流水关联，最多退款一次。

### 13.4 外部调用

- 不在持有用户积分数据库行锁期间执行长时间模型请求。
- 调用状态从 `pending` 到 `running` 再到终态。
- 超时任务由补偿任务收口，不依赖客户端保持连接。

## 14. 错误码

保留现有通用错误：

| code | 含义 |
| ---: | --- |
| 40001 | 请求参数错误 |
| 40101 | 登录已过期 |
| 40102 | 微信登录凭证无效、已使用或兑换失败 |
| 40301 | 无权操作资源 |
| 40401 | 资源不存在 |
| 40901 | 状态冲突或重复操作 |
| 42201 | 图片内容审核不通过 |
| 42901 | 请求过于频繁 |
| 46001 | 当前功能未开放 |
| 46002 | 积分功能未开放 |
| 47001 | 积分不足 |
| 47002 | 当前功能未解锁或不可用 |
| 47003 | 处理失败，积分已退还 |

需要在实现时补充并稳定：

| 建议 code | 含义 |
| ---: | --- |
| 40302 | 用户已被禁用 |
| 40902 | 幂等请求参数与原请求不一致 |
| 40903 | 资源已复制或已生成 |
| 42202 | 图片资源未审核通过或不属于当前用户 |
| 47004 | 免费次数或日限额已用完 |
| 47005 | 结果业务校验失败 |

错误码一旦被小程序使用，不随展示文案调整而改变。

## 15. 后台 API 家族

后台第一版接口按领域组织：

```text
/orderfood/dashboard
/orderfood/users
/orderfood/platform-capability-policy
/orderfood/user-capability-overrides
/orderfood/client-feature-labels
/orderfood/point-entries
/orderfood/point-adjustments
/orderfood/categories
/orderfood/tags
/orderfood/units
/orderfood/discoverable-dishes
/orderfood/recommendations
/orderfood/official-dishes
/orderfood/media
/orderfood/moderation-records
/orderfood/governance-records
/orderfood/meals
/orderfood/shopping-lists
/orderfood/ai-providers
/orderfood/ai-models
/orderfood/ai-capabilities
/orderfood/ai-usages
/orderfood/notifications
/orderfood/subscribe-templates
/orderfood/subscribe-logs
```

后台写操作必须分别设计权限，不使用单个“orderfood:write”覆盖全部领域。

## 16. 数据所有权

| 数据 | 小程序用户 | 后台管理员 |
| --- | --- | --- |
| 个人资料 | 可更新头像昵称 | 可看、禁用账号，不代改资料 |
| 私有菜品 | 本人增删改查 | 默认不维护，只在治理授权下查看必要内容 |
| 可发现菜品 | 本人开关发现 | 可精选、关闭发现或治理 |
| 个人菜谱 | 本人增删改查 | 不维护 |
| 饭局 | 创建者按状态操作 | 只读诊断 |
| 点选 | 参与者在收集中修改 | 只读诊断 |
| 采购清单 | 创建者维护 | 只读诊断 |
| 积分 | 查看和业务使用 | 查看、授权角色可人工调整 |
| 能力策略 | 只接收自己的有效状态与动态文案 | 配置平台默认、紧急停用和用户覆盖 |
| AI 配置 | 不可见模型选择 | 配置和审计 |
| 通知 | 查看、标记已读 | 查看投递，不代改用户已读状态 |

## 17. `api.json` 归一化门禁

开始接口实现前，`../miniApp/api.json` 必须通过以下检查：

1. 删除 `MealStatus.expired`，增加 `MealCloseReason`。
2. `DishStatus` 统一为 `draft|usable`。
3. 补充 `IngredientInput`。
4. 补充 `DishStepInput`。
5. 补充 `AuthorSummary`。
6. 补充 `DishCategory`。
7. 补充 `DishTag`。
8. 补充 `MealCandidate`。
9. 补充 `PrepStep`。
10. `Ingredient` 增加单位和备注字段。
11. 登录响应和启动聚合数据包含 `RuntimeConfig`。
12. 小程序接口模块、ID、路径、模型和字段不使用 AI 相关命名。
13. `AiRecommendation`、`AiUsage`、`AiUsageResult` 等客户端模型改为中性命名。
14. 增加平台策略、用户覆盖、积分联动和关闭状态规则。
15. 所有模型引用都能在 `models`、`enums` 或明确基础类型中解析。
16. JSON 解析通过，接口总数和模块总数可重复生成。

未经该门禁，不开始批量生成后端模型或前端 API 封装。

2026-07-24 第二轮校验结果：`api.json` 版本为 `1.2.0`，可解析，保留 13 个模块、63 个接口，包含 36 个模型和 7 个枚举；模型引用未发现未定义项，客户端契约中的 AI 敏感标识扫描结果为 0。

## 18. 契约测试最低范围

- envelope 和错误 envelope。
- 分页请求、总数和稳定排序。
- 小程序公开 ID 字符串。
- 时间和 `null`。
- 菜品封面、标签、步骤校验。
- 推荐副本来源锁。
- 饭局状态流转和截止关闭。
- 并发点选唯一性。
- 并发饭局确认。
- AI 幂等扣分和一次退款。
- 采购分享只读。
- 通知分页和全部已读。
- 后台角色和 API 权限。
- 冷启动微信 `code` 单飞兑换和并发首次登录唯一性。
- 平台默认、用户覆盖、紧急停用和配置版本缓存失效。
- 关闭状态下积分隐藏、打卡零奖励和偏好分析不触发。
- 微信实际上传包敏感字样及路径扫描为 0。

## 19. 变更流程

任何跨端契约变化都必须：

1. 先更新产品规则或本文。
2. 更新 `../miniApp/api.json`。
3. 更新 Swagger 和后端实现。
4. 更新小程序或后台 API 封装。
5. 增加或调整契约测试。
6. 在业务记忆中记录长期口径变化。

不能只改页面字段或数据库字段而不更新契约。
