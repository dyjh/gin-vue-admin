# OrderFood 功能域包重构实施清单

## 1. 基线与边界

- 记录初始工作区，保留用户已有 `miniApp` 修改。
- 使用 WSL Go 运行 `go test ./... -run '^$'`，确认重构前可编译。
- 以 `server/router/system/enter.go` 和 `server/api/v1/system/enter.go` 为依赖绑定基准。

## 2. 后台 Service

- 提取 `service/common` 的权限、幂等、访问审计和共享工具。
- 将业务服务迁入 `user`、`engagement`、`content`、`dish`、`meal`、`ai`、`dashboard`、`audit`。
- 每个域定义 `ServiceGroup`，根组只保留嵌套域字段。
- 统一通过 `database()` 延迟读取测试注入 DB 或 `global.GVA_DB`，不增加运行时数据库绑定器。
- 更新初始化数据、定时任务和 Front 后台服务引用。

## 3. 后台 API

- 提取 `api/v1/common` 请求绑定、权限和管理员上下文辅助。
- 将处理器迁入对应功能域，拆分原 `operations.go`。
- 域 `enter.go` 直接绑定嵌套 ServiceGroup；删除处理器构造注入和旧兼容字段。

## 4. 后台 Router

- 将路由和路由测试迁入对应功能域，拆分原 `operations.go`。
- 域 `enter.go` 直接绑定嵌套 ApiGroup。
- 路由注册方法不接收 API 参数；初始化入口按域 RouterGroup 调用。

## 5. Front Service

- 按依赖边界迁入 `user`、`experience`、`system`。
- 根 `ServiceGroup` 只聚合三个嵌套域组。
- 更新认证中间件和定时任务引用。

## 6. Front API 与 Router

- API 和 Router 迁入 `auth`、`system`、`profile`、`content`、`engagement`、`meal`、`assist`。
- 共享 API 辅助迁入 `api/common`。
- 各域 `enter.go` 分别绑定嵌套 ServiceGroup 和 ApiGroup。
- `front/register.go` 按域 RouterGroup 注册，保持原公开/鉴权边界。

## 7. 验证

从 `server` 目录使用 WSL Go：

```text
/usr/bin/env GOTOOLCHAIN=auto /www/server/go/root/1.23.9/bin/go test ./... -run '^$'
/usr/bin/env GOTOOLCHAIN=auto /www/server/go/root/1.23.9/bin/go test ./...
```

最后执行旧入口搜索、`git diff --check`、工作区状态检查，确认没有兼容层和无关文件修改。
