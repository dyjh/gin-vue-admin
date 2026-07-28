# Fixed Subscription Scenes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use the Code workflow to implement this plan task-by-task in the current workspace.

**Goal:** Replace free-form subscription-template CRUD with a fixed business-scene catalog where each supported scene has at most one current WeChat template binding.

**Architecture:** Business scenes are defined in Go and cannot be created or deleted by administrators. The existing `of_sub_templates` table stores the optional current binding for a scene, with `scene` uniquely constrained; the API exposes scene-oriented read, configure, and status endpoints. The first release exposes only `meal_status`, whose business fields are exactly `mealName`, `result`, and `resultAt`.

**Tech Stack:** Go, Gin, GORM, MySQL, Vue 3, Element Plus, Vite, Swagger, generated admin API contract.

---

### Task 1: Lock the fixed-scene behavior with service tests

**Files:**
- Modify: `server/service/subscription_test.go`
- Modify: `server/model/request/subscription.go`
- Modify: `server/model/response/subscription.go`

**Steps:**
1. Replace the free-form create/delete flow test with an unconfigured-scene → configure → enable → log-detail flow.
2. Assert that the catalog always returns `meal_status`, even before a database binding exists.
3. Assert that first configuration creates a disabled binding and later configuration updates the same row.
4. Assert that unknown scenes, missing required mappings, extra business fields, duplicate WeChat fields, and stale versions fail.
5. Run:
   `go test ./service -run 'TestSubscribeScene|TestSubscriptionAdmin' -count=1`
   and confirm the new tests fail before implementation.

### Task 2: Implement the fixed scene registry and single binding

**Files:**
- Modify: `server/model/subscription.go`
- Modify: `server/service/subscription.go`

**Steps:**
1. Add a unique index for `SubscribeMessageTemplate.Scene`.
2. Define the immutable `meal_status` scene metadata and required field definitions in the service.
3. Add `ListSubscribeScenes`, `GetSubscribeScene`, `ConfigureSubscribeScene`, and `UpdateSubscribeSceneStatus`.
4. Make configure perform an idempotent upsert: first configuration creates a disabled row; later configuration requires the current version and preserves enabled state.
5. Require field mappings to contain exactly the fixed business-field keys.
6. Keep send logs linked to the persistent binding ID and keep mutation auditing.
7. Remove free-form list/create/update/delete service methods and their request types.
8. Run the focused service tests and confirm they pass.

### Task 3: Replace HTTP routes and permissions

**Files:**
- Modify: `server/api/v1/subscription.go`
- Modify: `server/router/subscription.go`
- Modify: `server/source/orderfood_admin_seed.go`
- Modify: `server/api/v1/ai_test.go` or the applicable route/seed test

**Steps:**
1. Expose:
   - `GET /orderfood/subscribe-scenes`
   - `GET /orderfood/subscribe-scenes/{scene}`
   - `PUT /orderfood/subscribe-scenes/{scene}`
   - `PUT /orderfood/subscribe-scenes/{scene}/status`
2. Reuse read, update, and status permissions; remove create and delete buttons/APIs.
3. Keep subscription-log endpoints unchanged.
4. Update seed assertions and route tests.
5. Run:
   `go test ./api/v1 ./source -count=1`

### Task 4: Rebuild the admin page as scene configuration

**Files:**
- Modify: `web/src/api/orderfood/message.js`
- Modify: `web/src/view/orderFood/subscribeTemplate/index.vue`

**Steps:**
1. Replace paginated template CRUD calls with scene catalog/detail/configure/status calls.
2. Render one fixed scene card containing trigger, recipients, configuration completeness, enabled state, counters, and updated time.
3. Replace “新增模板” with “配置模板” or “更换模板”.
4. Remove search, delete, free-form name/purpose, and “新增扩展字段”.
5. Render the three fixed business fields read-only and only allow editing their WeChat field names.
6. Preserve conflict handling, status reason collection, log navigation, and edit loading/error states.
7. Run `npm run build`.

### Task 5: Synchronize contracts and documentation

**Files:**
- Modify: `aiDoc/admin/tools/build-api.cjs`
- Modify: `aiDoc/admin/tools/validate-api.cjs`
- Regenerate: `aiDoc/admin/api.json`
- Regenerate: `server/docs/docs.go`
- Regenerate: `server/docs/swagger.json`
- Regenerate: `server/docs/swagger.yaml`
- Modify: `aiDoc/frontend-backend/order-food-admin-api-contract.md`
- Modify: relevant PRD and business-memory files

**Steps:**
1. Replace template CRUD contracts with fixed-scene contracts and response models.
2. Remove create/delete permissions from the generated contract and seed expectations.
3. Document that new scenes require a code release with trigger, recipients, authorization entry, payload builder, and tests.
4. Regenerate Swagger and machine contracts.
5. Run both admin contract validators and confirm interface/permission counts match.

### Task 6: End-to-end verification

**Steps:**
1. Run focused Go service/API tests using `/usr/local/btgo/bin/go` with `GOPATH=$HOME/go`.
2. Run the web production build.
3. Confirm Air v1.61.7 and `npm run serve` are healthy in the existing WSL tmux sessions.
4. In the browser verify the unconfigured and configured scene states, fixed field mapping editor, status action, responsive scrolling, and absence of console errors.
5. Run `git diff --check` and review that unrelated `waitSql` files remain untouched.
