# Probo — Go Backend — pkg/coredata

**Purpose.** Sole owner of SQL in the codebase. One file per entity
(`vendor.go`, `cookie_banner.go`, ...) plus a paired
`<entity>_filter.go` and `<entity>_order_field.go`. Houses the
`Scoper` interface (tenant isolation), 300+ migrations under
`pkg/coredata/migrations/`, and the entity-type registry that backs
`pkg/gid`. Implements `agent.Checkpointer` via `PGCheckpointer`.

**Key files.**

- `scope.go` — `Scoper` interface (`SQLFragment`, `SQLArguments`,
  `GetTenantID`); `Scope` (multi-tenant) and `NoScope` (cross-tenant
  admin, panics on `GetTenantID`).
- `entity_type_reg.go` — sequential `uint16` constants and
  `NewEntityFromID`; never reuse a removed number (use `_` placeholder).
- `errors.go` — `ErrResourceNotFound`, `ErrResourceAlreadyExists`,
  `ErrResourceInUse`, `ErrNoDocumentPDFJobAvailable`.
- `cookie_banner.go` — canonical entity (full CRUD + cursor + filter +
  worker query).
- `agent_run.go` — `PGCheckpointer`; **deviation**: line 472 hardcodes
  `'PENDING'` SQL literal (drift to fix).
- `migrations.go` — `embed.FS` of all migration SQL.

**How to extend (add a new entity).**

1. Create `pkg/coredata/<entity>.go` — struct with `db` tags, `gid.GID`
   ID, `OrganizationID`, timestamps, slice alias `<Entity>s`,
   `AuthorizationAttributes()`, `CursorKey()`, CRUD methods.
2. Create `<entity>_filter.go` and `<entity>_order_field.go`.
3. Append a sequential constant to `entity_type_reg.go`; add a `case`
   in `NewEntityFromID`.
4. Add a migration file under `migrations/` named
   `YYYYMMDDTHHMMSSZ.sql` with **random 6-digit time portion** to avoid
   collisions (see [shared.md § 5 memory note](../../shared.md#5-git--workflow)).
5. Run `make test` to confirm migrations apply on a fresh DB.

**Top pitfalls.**

- Forgetting to declare every key in `pgx.StrictNamedArgs` (filter args
  must be present even when nil) — see
  [pitfalls.md](../pitfalls.md).
- Calling `scope.GetTenantID()` on `NoScope` — runtime panic.
- Reusing a removed entity-type number — corrupts every existing GID
  for that type.
