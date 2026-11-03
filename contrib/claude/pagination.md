# Cursor previous / next

Bidirectional lists use Relay cursor arguments: `first`/`after` to move forward, `last`/`before` to move backward. Previous must request the page **before the current start cursor**. It must not request `first` with an empty `after`, which is page one.

## Why Previous was reloading page one

Relay's `getOperationVariables` keeps an operation argument's `defaultValue` when the refetch passes `null` (`value != null` is false for `null`). A fragment written as:

```graphql
@argumentDefinitions(
  first: { type: "Int", defaultValue: 15 }
  after: { type: "CursorKey", defaultValue: null }
  last: { type: "Int", defaultValue: null }
  before: { type: "CursorKey", defaultValue: null }
)
```

sends `first: 15` on every request, including Previous. The variables that arrive are `first: 15`, `after: null`, `last: 15`, `before: <start cursor>`.

`cursor.NewCursor` used to take the `first` branch whenever `first` was set, and ignore `before`. That query is the first page.

`before` now wins over `first`, and `after` wins over `last`. A leaked default page size no longer changes direction. The page size still comes from `last` on the way back and `first` on the way forward.

## Client

Use [`useCursorPagination`](../../apps/console/src/lib/relay/useCursorPagination.ts) (the employee portal and compliance portal copies match it):

- Next: `{ first: pageSize, after: endCursor, last: null, before: null }`
- Previous: `{ first: null, after: null, last: pageSize, before: startCursor }`

Pass `first: null` on Previous anyway. Do not "correct" it to `first: pageSize` or drop `before`. The server is what makes the leaked default harmless.

Select `pageInfo { hasNextPage hasPreviousPage startCursor endCursor }` on the connection. Do not use `@connection` / `loadNext` for a previous/next control; that directive only appends the next page.

An in-memory list (a checkpoint array, a loaded slice) uses the same control with an index into the array. Previous decrements that index. It does not reset the index to 0.

## Server

Build the cursor with `types.NewCursor(first, after, last, before, orderBy)`, which calls `cursor.NewCursor`. Do not reimplement the `if first != nil` branch in a resolver: that branch is what dropped `before`.
