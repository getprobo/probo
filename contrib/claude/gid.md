# GID — Global Identifiers (`pkg/gid`)

Every entity ID in the system is a 24-byte tenant-scoped GID, serialized as base64url in the database, JSON, and API surfaces.

## GID layout (24 bytes / 192 bits)

| Bytes | Size | Content |
|-------|------|---------|
| 0–7 | 8 bytes | Tenant ID |
| 8–9 | 2 bytes | Entity type (`uint16`) |
| 10–17 | 8 bytes | Timestamp (milliseconds since epoch) |
| 18–23 | 6 bytes | Random data |

## Creating a GID

GIDs are created in the **service layer** (e.g. `pkg/probo/*_service.go`), not in coredata `Insert` methods. The entity type constant comes from `pkg/coredata/entity_type_reg.go`:

```go
assetID := gid.New(s.svc.scope.GetTenantID(), coredata.AssetEntityType)

asset := &coredata.Asset{
    ID:             assetID,
    OrganizationID: req.OrganizationID,
    Name:           req.Name,
    CreatedAt:      now,
    UpdatedAt:      now,
}

err := asset.Insert(ctx, conn, s.svc.scope)
```

`gid.New` panics on random source failure (should never happen). Use `gid.NewGID` if you need the error.

## Extracting fields

```go
id.TenantID()    // TenantID (first 8 bytes)
id.EntityType()  // uint16 (bytes 8–9)
id.Timestamp()   // time.Time (bytes 10–17)
```

## Parsing and serialization

- `gid.ParseGID(encoded)` — base64url string to GID
- `gid.String()` — GID to base64url string
- Implements `sql.Scanner`, `driver.Valuer`, `MarshalText`, `UnmarshalText`
- `gid.Nil` — zero-value GID

## TenantID

`TenantID` is an 8-byte type with its own layout:

| Bytes | Size | Content |
|-------|------|---------|
| 0–2 | 3 bytes | Machine ID (random per process) |
| 3–5 | 3 bytes | Timestamp (truncated Unix seconds) |
| 6–7 | 2 bytes | Atomic counter |

Create with `gid.NewTenantID()`. Check with `tenantID.IsValid()` (non-nil). Same serialization interfaces as GID (base64url, SQL scanner/valuer).

## Entity type registry

All entity type constants live in `pkg/coredata/entity_type_reg.go` as sequential `uint16` values:

```go
const (
    OrganizationEntityType uint16 = 0
    FrameworkEntityType    uint16 = 1
    MeasureEntityType      uint16 = 2
    // ...
    _                      uint16 = 8  // PeopleEntityType - removed
    // ...
)
```

**Never reuse removed type numbers.** Use `_` placeholders with a comment noting what was removed. New types get the next available number.

`NewEntityFromID(id gid.GID) (any, bool)` switches on `id.EntityType()` and returns a pointer to the concrete coredata struct with `ID: id` set, or `nil, false` for unknown types. Add a case here when registering a new entity type.

## New entity checklist (GID-related steps)

1. Add `FooEntityType uint16 = N` in the `const` block in `entity_type_reg.go` (next sequential number)
2. Add a `case FooEntityType` in `NewEntityFromID` returning `&Foo{ID: id}, true`
3. In the service `Create` method, call `gid.New(scope.GetTenantID(), coredata.FooEntityType)` to generate the ID before `Insert`
