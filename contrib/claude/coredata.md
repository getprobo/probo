# Coredata — Data Access Layer

All raw SQL lives in `pkg/coredata` — never in service, handler, or resolver packages. One file per entity, with companion `*_filter.go` and `*_order_field.go` files when needed.

- Database: `go.gearno.de/kit/pg`
- UUID: `go.gearno.de/crypto/uuid` (never use `github.com/google/uuid`)

## Entity struct pattern

Every entity uses `gid.GID` for its ID, `db` tags for pgx mapping, and `CreatedAt`/`UpdatedAt` timestamps. The `tenant_id` column exists in the database but is **never** stored on the Go struct — it is injected at query time via `Scoper`.

```go
type (
	Asset struct {
		ID             gid.GID   `db:"id"`
		SnapshotID     *gid.GID  `db:"snapshot_id"`
		Name           string    `db:"name"`
		OrganizationID gid.GID   `db:"organization_id"`
		AssetType      AssetType `db:"asset_type"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	Assets []*Asset
)
```

Use pointer types (`*T`) for nullable database columns.

## Denormalized `organization_id`

Every entity that belongs to an organization carries its own `organization_id` column and Go field — even when the organization can be inferred by walking a foreign key chain. This avoids JOIN queries in `AuthorizationAttributes`, which is called on every authorized request.

When creating a child entity, copy `OrganizationID` from its parent (e.g. from the banner when creating a category or version). The `AuthorizationAttributes` method then returns the field directly without any database query:

```go
func (c *CookieCategory) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	return map[string]string{"organization_id": c.OrganizationID.String()}, nil
}
```

## Scoper interface

`Scoper` provides tenant isolation. Two implementations:


| Type      | Constructor                                         | `SQLFragment()`            | `GetTenantID()`         | Use case                              |
| --------- | --------------------------------------------------- | -------------------------- | ----------------------- | ------------------------------------- |
| `Scope`   | `NewScope(tenantID)` or `NewScopeFromObjectID(gid)` | `"tenant_id = @tenant_id"` | Returns tenant ID       | Multi-tenant queries (default)        |
| `NoScope` | `NewNoScope()`                                      | `"TRUE"`                   | **Panics** — never call | Cross-tenant / administrative queries |


Always inject `tenant_id` at INSERT time using `scope.GetTenantID()`, never from the struct.

## SQL query composition

All queries use `fmt.Sprintf` to inject scope/filter/cursor fragments, then `pgx.StrictNamedArgs` for parameters. Merge args with `maps.Copy`.

```go
q := `
SELECT id, name, created_at, updated_at
FROM assets
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
LIMIT %d;
`

q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment(), cursor.Limit())

args := pgx.StrictNamedArgs{"organization_id": organizationID}
maps.Copy(args, scope.SQLArguments())
maps.Copy(args, filter.SQLArguments())
maps.Copy(args, cursor.SQLArguments())
```

**All SQL must be static** after `fmt.Sprintf()` injection — no conditional string building. Use `CASE WHEN` in SQL for optional filter logic.

**Use Go enum constants as named parameters** — never hardcode string literals like `'ACTIVE'` or `'PUBLISHED'` in SQL. Use a named parameter (`@state`) and pass the Go constant via `pgx.StrictNamedArgs`:

```go
// Good — Go constant as named parameter
q := `SELECT ... FROM cookie_banners WHERE id = @id AND state = @state;`
args := pgx.StrictNamedArgs{
    "id":    bannerID,
    "state": CookieBannerStateActive,
}

// Bad — hardcoded string literal in SQL
q := `SELECT ... FROM cookie_banners WHERE id = @id AND state = 'ACTIVE';`
```

This ensures the compiler catches renamed or removed enum values instead of silently producing wrong results at runtime.

## Standard method signatures


| Method                                                   | Receiver    | Returns                      | Purpose                              |
| -------------------------------------------------------- | ----------- | ---------------------------- | ------------------------------------ |
| `LoadByID(ctx, conn, scope, id)`                         | `*Entity`   | `error`                      | Single entity by ID                  |
| `LoadBy*(ctx, conn, scope, key)`                         | `*Entity`   | `error`                      | Single entity by unique key          |
| `LoadAllBy*(ctx, conn, scope, parentID, cursor, filter)` | `*Entities` | `error`                      | Paginated list                       |
| `CountBy*(ctx, conn, scope, parentID, filter)`           | `*Entities` | `(int, error)`               | Count matching rows                  |
| `Insert(ctx, conn, scope)`                               | `*Entity`   | `error`                      | Insert, uses `scope.GetTenantID()`   |
| `Update(ctx, conn, scope)`                               | `*Entity`   | `error`                      | Update with `RETURNING`              |
| `Delete(ctx, conn, scope)`                               | `*Entity`   | `error`                      | Delete entity                        |
| `CursorKey(orderField)`                                  | `*Entity`   | `page.CursorKey`             | Cursor for pagination                |
| `AuthorizationAttributes(ctx, conn)`                     | `*Entity`   | `(map[string]string, error)` | Attributes for IAM policy evaluation |


## Row collection

```go
// Single row
rows, err := conn.Query(ctx, q, args)
asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])
if errors.Is(err, pgx.ErrNoRows) {
    return ErrResourceNotFound
}
*a = asset

// Multiple rows
rows, err := conn.Query(ctx, q, args)
assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Asset])
*a = assets
```

## Sentinel errors

```go
var (
    ErrResourceNotFound      = errors.New("resource not found")
    ErrResourceAlreadyExists = errors.New("resource already exists")
    ErrResourceInUse         = errors.New("resource is in use")
)
```

Map `pgx.ErrNoRows` to `ErrResourceNotFound`. Check unique constraint violations for `ErrResourceAlreadyExists`, foreign key violations for `ErrResourceInUse`.

## Filters

Filters implement `SQLFragment() string` and `SQLArguments() pgx.NamedArgs`. Use double pointers for three-state filtering: `nil` = no filter, `*nil` = IS NULL, `*val` = equals.

```go
type AssetFilter struct {
    snapshotID **gid.GID
}

func NewAssetFilter(snapshotID **gid.GID) *AssetFilter {
    return &AssetFilter{snapshotID: snapshotID}
}

func (f *AssetFilter) SQLFragment() string {
    if f.snapshotID == nil {
        return "TRUE"
    }

    if *f.snapshotID == nil {
        return "snapshot_id IS NULL"
    }

    return "snapshot_id = @filter_snapshot_id"
}

func (f *AssetFilter) SQLArguments() pgx.NamedArgs {
    if f.snapshotID == nil || *f.snapshotID == nil {
        return pgx.NamedArgs{}
    }

    return pgx.NamedArgs{"filter_snapshot_id": **f.snapshotID}
}
```

For complex multi-field filters, use `CASE WHEN` in SQL and always declare all argument keys in every code path (use `nil` for inactive ones).

## Order fields

String-based enums with `Column()`, `IsValid()`, `String()`, and `MarshalText`/`UnmarshalText`:

```go
type AssetOrderField string

const (
    AssetOrderFieldCreatedAt AssetOrderField = "CREATED_AT"
    AssetOrderFieldName      AssetOrderField = "NAME"
)
```

Each entity implements `CursorKey(field)` returning `page.NewCursorKey(entity.ID, sortValue)`, with a `panic` on unknown fields.

## Entity type registry

Each entity gets a unique `uint16` constant in `entity_type_reg.go`. **Never reuse** removed type numbers — use `_` placeholder. Register new entities in the `NewEntityFromID` switch statement.

## Migrations

- Files in `pkg/coredata/migrations/` use timestamp naming: `YYYYMMDDTHHMMSSZ.sql` (UTC).
- Run `date -u +"%Y%m%dT%H%M%SZ.sql"` to get the name of the new migration file.
- One logical change per file.

**No indexes by default.** Only add indexes when justified by observed query latency in production environments. Do not speculatively create indexes on new tables or columns. This rule does not apply to indexes that enforce constraints, such as unique indexes.

**Avoid default values.** Columns should not have `DEFAULT` clauses. When adding a non-nullable column to an existing table, use a `DEFAULT` to backfill existing rows, then drop it in the same migration.

## Sensitive data protection

Every column that stores a secret, credential, or private key **must** be protected at rest in the application layer. Never store sensitive values as plaintext in the database. There are three protection strategies depending on the data's nature.

### Strategy 1 — SHA-256 hash (high-entropy tokens)

Use for values generated by the application with guaranteed entropy: bearer tokens, API keys, SCIM tokens, one-time tokens, SAML relay state tokens. These values are random and never chosen by a human, so a fast non-reversible hash is sufficient.

- Package: `pkg/crypto/hash` → `hash.SHA256Hex([]byte) string`
- DB column type: `BYTEA` (store the raw hash bytes) or `TEXT` (store hex-encoded hash)
- Go field name: `Hashed*` (e.g. `HashedToken`, `HashedValue`)
- Lookup: compute SHA-256 of the presented token, then `WHERE hashed_token = @hashed_token`
- The plaintext token is returned to the caller **once** at creation time and never stored

Existing examples: `Token.HashedValue`, `SCIMConfiguration.HashedToken`.

```go
// At creation time — hash before insert
hashedValue := hash.SHA256Hex([]byte(rawToken))
token.HashedValue = []byte(hashedValue)

// At verification time — hash the presented value, then query
hashedValue := hash.SHA256Hex([]byte(presentedToken))
token.LoadByHashedValueForUpdate(ctx, conn, []byte(hashedValue))
```

### Strategy 2 — PBKDF2 (human-chosen passwords)

Use for values chosen by humans with low or unpredictable entropy: passwords, passphrases, PINs. PBKDF2 with HMAC-SHA256 pepper provides brute-force resistance.

- Package: `pkg/crypto/passwdhash`
- DB column type: `BYTEA NOT NULL`
- Go field name: `HashedPassword`
- Hash on write: `profile.HashPassword([]byte(password))`
- Compare on read: `profile.ComparePasswordAndHash([]byte(password), storedHash)`
- Parameters: minimum 600 000 iterations, 32-byte salt, 32-byte pepper

Existing example: `Identity.HashedPassword`.

```go
// At registration / password change
hashed, err := passwdProfile.HashPassword([]byte(plainPassword))
identity.HashedPassword = hashed

// At login
ok, err := passwdProfile.ComparePasswordAndHash([]byte(inputPassword), identity.HashedPassword)
```

### Strategy 3 — AES-256-GCM encryption (secrets that must be read back)

Use for values the application needs to decrypt later: OAuth `access_token` / `refresh_token`, `client_secret`, API keys for third-party services, TLS private keys, webhook signing secrets.

- Package: `pkg/crypto/cipher`
- DB column type: `BYTEA NOT NULL`
- Go field name: `Encrypted*` (e.g. `EncryptedConnection`, `EncryptedSigningSecret`)
- Encrypt on write: `cipher.Encrypt(plaintext, encryptionKey)`
- Decrypt on read: `cipher.Decrypt(ciphertext, encryptionKey)`
- The `cipher.EncryptionKey` is a 32-byte key loaded from configuration — never stored in the database

Existing examples: `Connector.EncryptedConnection`, `WebhookSubscription.EncryptedSigningSecret`, `CustomDomain.EncryptedSSLPrivateKey`.

```go
// On insert / update — encrypt before writing
connection, _ := json.Marshal(c.Connection)
encrypted, err := cipher.Encrypt(connection, encryptionKey)
c.EncryptedConnection = encrypted

// On load — decrypt after reading
plaintext, err := cipher.Decrypt(c.EncryptedConnection, encryptionKey)
json.Unmarshal(plaintext, &c.Connection)
```

### Decision table

| Data kind | Entropy source | Needs decryption? | Strategy | Go field prefix | Package |
|-----------|---------------|-------------------|----------|----------------|---------|
| Bearer / API / SCIM / one-time tokens | Application CSPRNG | No — compare by hash | SHA-256 | `Hashed*` | `pkg/crypto/hash` |
| Passwords, passphrases | Human | No — compare with constant-time check | PBKDF2 | `HashedPassword` | `pkg/crypto/passwdhash` |
| OAuth tokens, client secrets, private keys, signing secrets | External provider or application | Yes — must read back | AES-256-GCM | `Encrypted*` | `pkg/crypto/cipher` |

### Rules

- **Never store a plaintext secret** in a `TEXT` or `VARCHAR` column. If a column holds a secret, it must be `BYTEA` with one of the three strategies above.
- **Never log sensitive values.** Do not pass raw tokens, passwords, or decrypted secrets to `slog` or `fmt.Errorf` messages.
- **Name columns and fields consistently.** Use `hashed_` prefix for hashed values and `encrypted_` prefix for encrypted values. The Go struct field must mirror this (e.g. `HashedToken`, `EncryptedConnection`).
- **Return plaintext tokens once.** For SHA-256-hashed tokens, return the raw token to the caller at creation time only. After that, the application only ever sees the hash.
- **Migration columns.** When adding a new sensitive column, always use `BYTEA`. Never add `DEFAULT` on sensitive columns.

## New entity checklist

1. **Entity file** (`entity.go`) — struct with `db` tags, slice type alias, `LoadByID`, `Insert`, `Update`, `Delete`, `CursorKey`, `AuthorizationAttributes`
2. **Filter file** (`entity_filter.go`) — filter struct, `NewEntityFilter`, `SQLFragment`, `SQLArguments`
3. **Order field file** (`entity_order_field.go`) — order field type, constants, `Column`, `IsValid`, marshaling
4. **Entity type constant** — add to `entity_type_reg.go` and `NewEntityFromID`
5. **Migration** — `YYYYMMDDTHHMMSSZ.sql` with CREATE TABLE
