# 小程序头图统一重绘

## 基本信息

- 提出日期：2026-07-29
- 当前状态：`active`
- 需求类型：小程序 UI / 品牌插画资源
- 优先级：高
- 需求文件：`aiDoc/memory/business/active/miniapp-hero-redraw.md`

## 用户原始意图摘要

除首页外，将所有带头图的小程序页面按照已锁定的猫咪形象、画面风格和色彩规范统一重绘。前几个页面逐页对比确认风格稳定后，用户已授权将其余页面一次性批量完成并接入。

## 影响范围

- 后端：无
- 前端：`miniApp/` 中使用头图的页面与本地图片资源
- 文档：`aiDoc/memory/` 中的长期视觉规范和本业务记忆
- 插件 / 模块：无

## 涉及对象

- 模块：微信小程序品牌视觉
- 接口：无
- 页面：除首页外所有实际使用头图的页面
- 配置：本地资源引用；已稳定资源后续可按现有流程上传公共资源域名

## 已确认约束

- 唯一风格基准为 `miniApp/assets/images/home-approved-header-v3.jpg`。
- 详细规范见 `aiDoc/memory/long-term/miniapp-illustration-visual-style.md`。
- 前几个页面逐页展示、确认；风格确认后，其余页面按用户最新要求一次性批量生成、接入和验收。
- 每个业务场景必须使用自然且不同的动作与道具，不机械重复方巾、小番茄和迷迭香。
- 预览和对比资源可缓存于用户指定的 `D:\www\aiCache`，未确认稿不进入项目运行资源。
- 接入资源采用 `1500 × 804` 横幅，JPG 小于 1 MB，并为状态栏、返回按钮、标题和微信胶囊预留安全区。
- 添加菜品与编辑菜品属于同一菜品维护流程，确认共用同一张“记录菜谱”头图。

## 当前进展

- 已盘点除首页外的现有头图页面和共享资源关系。
- “添加菜品 / 编辑菜品”共用的记录菜谱头图已完成接入，并通过用户实际页面验收；项目资源为 `miniApp/assets/images/dish-form-cat-writing-v1.jpg`。
- “菜品详情”头图已完成三轮调整：移除按在书本上的勺子并改为翻页动作，再将左下角桌布、小番茄组合替换为低矮香草盆栽。
- 菜品详情最终确认稿已生成项目资源 `miniApp/assets/images/dish-detail-cat-reading-v1.jpg` 并接入页面，微信开发者工具实际页面验收通过。
- “全部菜品 / 个人菜品库”列表页 v2 已确认并接入，项目资源为 `miniApp/assets/images/dish-library-cat-organizing-v1.jpg`；该版移除机械复用的格纹方巾、番茄和迷迭香，左侧保留干净木台面。 微信开发者工具实际页面验收通过。
- “菜品推荐”列表页 v1 已确认并接入，项目资源为 `miniApp/assets/images/recommend-dishes-cat-presenting-v1.jpg`：以双手端出精选家常菜替代举旗动作。 微信开发者工具实际页面验收通过。
- “推荐菜品详情”候选 v1 已接入，项目资源为 `miniApp/assets/images/recommend-dish-detail-cat-reviewing-v1.jpg`；实际页面头图裁切通过，标题调整为“推荐详情”，副标题调整为“平台精选推荐”；不影响仍使用旧图的“今天吃什么结果详情”。

- 其余 16 个页面已按最新批量要求完成 15 张新头图生成与接入；购物清单页和购物分享页因属于同一采购场景共用一张，其余页面均使用独立场景：
  - 今天吃什么：`what-to-eat-cat-choosing-v1.jpg`
  - 今天吃什么结果详情：`what-to-eat-result-cat-reveal-v1.jpg`
  - 菜谱列表：`recipes-cat-shelving-v1.jpg`
  - 菜谱详情：`recipe-detail-cat-album-v1.jpg`
  - 个人中心：`profile-cat-kitchen-cubby-v1.jpg`
  - 打卡：`checkin-cat-photographing-v1.jpg`
  - 创建饭局：`meal-create-cat-setting-table-v1.jpg`
  - 饭局邀请：`meal-invite-cat-numeric-code-v3.jpg`（原长桌与三套餐位，展示六位数字码 `528 316`）
  - 饭局投票：`meal-vote-cat-choosing-v1.jpg`
  - 饭局统计：`meal-stats-cat-portions-v1.jpg`
  - 饭局历史：`meal-history-cat-calendar-v1.jpg`
  - 购物清单 / 分享：`shopping-list-cat-packing-v1.jpg`
  - 备菜指引：`meal-prep-cat-sequencing-v1.jpg`
  - 积分：`points-cat-saving-tokens-v1.jpg`
  - 消息中心：`notifications-cat-reading-mail-v1.jpg`
- “饭局统计”和“备菜指引”原先共图，现已拆为称量份量与备菜顺序两个独立场景；“推荐详情”和“今天吃什么结果详情”也已拆分，不再共图。
- 批量生成原稿已缓存于用户指定的 `D:\www\aiCache`；项目运行资源统一为 `1500 × 804` JPG，均小于 1 MB。

## 后续待办

- 完成资源尺寸、引用、差异和小程序全量校验。
- 在微信开发者工具中一次性复核剩余页面的标题安全区、胶囊避让和头图裁切。

## 更新规则

- 同一需求始终维护在本文件中。
- 每完成或确认一个页面，只更新“当前进展”和“后续待办”。
- 全部页面完成并回归后，再将本文件移入 `done/`。