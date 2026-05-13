# GraphQL (Go Backend — gqlgen)

Schema-first GraphQL using [gqlgen](https://gqlgen.com/). The schema is hand-written and split into per-entity files under `graphql/`; Go types and resolvers are generated.

## Schema file organization

Each API's schema lives in `pkg/server/api/{api}/v1/graphql/` as multiple `.graphql` files, one per coredata model:

- `base.graphql` — directives, scalars, Node interface, PageInfo, OrderDirection, root Query/Mutation/Organization types
- Entity files (e.g., `thirdParty.graphql`, `control.graphql`) — use `extend type Mutation` to add their mutations.

gqlgen's `follow-schema` layout generates one resolver file per schema file (e.g., `thirdParty.resolvers.go`). Types that get extended across files (Organization, Mutation, Viewer, TrustCenter) must be defined in `base.graphql`.

### `extend type` restrictions

**The only permitted use of `extend type` is `extend type Mutation`.** Never use `extend type` on any other type — not on entity types, not on `Organization`, not on `Query`. If `CookieBanner` needs a `consentRecords` connection, add the field directly to the `CookieBanner` type definition in `cookie_banner.graphql` — do not write `extend type CookieBanner` in another file. This keeps each entity's full field set visible in one place and avoids resolver mis-routing across generated files.

## Connection types and `@goModel`

**Always define a custom Go type for connection types** using the `@goModel` directive. The model path points to the `types` package for the relevant API. The `totalCount` field must use `@goField(forceResolver: true)`. Edge types do not need `@goModel`.

```graphql
type ThirdPartyConnection
    @goModel(
        model: "go.probo.inc/probo/pkg/server/api/console/v1/types.ThirdPartyConnection"
    ) {
    totalCount: Int! @goField(forceResolver: true)
    edges: [ThirdPartyEdge!]!
    pageInfo: PageInfo!
}

type ThirdPartyEdge {
    cursor: CursorKey!
    node: ThirdParty!
}
```

Without `@goModel`, gqlgen generates a default struct that lacks the custom fields (`ParentID`, `Resolver`, `Filter`) needed by the pagination resolvers.

## Enums and `@goModel` / `@goEnum`

Map GraphQL enums to existing Go types using `@goModel` on the enum and `@goEnum` on each value:

```graphql
enum ThirdPartyOrderField
    @goModel(model: "go.probo.inc/probo/pkg/coredata.ThirdPartyOrderField") {
    CREATED_AT
        @goEnum(value: "go.probo.inc/probo/pkg/coredata.ThirdPartyOrderFieldCreatedAt")
    NAME
        @goEnum(value: "go.probo.inc/probo/pkg/coredata.ThirdPartyOrderFieldName")
}
```

## Schema directives


| Directive                       | Target                                                           | Purpose                                                      |
| ------------------------------- | ---------------------------------------------------------------- | ------------------------------------------------------------ |
| `@goModel(model: "...")`        | `OBJECT`, `ENUM`, `INPUT_OBJECT`, `SCALAR`, `INTERFACE`, `UNION` | Map GraphQL type to existing Go type                         |
| `@goEnum(value: "...")`         | `ENUM_VALUE`                                                     | Map enum value to Go constant                                |
| `@goField(forceResolver: true)` | `FIELD_DEFINITION`                                               | Force a resolver function instead of struct field            |
| `@goField(name: "...")`         | `FIELD_DEFINITION`, `INPUT_FIELD_DEFINITION`                     | Override Go field name                                       |
| `@goField(omittable: true)`     | `INPUT_FIELD_DEFINITION`                                         | Use `graphql.Omittable[T]` for distinguishing null vs absent |


## Cursor pagination schema types

Every paginated field uses shared base types plus entity-specific types:

```graphql
type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: CursorKey
    endCursor: CursorKey
}

enum OrderDirection
    @goModel(model: "go.probo.inc/probo/pkg/page.OrderDirection") {
    ASC @goEnum(value: "go.probo.inc/probo/pkg/page.OrderDirectionAsc")
    DESC @goEnum(value: "go.probo.inc/probo/pkg/page.OrderDirectionDesc")
}
```

Each entity defines: `enum XxxOrderField`, `input XxxOrder`, `type XxxConnection` (with `@goModel`), `type XxxEdge`.

Connection fields on parent types use standard Relay arguments:

```graphql
type Organization {
    thirdParties(
        first: Int
        after: CursorKey
        last: Int
        before: CursorKey
        orderBy: ThirdPartyOrder
        filter: ThirdPartyFilter
    ): ThirdPartyConnection!
}
```

## Go connection type pattern

Each connection type lives in `types/*_connection.go` and follows this structure:

```go
type (
    ThirdPartyOrderBy OrderBy[coredata.ThirdPartyOrderField]

    ThirdPartyConnection struct {
        TotalCount int
        Edges      []*ThirdPartyEdge
        PageInfo   PageInfo

        Resolver any
        ParentID gid.GID
    }
)

func NewThirdPartyConnection(
    p *page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField],
    parentType any,
    parentID gid.GID,
) *ThirdPartyConnection {
    edges := make([]*ThirdPartyEdge, len(p.Data))
    for i, v := range p.Data {
        edges[i] = NewThirdPartyEdge(v, p.Cursor.OrderBy.Field)
    }

    return &ThirdPartyConnection{
        Edges:    edges,
        PageInfo: *NewPageInfo(p),

        Resolver: parentType,
        ParentID: parentID,
    }
}

func NewThirdPartyEdge(
    v *coredata.ThirdParty,
    orderBy coredata.ThirdPartyOrderField,
) *ThirdPartyEdge {
    return &ThirdPartyEdge{
        Cursor: v.CursorKey(orderBy),
        Node:   NewThirdParty(v),
    }
}
```

## Cursor format

Cursors are opaque `CursorKey` scalars. Internally they encode as base64url(JSON):

```
["<entity_global_id>", <sort_field_value>]
```

This enables keyset pagination — the database seeks directly to the right position instead of using OFFSET.

## Keyset pagination

The database query uses the cursor to build a WHERE clause:

- `DESC`: rows where `(field <= cursor_value) AND NOT (field = cursor_value AND id > cursor_id)`
- `ASC`: rows where `(field >= cursor_value) AND NOT (field = cursor_value AND id < cursor_id)`

The query fetches `size + 1` (or `size + 2` with a cursor) rows to detect whether more pages exist. `NewPage` trims extra rows and sets `hasNextPage` / `hasPreviousPage`.

For backward pagination (`last` / `before`), SQL sort direction is reversed, then the result slice is reversed back.

Default page size is **25** when neither `first` nor `last` is provided.

## Adding a new paginated field — checklist

1. **Schema** — add `enum XxxOrderField` (with `@goModel`/`@goEnum`), `input XxxOrder`, `type XxxConnection` (with `@goModel` and `totalCount` using `@goField(forceResolver: true)`), `type XxxEdge`, and the connection field with Relay arguments on the parent type
2. **Coredata** — add `*_order_field.go` (with `Column()`, `IsValid()`, marshaling), `CursorKey(field)` method on the entity, and the `LoadAllBy`* query using cursor SQL fragments + `page.NewPage()`
3. **API types** — add `*_connection.go` with `OrderBy` alias, connection struct, `NewXxxConnection`, `NewXxxEdge`
4. **Resolver** — implement the resolver (authorize, build order, build cursor, call service, build connection)
5. **Codegen** — run `go generate` for the relevant API package

