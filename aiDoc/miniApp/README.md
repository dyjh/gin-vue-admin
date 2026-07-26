# 小程序文档索引

本目录保存“来干饭”微信小程序的设计归档、接口契约和阶段交接文档，不属于小程序运行包。

## 当前阶段

- 产品需求与核心业务口径已确认。
- 20 组 UI 设计稿已逐页确认并归档。
- `miniApp/` 已完成原生微信小程序页面、统一组件、真实接口接入和主要交互，Mock 路由及数据已移除。
- 历史饭局页面在设计归档阶段后直接补充到运行代码。
- 小程序认证客户端已完成冷启动共享 Promise、`40101` 单飞重登、安全重放和公开请求免登录；后端已接入微信 code 兑换和业务会话。
- 正式客户端和 `/api/miniapp/v1` 后端已覆盖 67 个契约接口，历史接口路径和旧 DTO 字段已清零。
- 饭局已冻结“关闭点餐”和“结束饭局”两个独立动作；`confirmed` 在采购、备菜和就餐期间仍属于进行中状态。
- 管理端菜单、页面和接口已按冻结契约实现；当前阶段进入 MySQL 联调与发布前回归。

## 文件说明

| 路径 | 用途 | 状态 |
| --- | --- | --- |
| `task.json` | UI 设计任务、页面范围、风格约束和历史确认记录 | 已归档 |
| `Design/` | 最终设计稿，包含 20 个 HTML 和 57 张 PNG | 已归档 |
| `api.json` | 小程序 API 契约 1.9.0，13 个模块、67 个接口 | 后端实现基准 |
| `tools/build-api.cjs` | API 清单生成工具 | 按需使用 |
| `tools/validate-implementation.cjs` | 比对契约与 Go Swagger 的接口签名和同名 DTO 字段 | 回归校验 |
| `backend-integration-plan.md` | 后端、后台 Web、联调顺序和完成标准 | 实现与回归清单 |

配套基线：

- `../prd/order-food-admin-web-prd.md`
- `../modules/order-food-package-architecture.md`
- `../frontend-backend/order-food-contract-baseline.md`
- `../frontend-backend/miniapp-auth-capability-compliance.md`

## 信息优先级

发生描述冲突时，按以下顺序处理：

1. `../prd/family-menu-miniapp-prd.md` 中的现行产品口径。
2. `../frontend-backend/miniapp-auth-capability-compliance.md` 中的微信身份、能力策略、积分联动和发布包规则。
3. `../frontend-backend/order-food-contract-baseline.md` 中的跨端实现约束。
4. `api.json` 中的接口字段、状态枚举和错误码。
5. `../../miniApp/` 中已确认的页面行为与交互。
6. `../prd/order-food-admin-web-prd.md` 中的后台范围。
7. `Design/` 中的视觉布局和状态表现。
8. `task.json` 中的历史讨论和过程记录。

`task.json` 主要用于追溯设计决策，不应作为下一阶段的唯一需求来源。

## 资源约定

- 已稳定的公共资源通过 `https://cache.ljdyjh.cn/assets` 提供。
- 新增或仍在调试的资源可以暂存于 `miniApp/assets/`，确认后再上传公共资源域名。
- 本地资源与远程资源可以并存，不进行全局二选一切换。
- 微信分享封面暂存于本地资源目录，上传 CDN 后再更新分享配置。

## 下一步

按 [后端实现与联调计划](backend-integration-plan.md) 执行 MySQL 环境迁移、端到端联调、微信订阅消息实机验证和发布包扫描。
