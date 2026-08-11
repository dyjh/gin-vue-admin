# OrderFood Model Functional Packages Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 OrderFood 业务模型从 `server/model` 的扁平目录迁移到按功能划分的包，并把各业务域的 `request`、`response` 收进对应目录。

**Architecture:** 以业务能力而不是接口端区分模型。持久化实体归入 `ai`、`audit`、`content`、`dish`、`engagement`、`meal`、`user`；共享标识、分页、操作者和幂等 DTO 归入 `common`；没有持久化实体的运营看板 DTO 归入 `dashboard`。迁移后直接引用新包，不提供旧路径别名或兼容转发。

**Tech Stack:** Go 1.23、GORM、Gin、Go AST、gofmt、Go test

---

## Package ownership

| Package | Entity ownership | Request/response ownership |
| --- | --- | --- |
| `model/ai` | AI 提供商、模型、能力、提示词、配置变更、AI 功能用量 | AI 配置与用量管理 |
| `model/audit` | 后台审计日志 | 审计日志查询 |
| `model/common` | ID、后台访问审计、后台/前台幂等记录 | 分页、审计分页、当前管理员、用户/目录摘要、幂等头 |
| `model/content` | 用户菜品/食谱、媒体、内容治理、内容审核 | 内容、治理、媒体与审核管理 |
| `model/dashboard` | 无持久化实体 | 运营看板查询与结果 |
| `model/dish` | 分类、标签、单位、官方菜品、推荐、标准菜品索引 | 目录、官方菜品、推荐与建议目录管理 |
| `model/engagement` | 活跃、签到、积分、通知、订阅消息 | 积分与订阅管理 |
| `model/meal` | 饭局、参与者、候选菜、投票、购物清单、备餐计划 | 饭局与购物清单管理 |
| `model/user` | 小程序用户、会话、偏好证据、微信配置 | 用户与微信配置管理 |

## Migration rules

1. 同一文件只保留一个紧密相关的实体聚合；字段枚举及其常量与所属实体放在一起，把当前混合文件中的无关模型拆开。
2. 每个业务包按需建立 `request` 和 `response` 子包，命名与 `model/system` 保持一致。
3. `operations.go` 拆成 `dashboard` 和 `ai` 两套 DTO；`meal_admin.go` 拆成饭局和购物清单 DTO。
4. 所有 API、router、service、front、initialize、source 和测试直接导入新路径。
5. 删除旧的 `model/*.go`、`model/request/*.go`、`model/response/*.go` 业务文件，不保留兼容层。
6. 保留 `model/system`、`model/example` 和框架已有的 `model/common` 基础设施。
7. 扩展模型约束测试，使其递归校验新的业务实体包，并排除 DTO 子包。

## Verification

1. 对所有迁移文件执行 `gofmt`。
2. 使用 WSL Go 执行 `go test ./...`。
3. 执行路由与前台重点测试，确认 URL 和 JSON 契约未变化。
4. 执行 `git diff --check`，检查无残留旧导入路径。

## Repository constraints

- 不改动 `miniApp` 当前未提交工作。
- 不创建兼容别名或转发文件。
- 不自动提交或暂存代码。
