# Probo -- Go Backend -- pkg/coredata

> Module-specific patterns that differ from stack-wide conventions.
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md).

## Purpose

All raw SQL queries, entity struct definitions, filters, order fields, enumerations, and database migrations for every domain object. No ORM -- all SQL is hand-written inline using pgx named arguments. This is the **single source of truth** for database interaction.

## Entity File Pattern

One file per entity in snake_case (e.g., `asset.go`, `vendor_risk_assessment.go`). Companion files use suffixes: `_filter.go`, `_order_field.go`, `_type.go`, `_state.go`, `_status.go`.

Every entity struct uses:
- `gid.GID` for primary key
- `db` struct tags for pgx column mapping
- `time.Time` for CreatedAt/UpdatedAt
- Pointer types for nullable columns
- **No TenantID field** (injected by Scoper at query time)

```go
// See: pkg/coredata/asset.go
type Asset struct {
    ID              gid.GID   `db:"id"`
    SnapshotID      *gid.GID  `db:"snapshot_id"`
    Name            string    `db:"name"`
    OwnerID         gid.GID   `db:"owner_profile_id"`
    OrganizationID  gid.GID   `db:"organization_id"`
    CreatedAt       time.Time `db:"created_at"`
    UpdatedAt       time.Time `db:"updated_at"`
}
```

## Method Signatures

| Method | Receiver | Returns | Notes |
|--------|----------|---------|-------|
| `LoadByID` | `*Entity` | `error` | Assigns into receiver via `*e = entity` |
| `LoadAllBy*` | `*Entities` (slice) | `error` | Paginated with `page.Cursor[OrderField]` |
| `CountBy*` | `*Entities` | `(int, error)` | Uses `COUNT(id)` |
| `Insert` | `*Entity` | `error` | Uses `scope.GetTenantID()` for tenant_id |
| `Update` | `*Entity` | `error` | Uses `RETURNING` and reassigns receiver |
| `Delete` | `*Entity` | `error` | Checks `snapshot_id IS NULL` to prevent deleting snapshots |
| `CursorKey` | `*Entity` | `page.CursorKey` | Switch on OrderField; panics on unknown |
| `AuthorizationAttributes` | `*Entity` | `(map[string]string, error)` | Returns org/tenant IDs for ABAC |

## Row Collection

```go
// Single row -- See: pkg/coredata/asset.go
asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])

// Multiple rows
assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Asset])

// Map pgx.ErrNoRows to sentinel
if errors.Is(err, pgx.ErrNoRows) {
    return ErrResourceNotFound
}
```

## Filter Pattern (double-pointer three-state logic)

Filter fields use double pointers for three-state logic:
- `nil` = no filter
- `*nil` = IS NULL
- `*val` = equals

```go
// See: pkg/coredata/vendor_filter.go
func (f *VendorFilter) SQLArguments() pgx.StrictNamedArgs {
    args := pgx.StrictNamedArgs{
        "show_on_trust_center": nil,      // always declared
        "has_snapshot_filter":  false,     // always declared
        "filter_snapshot_id":   nil,       // always declared
    }
    if f.showOnTrustCenter != nil {
        args["show_on_trust_center"] = *f.showOnTrustCenter
    }
    return args
}
```

## OrderField Pattern (complete version)

Newer order fields implement the full set: `Column()`, `IsValid()`, `String()`, `MarshalText()`, `UnmarshalText()`. Follow the `InvitationOrderField` pattern in `pkg/coredata/invitation_order_field.go` for new entities.

Some older order fields (e.g., `AssetOrderField`) only implement `Column()` and `String()`. This is an inconsistency -- new entities should use the complete pattern.

## Migration Rules

- Files in `pkg/coredata/migrations/` named `YYYYMMDDTHHMMSSZ.sql` (UTC timestamps)
- One logical change per file
- No indexes by default (only when justified by observed latency)
- No DEFAULT clauses on new tables
- When adding non-nullable columns to existing tables: use DEFAULT to backfill, then drop DEFAULT in the same migration
- ON DELETE CASCADE for audit log FKs to organizations

## Entity Type Registry

`entity_type_reg.go` contains all entity type uint16 constants. Critical rules:
- Never reuse removed type numbers (tombstone with `_ uint16 = N // Removed`)
- Always append new types at the end
- Every new entity must have a case in the `NewEntityFromID` switch

## No Tests in This Package

`pkg/coredata` has no unit tests. Coverage comes from e2e tests in `e2e/console/`. This is intentional -- the package is pure data access with no business logic worth unit-testing in isolation.

## No Logging

The data access layer does not log. Errors are returned to callers who log at the service layer.
