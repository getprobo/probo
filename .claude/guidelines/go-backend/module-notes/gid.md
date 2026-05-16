# Probo — Go Backend — pkg/gid

**Purpose.** 24-byte global identifier used everywhere instead of UUIDs.
Embeds the tenant ID, entity-type uint16, timestamp ms, and 6 random
bytes. Wire format is base64url. Implements `sql.Scanner`,
`driver.Valuer`, `MarshalText`, `UnmarshalText` for transparent
PostgreSQL and JSON usage.

See [shared.md § 10](../../shared.md#10-global-identifiers-gid--crossing-the-go--ts-boundary)
for the layout, wire format, and frontend interop.

**Key files.**

- `gid.go` — `GID` type, `New(tenantID, entityType)`,
  `NewEntityFromID`, `TenantID()` accessor.
- `tenant.go` — `TenantID` byte-array type with its own (Un)MarshalText.

**How to extend.** GID layout itself is fixed; extending means adding a
new entity type in `pkg/coredata/entity_type_reg.go` (see
[coredata.md](./coredata.md)).

**Top pitfalls.**

- Using `github.com/google/uuid` instead of the GID toolchain. The
  project removed `google/uuid` deliberately — use `pkg/crypto/rand` for
  raw randomness or `gid.New(...)` for entity IDs.
- Calling `gid.New(tenantID, ...)` from inside `pkg/coredata`. By
  convention IDs are minted **in the service layer** (`pkg/probo`) and
  passed to `coredata.Insert` already populated.
