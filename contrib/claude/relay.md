# Relay Cursor Pagination

This project implements [Relay-style cursor pagination](https://relay.dev/graphql/connections.htm) for all list fields across GraphQL APIs.

## Schema types

Every paginated field uses the same set of types:

```graphql
type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: CursorKey
    endCursor: CursorKey
}

enum OrderDirection {
    ASC
    DESC
}
```

Each entity defines its own order field enum, order input, connection, and edge:

```graphql
enum VendorOrderField {
    CREATED_AT
    NAME
}

input VendorOrder {
    direction: OrderDirection!
    field: VendorOrderField!
}

type VendorConnection {
    totalCount: Int!
    edges: [VendorEdge!]!
    pageInfo: PageInfo!
}

type VendorEdge {
    cursor: CursorKey!
    node: Vendor!
}
```

## Field arguments

Connection fields on parent types always use the standard Relay arguments:

```graphql
type Organization {
    vendors(
        first: Int
        after: CursorKey
        last: Int
        before: CursorKey
        orderBy: VendorOrder
        filter: VendorFilter
    ): VendorConnection!
}
```

- `first` / `after` — forward pagination (returns `Head` position)
- `last` / `before` — backward pagination (returns `Tail` position)
- `orderBy` — optional, defaults to `CREATED_AT` / `DESC`
- `filter` — optional, entity-specific filtering

## Cursor format

Cursors are opaque `CursorKey` scalars. Internally they encode as base64url(JSON):

```
["<entity_global_id>", <sort_field_value>]
```

For example, a cursor sorting by `created_at` encodes the entity ID and its `created_at` timestamp. This enables keyset pagination — the database uses the cursor values to seek directly to the right position instead of using OFFSET.

## Keyset pagination

The database query uses the cursor to build a WHERE clause that skips to the correct position:

- For `DESC` ordering: rows where `(field <= cursor_value) AND NOT (field = cursor_value AND id > cursor_id)`
- For `ASC` ordering: rows where `(field >= cursor_value) AND NOT (field = cursor_value AND id < cursor_id)`

The query fetches `size + 1` (or `size + 2` when a cursor is provided) rows to detect whether more pages exist in either direction. `NewPage` trims the extra rows and sets `hasNextPage` / `hasPreviousPage` accordingly.

For backward pagination (`last` / `before`), the SQL sort direction is reversed, and the result slice is reversed back to the correct order before building edges.

## Default page size

When neither `first` nor `last` is provided, the default page size is **25**.

## Adding a new paginated field — checklist

1. **Schema** — add `enum XxxOrderField`, `input XxxOrder`, `type XxxConnection`, `type XxxEdge`, and the connection field with Relay arguments on the parent type
2. **Coredata** — add `*_order_field.go` (with `Column()`, `IsValid()`, marshaling), `CursorKey(field)` method on the entity, and the `LoadAllBy*` query using cursor SQL fragments + `page.NewPage()`
3. **API types** — add `*_connection.go` with `OrderBy` alias, connection struct, `NewXxxConnection`, `NewXxxEdge`
4. **Resolver** — implement the resolver (authorize, build order, build cursor, call service, build connection)
5. **Codegen** — run `go generate` for the relevant API package
