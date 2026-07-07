# Front 接口组

`server/front` 是面向前台/小程序端的接口组，和后台 admin API 分开注册、分开鉴权。

## 目录结构

```text
server/front/
  register.go      # front 对外注册入口，只由 initialize/router_biz.go 调用
  api/             # handler 层，负责参数绑定、调用 service、返回 response
  router/          # 路由层，负责把 front 接口挂到 /front 下
  service/         # 业务层，复用 server/model 下的实体和数据访问能力
  request/         # front 专属请求 DTO，不复用后台 request
  response/        # front 专属响应 DTO，用于裁剪或组合返回字段
  docs/            # front 专属 swag 文档产物目录
```

## 注册方式

front 只挂在 `publicGroup` 下：

```go
front.Register(publicGroup)
```

当前测试接口：

```text
GET /front/test
```

在默认 `system.router-prefix: /api` 下，实际访问路径为：

```text
GET /api/front/test
```

## 权限边界

- 不注册到后台 API 权限表。
- 不调用 `middleware.JWTAuth()`。
- 不调用 `middleware.CasbinHandler()`。
- 如前台需要登录态，应在 `server/front` 内新增独立 middleware。

## Swagger

front handler 使用和 `server/api/v1` 相同的 swag 注释风格。生成文档时，`@Router` 路径不包含 `system.router-prefix`，例如：

```go
// @Router /front/test [get]
```
