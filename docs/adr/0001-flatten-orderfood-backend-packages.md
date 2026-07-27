# ADR-0001: Flatten the redundant OrderFood backend package layer

## Status
Accepted

## Context

This repository is itself the OrderFood application, but business code is nested again under `model/orderfood`, `service/orderfood`, `api/v1/orderfood`, and `router/orderfood`. The extra namespace adds wrapper groups and import aliases without separating deployable systems. Several services also share transaction, permission, audit, idempotency, and moderation helpers, so forcing every feature into an independent Go package would create cycles or expose implementation-only symbols.

The refactor must preserve existing HTTP paths, permissions, database tables, initialization behavior, Swagger contracts, and the user's uncommitted changes. Existing databases must not be connected to or modified automatically.

## Decision

Adopt a modular monolith with flattened layer packages and feature-named files:

- Move core business models into `server/model`; move DTOs into `server/model/request` and `server/model/response`.
- Move business services into `server/service` and merge the business service aggregate into the root service aggregate.
- Move admin API handlers into `server/api/v1` and merge the business API aggregate into the root v1 aggregate.
- Move admin routers into `server/router` and merge the business router aggregate into the root router aggregate.
- Keep functional file boundaries such as user, points, catalog, recommendation, governance, meals, AI, subscription, and WeChat.
- Preserve Gin's `/orderfood` route prefix, the external `/api/orderfood` contract, permission identifiers, model table names, and seeded API paths.

## Consequences

### Positive

- Removes redundant `orderfood` package paths and `OrderFood*Group` wrappers.
- Keeps shared transactions and internal helpers in one layer package, avoiding artificial cycles.
- Makes feature handlers and services directly discoverable from each layer root.
- Requires no existing-database API route migration while route paths remain unchanged.

### Negative

- Root layer packages contain more files.
- The migration changes many imports and package declarations at once.
- Compile-time feature isolation is not introduced; boundaries remain conventions enforced by filenames, aggregates, and tests.

### Neutral

- Swagger tags may continue to use `OrderFood*` because they are external documentation names, not Go package namespaces.
- Database model table names and configuration keys retain their existing compatibility names.

## Alternatives Considered

1. Keep the current nested packages and only split large files: rejected because it preserves the redundant namespace the project owner asked to remove.
2. Create a Go package per feature in every layer: rejected for this migration because existing cross-feature transactions and unexported helpers would create cycles and broad API exposure.
3. Move to vertical slices under `server/modules`: rejected because it would fight the current Gin/GVA bootstrapping conventions and expand the change beyond the requested cleanup.

## Risks and Mitigations

- Missing imports or stale package declarations: run `gofmt`, `go test ./... -run '^$'`, and focused router tests.
- Nil aggregate services after merging entry points: retain constructor coverage and initialize through one root `ServiceGroupApp`.
- Route drift: keep route literals unchanged and run route contract tests plus admin implementation validation.
- Existing database drift: generate SQL only if seeded API method/path values change; otherwise explicitly document that no SQL is required.