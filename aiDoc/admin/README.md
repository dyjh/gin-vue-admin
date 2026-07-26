# 来干饭后台管理端契约

本目录保存 `orderfood` 后台管理端的可实施接口契约。

## 文件

- `api.json`：管理端 API 的机器可读唯一真源，由 `tools/build-api.cjs` 生成。
- `tools/build-api.cjs`：接口清单、模型、页面和权限码的生成源。
- `orderfood-menu-wechat-current-db.sql`：已有 MySQL 数据库移除旧菜单包装层、替换仪表盘并新增微信配置的直接调整脚本。
- `../prd/order-food-admin-menu-page-review.md`：6 个菜单分组、2 个一级直达页面、29 个页面的菜单顺序、页面按钮和角色可见性实施基线。
- `../frontend-backend/order-food-admin-api-contract.md`：认证、分页、幂等、错误语义、指标口径、角色边界和关键工作流。

错误码常量的代码唯一真源是 `server/errors/error.go`。本目录只镜像管理端 `20000-29999` 契约，不在业务模块内另建错误码表。

## 使用顺序

1. 先阅读 `../prd/order-food-admin-web-prd.md` 和菜单页面实施基线。
2. 再阅读 `../frontend-backend/order-food-admin-api-contract.md`。
3. 后端、Swagger、Casbin API 和 Web API 封装逐条对照 `api.json`。
4. 接口变更先修改契约和生成源，再修改实现。

## 生成与校验

```bash
node aiDoc/admin/tools/build-api.cjs
node aiDoc/admin/tools/validate-api.cjs
node aiDoc/admin/tools/validate-implementation.cjs
```

生成结果必须满足：

- 接口 ID 唯一。
- `method + path` 唯一。
- 所有写接口声明按钮权限和幂等策略。
- 所有页面引用的接口都真实存在。
- 页面固定为 6 个菜单分组、2 个一级直达页面、29 个叶子页面，页码、路由名和排序唯一。
- 所有接口使用 `/api/orderfood` 管理员认证边界。
- 29 个页面组件、114 条 Swagger 接口和 114 个 Web API 封装与契约完全一致。
- 87 个业务权限均进入 GVA 按钮权限真源，且没有契约外权限。

当前生成统计：29 个页面、114 条接口、141 个模型、87 个权限码、4 类默认角色模板。
