# OrderFood Backend Package Refactor Implementation Plan

> **Execution note:** Work in the current shared worktree; do not commit or discard unrelated user changes.

**Goal:** Remove redundant backend `orderfood` package layers and expose the same application through function-oriented files in the existing layer roots.

**Architecture:** Flatten business code into the root model, service, v1 API, and router packages. Merge aggregate entry points while preserving HTTP routes, permission identifiers, database models, and initialization order.

**Tech Stack:** Go, Gin, GORM, GVA service/API/router aggregates, MySQL-compatible seed SQL, Swagger tooling.

---

### Task 1: Capture invariants and migration map

**Files:**
- Read: `server/api/v1/orderfood/*.go`
- Read: `server/router/orderfood/*.go`
- Read: `server/service/orderfood/*.go`
- Read: `server/model/orderfood/**/*.go`
- Test: `server/router/orderfood/*_test.go`

1. Record every source and destination path.
2. Confirm no destination filename conflicts.
3. Confirm Gin route literals and seeded API paths remain `/orderfood/...`.
4. Save the current dirty worktree status without modifying unrelated files.

### Task 2: Flatten models and DTOs

**Files:**
- Move: `server/model/orderfood/*.go` → `server/model/*.go`
- Move: `server/model/orderfood/request/*.go` → `server/model/request/*.go`
- Move: `server/model/orderfood/response/*.go` → `server/model/response/*.go`
- Modify: all Go imports referencing the old model paths

1. Move tracked files without rewriting their contents.
2. Change core model package declarations from `orderfood` to `model`.
3. Update request/response imports to the new root model path.
4. Run `gofmt` on moved model files.
5. Compile the model and dependent front packages without running database tests.

### Task 3: Flatten services and merge the service aggregate

**Files:**
- Move: `server/service/orderfood/*.go` except `enter.go` → `server/service/*.go`
- Modify: `server/service/enter.go`
- Delete after merge: `server/service/orderfood/enter.go`
- Modify: all Go imports referencing `server/service/orderfood`

1. Move service and test files and change package declarations to `service`.
2. Merge `ServiceGroup`, `NewServiceGroup`, and business fields into the root service entry point.
3. Preserve system/example service groups and construct business services in the same dependency order.
4. Update front services, initialization, timers, APIs, and tests to import the root service package.
5. Run compile-only tests for `server/service` and `server/front/...`.

### Task 4: Flatten v1 APIs and merge the API aggregate

**Files:**
- Move: `server/api/v1/orderfood/*.go` except `enter.go` → `server/api/v1/*.go`
- Modify: `server/api/v1/enter.go`
- Delete after merge: `server/api/v1/orderfood/enter.go`
- Modify: API imports for model request/response and root service

1. Move handler and test files and change package declarations to `v1`.
2. Merge business API fields, constructors, permission helper, and binding helper into root `ApiGroup`.
3. Preserve system/example API groups.
4. Compile the v1 API package and its tests without database execution.

### Task 5: Flatten routers and merge the router aggregate

**Files:**
- Move: `server/router/orderfood/*.go` except `enter.go` → `server/router/*.go`
- Modify: `server/router/enter.go`
- Delete after merge: `server/router/orderfood/enter.go`
- Modify: `server/initialize/router_biz.go`

1. Move router and route-test files and change package declarations to `router`.
2. Merge business router structs into the root `RouterGroup`.
3. Replace API imports with the root v1 package.
4. Register business routers directly from `router.RouterGroupApp`.
5. Run route contract tests and confirm every path is unchanged.

### Task 6: Update initialization and external references

**Files:**
- Modify: `server/initialize/gorm_biz.go`
- Modify: `server/initialize/orderfood_admin_seed.go`
- Modify: `server/initialize/orderfood_admin_seed_test.go`
- Modify: `server/initialize/timer.go`
- Modify: `server/front/**/*.go`

1. Replace old model/service aliases with imports from the flattened packages.
2. Preserve AutoMigrate model order and default initialization order.
3. Keep seeded API method/path pairs unchanged.
4. Search the repository for stale `/orderfood` package imports and obsolete aggregate fields.

### Task 7: Format and verify

**Files:** all moved and modified Go files

1. Run `gofmt` over changed Go files.
2. Run `go test ./... -run '^$'` for compile-only validation without database access.
3. Run focused router tests that do not require MySQL.
4. Run admin API contract and implementation validators.
5. Run `git diff --check` and verify no unrelated changes were overwritten.

### Task 8: Decide existing-database SQL requirement

**Files:**
- Inspect: `server/initialize/orderfood_admin_seed.go`
- Create only if needed: `waitSql/20260727_refactor_orderfood_api_routes.sql`

1. Compare every seeded HTTP method/path before and after the refactor.
2. If paths are unchanged, do not create SQL and document that the refactor is internal only.
3. If any method/path must change, create idempotent pure SQL in `waitSql`; never connect to the existing database.