# Probo -- TypeScript Frontend -- packages/n8n-node

> Module-specific notes for `packages/n8n-node` (`@probo/n8n-nodes-probo`)
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md)

## Purpose

An n8n community node package that exposes the Probo API as n8n workflow actions. It provides a single `INodeType` (Probo) with approximately 80 operation files covering CRUD for all major resources, plus a raw GraphQL execute escape hatch.

## Architecture

```
nodes/Probo/
  Probo.node.ts              # INodeType entry point
  Probo.node.json            # Node metadata
  GenericFunctions.ts        # HTTP transport layer (GraphQL requests, pagination)
  actions/
    index.ts                 # Resource registry, dynamic dispatch
    <resource>/
      index.ts               # Operation dropdown + re-exports
      create.operation.ts    # INodeProperties + execute function
      get.operation.ts
      getAll.operation.ts
      update.operation.ts
      delete.operation.ts
credentials/
  ProboApi.credentials.ts    # Server URL + API key
```

## Adding a New Operation

1. Create `<verb>.operation.ts` in `actions/<resource>/` with exported `description` (INodeProperties[]) and `execute` function
2. Add the operation value to the resource's `index.ts` operation dropdown
3. Re-export the operation module from `index.ts` under a camelCase alias matching the operation value
4. If new resource: register it in `actions/index.ts` resource array

## Key Conventions

- **GraphQL queries are inline template literals** inside `execute()` -- never separate files
- **`displayOptions.show`** must be set on every `INodeProperty` to scope fields to their resource + operation
- **`additionalFields`** collection pattern for optional create/update fields
- **Options collection** for optional response shaping (toggling GraphQL fragment inclusion)
- **Cursor pagination**: `proboApiRequestAllItems` loops with page size 100 until `hasNextPage` is false

## Transport Layer

All HTTP calls funnel through `GenericFunctions.ts`:

| Function | API | Purpose |
|----------|-----|---------|
| `proboApiRequest` | Console | Single GraphQL request |
| `proboConnectApiRequest` | Connect/IAM | Single GraphQL request |
| `proboApiRequestAllItems` | Console | Cursor-paginated list |
| `proboConnectApiRequestAllItems` | Connect/IAM | Cursor-paginated list |
| `proboApiMultipartRequest` | Console | File upload (multipart/form-data) |

GraphQL-level errors (HTTP 200 with `response.errors`) are detected and converted to `NodeApiError` with `httpCode: '200'`.

## Known Inconsistency: User Resource

The `user` resource uses non-standard operation values: `createUser`, `getUser`, `listUsers`, `inviteUser`, `updateUser`, `updateMembership`, `removeUser` instead of the standard `create`, `get`, `getAll` pattern used by all other resources. The dispatch logic in `actions/index.ts` relies on exact key matching, so these longer names must match the export aliases in `user/index.ts`.

## Publishing

This package is published to npmjs.org as `@probo/n8n-nodes-probo` on tag push via CI. The version in `package.json` is currently `0.0.1`.

## No Tests

There are no test files in this package. Operations, GenericFunctions pagination, and error handling have zero automated coverage.
