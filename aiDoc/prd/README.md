# 产品文档索引

本目录保存“来干饭”微信小程序的现行产品需求、功能结构和关键业务流程。

## 主文档

- [家庭点餐微信小程序 PRD](family-menu-miniapp-prd.md)：现行产品口径、范围、模型、状态、接口要求和验收标准。
- [后台 Web 管理端 PRD](order-food-admin-web-prd.md)：后台角色、菜单、运营、内容安全、AI、问题排查和验收边界。
- [后台管理端菜单与页面功能实施基线](order-food-admin-menu-page-review.md)：6 个菜单分组、数据概览与微信配置 2 个一级直达页面、共 28 个页面，以及对应按钮、功能、权限和风险交互。
- [功能点思维导图](family-menu-miniapp-feature-mindmap.md)：第一版功能边界与模块关系。

## 核心流程

- [菜品与菜谱创建流程](dish-and-recipe-creation-flowchart.md)：手动录入、AI 识别、菜品保存和菜谱维护。
- [饭局点餐到采购流程](meal-order-to-purchase-flowchart.md)：创建饭局、点餐、确认菜单、快照与采购清单。

同名 PNG 为流程图渲染结果，Markdown 为可维护源文件。2026-07-24 已更新两份流程图 Markdown 的状态和封面口径；当前工作区未安装 Mermaid CLI，PNG 需在下次具备渲染环境时重新生成，实施时以 Markdown 源文件为准。

## 维护规则

- 产品口径变化优先更新主 PRD，再同步流程图、`../miniApp/api.json` 和业务记忆。
- 设计任务的历史讨论不能覆盖主 PRD 的现行口径。
- 新增第一版范围前，必须明确数据模型、权限、状态流转和验收标准。
