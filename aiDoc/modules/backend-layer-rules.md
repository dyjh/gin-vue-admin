# 后端分层约束

## 总原则

- 严格遵守 `Router -> API -> Service -> Model` 依赖方向
- 禁止跨层直接调用
- `enter.go` 作为组装与暴露入口，避免循环引用

## Model 层

- 数据模型优先继承 `global.GVA_MODEL`
- 持久化字段应补全清晰的 `json` 与 `gorm` 标签；`gorm` 至少明确 `column`、数据库 `type`、空值或默认值、必要索引和 `comment`
- `ID`、`CreatedAt`、`UpdatedAt` 这些基础字段沿用项目现有约定
- 请求模型放在 `model/request/`
- 列表查询模型应定义 `XxxSearch`，并内嵌通用的 `request.PageInfo`
- 结构体字段说明写在字段同行，不能用脱离字段的整段说明代替
- 业务表名应使用模块短前缀，`orderfood` 统一使用 `of_` 且表名不超过 30 个字符

## Go 注释

- 导出类型、函数和方法必须有以名称开头的用途注释
- 非导出方法也要在职责、输入输出或副作用不直观时补充方法注释
- 状态流转、事务边界、并发锁、幂等、权限校验和失败补偿必须写逻辑注释，说明设计原因
- 不写“给变量赋值”“调用某方法”这类只复述代码的无效注释

## 类型一致性

- 同一字段在模型、请求结构、响应结构、前端使用处必须保持一致
- 状态字段、ID 字段、枚举字段、时间字段是高风险字段，必须重点检查
- 若涉及指针类型与非指针类型互转，必须在 Service 层显式处理 `nil`

## Service 层

- 只承载业务逻辑，不处理 HTTP 语义
- 不要依赖 `gin.Context`
- 函数应返回业务结果和 `error`
- 每个模块在 `service/` 下建立独立文件，并在 `service/enter.go` 注册

## API 层

- 负责参数提取、参数校验、调用 Service 和统一响应
- 参数从哪里取，取决于前端怎么传、协议怎么设计、当前逻辑需要什么，以及哪个位置更合理
- 不要把绑定方式写死成某一种固定模板

### 常见参数来源

- JSON body
- Query string
- Path params
- `multipart/form-data`
- Header
- Cookie

### 常见取法

- JSON body: `ShouldBindJSON`
- Query: `ShouldBindQuery`、`c.Query(...)`、`c.DefaultQuery(...)`
- Path: `c.Param(...)`
- form-data / file upload: `c.FormFile(...)`、`c.DefaultPostForm(...)`、`c.Request.FormValue(...)`
- Header: `c.GetHeader(...)`、`c.Request.Header.Get(...)`
- Cookie: `c.Cookie(...)`

### 使用原则

- 绑定方式要与真实参数来源一致
- 不要为了套模板，把 Header / Cookie / Query / form-data 中的数据强行改成 body
- 认证、追踪、网关透传等信息，很多时候本来就应该从 Header 或 Cookie 获取
- 上传文件时，应按上传协议从 `multipart/form-data` 中取文件和附带字段
- 前台小程序与公共分享接口在绑定完成后必须调用 `server/utils/front_validator.go` 的 `utils.VerifyAll(&req)`
- 前台请求结构体使用 `validate` 标签；自然语言字段按需使用 `checksql:"false"`，SQL 查询始终使用参数绑定
- 对外业务错误码只在 `server/errors/error.go` 显式定义，业务包中禁止散落裸数字错误码
- Service 返回类型化错误；API 使用 `c.Error(err)` 交给统一错误处理中间件，不在各处理函数重复拼失败响应

- 必须通过 `service.ServiceGroupApp` 访问服务层
- 必须使用项目统一的 `response` 包输出结果
- 每个对外 API 都必须写完整且准确的 Swagger 注释

## Router 层

- 负责路由分组、中间件挂载和处理函数绑定
- 必须通过 `api.ApiGroupApp` 引用 API 层
- 每个模块在 `router/` 下建立独立文件，并在 `router/enter.go` 注册

## Initialize 层

插件或模块若需要初始化入口，至少关注以下职责：

- `gorm.go`: 表结构迁移
- `router.go`: 路由注册
- `menu.go`: 菜单与权限初始化
- `viper.go`: 配置加载
- `api.go`: API 注册

## Swagger 约束

对外 API 的 Swagger 注释至少要准确说明：

- `@Tags` 和 `@Summary`：模块归属与功能说明
- `@Security`：明确 `ApiKeyAuth` 或 `NoAuth`
- `@accept` 和 `@Produce`：真实请求、响应媒体类型
- `@Param`：真实存在的 Path、Query、Header、Body 或 form-data 参数
- `@Success`：统一响应外壳和具体 `data` 数据模型；有业务数据时不得只写裸 `response.Response` 或 `data=object`
- `@Router`：与实际注册一致的完整路由和 HTTP 方法

`orderfood` 模块通过 `server/orderfood_comment_convention_test.go` 检查 API、Service、Request 和 Response 的导出符号注释、请求响应字段同行注释及 Swagger 必填项。修改相关代码后应同时执行 Swagger 生成，确保注解类型可以被解析。

## 数据库测试约束

- `orderfood` 的 GORM 和数据库集成测试以 MySQL 8 为唯一数据库基线，不使用 SQLite 兼容模式。
- 测试数据库由 `server/testutil.OpenMySQL` 创建；业务测试文件不得自行拼接数据库 DSN。
- 测试只读取 `ORDERFOOD_TEST_MYSQL_DSN`，基础数据库名必须以 `_test` 结尾，并为每个测试创建、回收独立数据库。
- 测试不得读取项目运行配置，不得连接开发或生产数据库，不得因 MySQL 未配置而静默跳过或回退 SQLite。
