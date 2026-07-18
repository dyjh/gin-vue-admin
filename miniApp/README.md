# 来干饭微信小程序

本目录已经生成可导入微信开发者工具的原生小程序工程，覆盖 `task.json` 中确认的 20 个页面。页面默认走本地 Mock，方便在后端接口完成前先检查布局和交互。

## 运行

1. 在微信开发者工具中选择“导入项目”。
2. 项目目录选择本 `miniApp/` 目录。
3. 当前 `project.config.json` 使用 `touristappid`；联调前替换为实际小程序 AppID。
4. 需要连接真实服务时，编辑 `config/env.js`：
   - `useMock: false`
   - `baseUrl` 改为实际 HTTPS API 地址

## 目录

- `pages/`：20 个业务页面。
- `components/`：统一头图、图标、菜品行、空状态和菜品表单。
- `services/`：请求、上传审核与 API 封装。
- `mock/`：可跨页面保持状态的本地演示数据。
- `assets/`：已按小程序尺寸压缩的定稿插画、菜品照片和统一图标。
- `api.json`：后端接口清单与数据模型。
- `.preview/`、`design/`：设计阶段资料，不会打入运行包。

## 本地校验

```powershell
node miniApp/tools/validate-miniapp.cjs
node miniApp/tools/test-mock-flow.cjs
```

`validate-miniapp.cjs` 会检查路由文件、JSON、JavaScript、WXML 标签、事件处理器、素材引用和接口清单；`test-mock-flow.cjs` 会串行验证菜品、推荐、菜谱、打卡积分、饭局、采购、AI 和通知流程。

## 接口约定

- 基础路径：`/api/miniapp/v1`
- 统一响应：`{ code, data, msg }`
- 分页：`{ page, pageSize, total, list }`，默认每页 20 条
- 变更类请求使用 `X-Idempotency-Key`
- 图片先调用 `/uploads/images` 完成上传与同步内容审核

完整字段和 60 个接口见 [api.json](./api.json)。
