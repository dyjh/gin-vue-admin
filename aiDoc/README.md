# aiDoc

`aiDoc/` 是本仓库的结构化 AI 文档层，用于把长期有效的项目上下文从工具目录中抽离出来，并按主题拆分成可维护的约束文档。

## 使用方式

1. 先读取 `AGENT.MD`
2. 再查看本索引文件
3. 按任务只打开相关子目录
4. 不再把项目级规则塞回 `.codex/`、`.claude/`、`.cursor/`、`.trae/`

## 目录说明

- `relations/`: 仓库结构、技术栈、依赖关系、开发流程
- `modules/`: 后端分层规则、插件结构、模块职责
- `frontend-backend/`: 前后端契约、前端规范、工具函数复用规则
- `examples/`: 讲解型示例，告诉 AI 每一层应该按什么标准组织和书写
- `memory/`: AI 记忆层，拆分为长期记忆与业务记忆
- `prd/`: 现行产品需求、功能结构和关键业务流程
- `miniApp/`: 小程序设计归档、API 契约和阶段交接资料

## 常用入口

- `relations/repo-profile.md`: 项目定位、技术栈、核心特性
- `relations/development-workflow.md`: GVA 开发流程、分支与提交规范
- `relations/licensing-and-branding.md`: 版权、授权与品牌相关任务的协作边界
- `modules/backend-layer-rules.md`: 后端分层、`enter.go`、统一响应、Swagger 约束
- `modules/plugin-development.md`: 前后端插件结构与开发流程
- `modules/order-food-package-architecture.md`: “来干饭” `orderfood` 业务包、领域边界与 GVA MCP 生成规范
- `frontend-backend/boundary.md`: 前后端契约与字段类型约束
- `frontend-backend/frontend-rules.md`: 前端代码、状态、路由、样式规范
- `frontend-backend/frontend-utils.md`: `src/utils/` 工具库的强制复用规则
- `frontend-backend/order-food-contract-baseline.md`: “来干饭”小程序、后端和后台 Web 的跨端契约基线
- `frontend-backend/miniapp-auth-capability-compliance.md`: 微信身份初始化、能力策略、积分联动和小程序发布合规基线
- `examples/README.md`: 示例层总入口
- `memory/project-memory.md`: 记忆层总入口
- `memory/long-term/`: 长期记忆
- `memory/business/`: 业务需求记忆
- `prd/README.md`: 产品文档总入口
- `prd/family-menu-miniapp-prd.md`: “来干饭”现行产品口径
- `prd/order-food-admin-web-prd.md`: “来干饭”后台 Web 产品范围、角色与验收标准
- `miniApp/README.md`: 小程序设计、接口与阶段状态总入口
- `miniApp/backend-integration-plan.md`: 下一阶段后端、后台 Web 实现与联调计划

## 维护原则

- 稳定规则放这里，不放到工具私有目录里
- 临时会话草稿不要入库，只有变成长期知识时才记录
- 适用于所有 AI 的项目级规则，先写进 `AGENT.MD`
- 细节说明再拆到 `aiDoc/` 对应子目录
- 只要用户提出业务需求，就要同步更新 `memory/business/`
