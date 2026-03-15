# pkg/coredata

All raw SQL lives here — never in service or handler packages.

## Entity files

One file per entity (`asset.go`, `vendor.go`, etc.), plus optional companion `_filter.go` and `_order_field.go` files when needed. No codegen — everything is hand-written.

Entity structs do **not** have a `TenantID` field — the `tenant_id` column is provided by the `Scoper` (via `scope.GetTenantID()`) at query time, not stored on the Go struct.

## SQL query pattern

Every query uses raw SQL with `pgx.StrictNamedArgs` and scope injection via `fmt.Sprintf`:

```go
q := `
SELECT id, name, created_at, updated_at
FROM assets
WHERE
    %s
    AND id = @asset_id
LIMIT 1;
`

q = fmt.Sprintf(q, scope.SQLFragment())

args := pgx.StrictNamedArgs{"asset_id": assetID}
maps.Copy(args, scope.SQLArguments())
```

## Scoper interface

Every Load/Insert/Update/Delete method takes a `Scoper` parameter for tenant isolation:
- `Scope` — adds `tenant_id = @tenant_id` WHERE clause
- `NoScope` — returns `TRUE` (for cross-tenant operations)

Insert uses `scope.GetTenantID()` for the tenant_id value.

## Method patterns

| Method | Receiver | Returns | Notes |
|--------|----------|---------|-------|
| `LoadByID` | `*Entity` | `error` | Assigns into receiver via `*e = entity` |
| `LoadAllBy*` | `*Entities` (slice type) | `error` | Paginated with `page.Cursor[OrderField]` |
| `CountBy*` | `*Entities` | `(int, error)` | Uses `COUNT(id)` |
| `Insert` | `*Entity` | `error` | Uses `scope.GetTenantID()` for tenant_id |
| `Update` | `*Entity` | `error` | Uses `RETURNING` to reassign receiver |
| `Delete` | `*Entity` | `error` | — |
| `CursorKey` | `*Entity` | `page.CursorKey` | Switch on OrderField, panic on unknown |
| `AuthorizationAttributes` | `*Entity` | `(map[string]string, error)` | Returns org/tenant IDs for authz |

## Row collection

- Single row: `pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])`
- Multiple rows: `pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[T])`
- Check `pgx.ErrNoRows` → return `ErrResourceNotFound`

## Sentinel errors

Defined in `errors.go`:
- `ErrResourceNotFound` — row not found (`pgx.ErrNoRows`)
- `ErrResourceAlreadyExists` — unique constraint violation
- `ErrResourceInUse` — foreign key constraint prevents deletion

## Filter pattern

Double pointer fields: `nil` = no filter, `*nil` = IS NULL, `*val` = equals.

`SQLFragment()` must return a **static** SQL string (no conditional string building) so the prepared statement is always the same. Use `CASE WHEN` in SQL to handle optional filters.

`SQLArguments()` returns `pgx.StrictNamedArgs` — **every key referenced in the SQL must be set in every code path** (use `nil` for inactive filters). `StrictNamedArgs` rejects missing keys at runtime.

```go
func (f *VendorFilter) SQLArguments() pgx.StrictNamedArgs {
    args := pgx.StrictNamedArgs{
        "show_on_trust_center": nil,
        "has_snapshot_filter":  false,
        "filter_snapshot_id":   nil,
    }

    if f.showOnTrustCenter != nil {
        args["show_on_trust_center"] = *f.showOnTrustCenter
    }

    if f.snapshotID != nil {
        args["has_snapshot_filter"] = true
        if *f.snapshotID != nil {
            args["filter_snapshot_id"] = **f.snapshotID
        }
    }

    return args
}

func (f *VendorFilter) SQLFragment() string {
    return `
(
    CASE
        WHEN @show_on_trust_center::boolean IS NOT NULL THEN
            show_on_trust_center = @show_on_trust_center::boolean
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_snapshot_filter::boolean = false THEN TRUE
        WHEN @has_snapshot_filter::boolean = true AND @filter_snapshot_id::text IS NOT NULL THEN
            snapshot_id = @filter_snapshot_id::text
        WHEN @has_snapshot_filter::boolean = true AND @filter_snapshot_id::text IS NULL THEN
            snapshot_id IS NULL
        ELSE TRUE
    END
)`
}
```

## OrderField pattern

OrderField types must validate their value via `IsValid()` and implement text marshalling:

```go
type InvitationOrderField string

const (
    InvitationOrderFieldCreatedAt InvitationOrderField = "CREATED_AT"
)

func (p InvitationOrderField) Column() string {
    switch p {
    case InvitationOrderFieldCreatedAt:
        return "created_at"
    }
    panic(fmt.Sprintf("unsupported order by: %s", p))
}

func (e InvitationOrderField) IsValid() bool {
    switch e {
    case InvitationOrderFieldCreatedAt:
        return true
    }
    return false
}

func (e InvitationOrderField) String() string { return string(e) }

func (e *InvitationOrderField) UnmarshalText(text []byte) error {
    *e = InvitationOrderField(text)
    if !e.IsValid() {
        return fmt.Errorf("%s is not a valid InvitationOrderField", string(text))
    }
    return nil
}

func (e InvitationOrderField) MarshalText() ([]byte, error) {
    return []byte(e.String()), nil
}
```

## Argument merging

Always use `maps.Copy` to combine args from scope, filter, and cursor. Never manually merge.

## Migrations

Pure SQL files in `pkg/coredata/migrations/` with timestamp names: `YYYYMMDDTHHMMSSZ.sql`.
