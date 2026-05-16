# Probo — TypeScript Frontend — @probo/n8n-node

The **community n8n node** for Probo. **236 files**, organised as resource × operation feature
slices. Published to npm as `@probo/n8n-nodes-probo` from the release pipeline.

> Authoritative source: [`contrib/claude/n8n.md`](../../../contrib/claude/n8n.md).

## Layout

```
packages/n8n-node/nodes/Probo/
  Probo.node.ts                         ← node manifest, properties array (resource selector)
  actions/
    index.ts                            ← resources map
    <resource>/
      <operation>.operation.ts          ← one file per operation
      index.ts                          ← exports a resource module
  helpers/
    proboApiRequest.ts                  ← Console / Trust API helper
    proboConnectApiRequest.ts           ← IAM (Connect) API helper
  graphql/                              ← generated GraphQL operations
```

The build generates GraphQL query strings into `graphql/` from `.graphql` files; **esbuild**
bundles the node for n8n consumption.

## How to add a new operation

1. Create `actions/<resource>/<operation>.operation.ts` exporting `description`, `properties`,
   and an `execute` function.
2. Add the export to `actions/<resource>/index.ts` — **the export name MUST match the operation's
   `value` string** referenced in the `Probo.node.ts` properties array. A mismatch makes the
   operation invisible at runtime with no error.
3. Pick the right API helper:
   - `proboApiRequest` for Console / Trust (`/api/console/v1/graphql`, `/api/trust/v1/graphql`).
   - **`proboConnectApiRequest`** for **IAM** (`/api/connect/v1/graphql`).
4. Run `npx n8n-node lint` — required by CI.

The four-surface API rule (see [shared.md § 3](../shared.md#3-the-four-surface-api-rule)) means
this update happens in lockstep with GraphQL schema, MCP tool, and CLI command additions.

## Top pitfalls

1. **Resource export-name mismatch** — see
   [pitfalls.md § 19](../pitfalls.md#19-packagesn8n-node-resource-export-name-mismatch).
2. **Wrong API helper for IAM operations** — see
   [pitfalls.md § 20](../pitfalls.md#20-packagesn8n-node-iam-operations-using-proboapirequest).
