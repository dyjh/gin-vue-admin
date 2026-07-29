# 来干饭微信小程序

本目录是可直接导入微信开发者工具的原生小程序工程，覆盖 `aiDoc/miniApp/task.json` 中确认的页面。所有业务请求均通过 `services/` 连接真实后端，不包含本地 Mock 路由或数据开关。

## 运行

1. 在微信开发者工具中选择“导入项目”。
2. 项目目录选择本 `miniApp/` 目录。
3. 当前 `project.config.json` 已配置项目小程序 AppID；不同环境使用各自有权限的 AppID，不再使用 `touristappid`。
4. 在 `config/env.js` 中将 `baseUrl` 配置为实际 HTTPS API 地址；小程序只连接真实服务，不再提供 Mock 开关。
5. `/assets/images/` 下的图片统一从 `https://cache.ljdyjh.cn/assets/images/` 读取，不再回退到本地文件：
   - 图片云端路径与页面使用的 `/assets/images/...` 路径保持一致
   - `assets/images` 已从小程序上传包排除，本地文件仅作为开发源图备份
   - 图标仍按 `config/remote-assets.js` 清单决定是否使用云端资源
   - 公共域名统一维护在 `config/domains.js`

在微信公众平台将 `https://cache.ljdyjh.cn` 添加到下载文件合法域名。原生底部 `tabBar` 图标按微信限制继续使用本地文件。

## 目录

- `pages/`：23 个已注册业务页面。
- `components/`：统一头图、图标、菜品行、空状态和菜品表单。
- `services/`：请求、上传审核与 API 封装。
- `assets/`：原生底部导航资源、图标，以及不参与打包的图片源文件备份。
任务规划、设计稿和接口文档统一存放在 `aiDoc/miniApp/`，不会打入小程序运行包。

## 本地校验

```powershell
node miniApp/tools/validate-miniapp.cjs
node aiDoc/miniApp/tools/validate-api.cjs
node miniApp/tools/validate-contract-usage.cjs
node miniApp/tools/test-auth-flow.cjs
```

`validate-api.cjs` 校验机器契约本身，`validate-contract-usage.cjs` 保证 67 个契约接口都存在客户端实现且没有遗留路径；`validate-miniapp.cjs` 检查路由文件、JSON、JavaScript、WXML 标签、事件处理器、素材引用、认证实现和接口清单；`test-auth-flow.cjs` 验证冷启动单飞登录、并发 `40101` 重登、幂等重放、公开请求免登录及上传不自动重放。

## 微信登录触发规则

- 冷启动：`App.onLaunch` 立即发起一次 `wx.login`，并把后端换取业务令牌的过程保存为全局共享 Promise。
- 页面请求：需要登录的请求统一等待该 Promise；页面的 `onLoad`、`onShow` 不各自调用 `wx.login`。
- 令牌过期：收到 `40101` 后通过单飞锁重新登录；并发失败请求只触发一次新的 `wx.login`。
- 手动重试：启动登录失败后，只在用户明确点击重试时再次登录。
- 公开分享：声明 `auth:false` 的公开采购清单请求不等待、也不触发登录。
- 自动重放：`GET` 可重放一次；`PUT`、`DELETE`、`POST` 必须沿用原 `X-Idempotency-Key` 才可重放。图片上传和微信一次性 code 兑换绝不自动重放。
- 安全存储：只持久化业务访问令牌、过期时间、用户摘要和运行时配置；微信一次性 `code` 不写缓存或日志。

## 接口约定

- 基础路径：`/api/miniapp/v1`
- 统一响应：`{ code, data, msg }`
- 分页：`{ page, pageSize, total, list }`，默认每页 20 条
- 变更类请求使用 `X-Idempotency-Key`
- 图片先调用 `/uploads/images` 完成上传与同步内容审核

完整字段和 67 个接口见 [api.json](../aiDoc/miniApp/api.json)。

## 饭局状态

- `collecting`：正在点菜；发起人可以提前关闭或取消。
- `closed`：只关闭点菜，饭局仍在进行中；发起人可以确认菜单或取消。
- `confirmed`：菜单已确认，进入采购、备菜和就餐阶段，仍占用唯一进行中饭局名额。
- `completed`：发起人手动结束整场饭局后进入；此时才能创建下一场。
- `cancelled`：仅允许在确认菜单前进入，进入后可以创建下一场。
