# n8n Node (`packages/n8n-node`)

Community node package `@probo/n8n-nodes-probo` exposing the Probo API as n8n operations. One `Probo` node with many resources; each resource maps to a set of GraphQL operations against the Console or Connect API.

## Directory structure

```
packages/n8n-node/
  credentials/ProboApi.credentials.ts   # OAuth access token credential
  nodes/Probo/
    Probo.node.ts                       # Node class — resource picker, dispatch
    Probo.node.json                     # n8n codex metadata
    GenericFunctions.ts                 # GraphQL request helpers, pagination
    actions/
      index.ts                          # Resource registry, dispatch, field aggregators
      <resource>/
        index.ts                        # Operation dropdown + spread descriptions + re-exports
        create.operation.ts             # One file per operation
        get.operation.ts
        getAll.operation.ts
        update.operation.ts
        delete.operation.ts
        ...
```

## Resource registration

Two places must be updated when adding a resource:

**1. `actions/index.ts`** — import the module and add it to the `resources` map:

```typescript
import * as myresource from './myresource';

export const resources: Record<string, ResourceModule> = {
    // ... existing resources ...
    myresource: myresource as ResourceModule,
};
```

**2. `Probo.node.ts`** — add a Resource dropdown entry in the `properties` array:

```typescript
{
    name: 'My Resource',
    value: 'myresource',
    description: 'Manage my resources',
},
```

The `value` must match the key in `resources` and the `displayOptions.show.resource` in every operation file.

## Per-resource file pattern

### `<resource>/index.ts`

1. Import each `*.operation.ts` as a namespace.
2. Export `description` — the operation dropdown (gated with `displayOptions.show.resource`) plus all spread operation descriptions.
3. Re-export each operation module with a name matching its `operation` value.

```typescript
import * as createOp from './create.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';

export const description: INodeProperties[] = [
    {
        displayName: 'Operation',
        name: 'operation',
        type: 'options',
        noDataExpression: true,
        displayOptions: {
            show: {
                resource: ['myresource'],
            },
        },
        options: [
            {
                name: 'Create',
                value: 'create',
                description: 'Create a new resource',
                action: 'Create a resource',
            },
            // ... more operations ...
        ],
        default: 'create',
    },
    ...createOp.description,
    ...getOp.description,
    ...getAllOp.description,
];

export {
    createOp as create,
    getOp as get,
    getAllOp as getAll,
};
```

Export names (`create`, `get`, `getAll`, etc.) **must match** the operation `value` strings — `getExecuteFunction` uses them as keys.

### `<resource>/<verb>.operation.ts`

Each file exports `description` (field definitions) and `execute` (the handler):

```typescript
export const description: INodeProperties[] = [
    {
        displayName: 'Organization ID',
        name: 'organizationId',
        type: 'string',
        displayOptions: {
            show: {
                resource: ['myresource'],
                operation: ['create'],
            },
        },
        default: '',
        required: true,
    },
    // ... more fields ...
    {
        displayName: 'Additional Fields',
        name: 'additionalFields',
        type: 'collection',
        placeholder: 'Add Field',
        default: {},
        displayOptions: {
            show: {
                resource: ['myresource'],
                operation: ['create'],
            },
        },
        options: [
            // optional field definitions
        ],
    },
];

export async function execute(
    this: IExecuteFunctions,
    itemIndex: number,
): Promise<INodeExecutionData> {
    const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
    const name = this.getNodeParameter('name', itemIndex) as string;

    const query = `
        mutation CreateMyResource($input: CreateMyResourceInput!) {
            createMyResource(input: $input) {
                myResourceEdge {
                    node {
                        id
                        name
                    }
                }
            }
        }
    `;

    const responseData = await proboApiRequest.call(this, query, {
        input: { organizationId, name },
    });

    return {
        json: responseData,
        pairedItem: { item: itemIndex },
    };
}
```

List operations return one n8n item per record. Build that array with `executionItems(records, itemIndex)` from `GenericFunctions.ts`, and type `execute` as `Promise<INodeExecutionData[]>`. `Probo.node.ts` flattens the array into the node's output. An empty list produces no items.

Organization, document, person, document version, and the other named records (framework, control, risk, internal control, asset, audit, task, third party, compliance portal, cookie banner, access review campaign) are resource locators in `locators.ts`. Read the ID with `getNodeParameter(name, itemIndex, '', { extractValue: true })`. Register the matching `search*` method on the node `listSearch` methods when a node shows that field. People and most records search inside the selected organization. Document versions search inside the selected document.

## GraphQL helpers

All helpers live in `GenericFunctions.ts`.

| Helper | API endpoint | Use case |
|--------|-------------|----------|
| `proboApiRequest` | `/api/console/v1/graphql` | Single mutations and queries |
| `proboConnectApiRequest` | `/api/connect/v1/graphql` | Organization/user operations (IAM) |
| `proboApiRequestAllItems` | Console API | Cursor-paginated list queries |
| `proboConnectApiRequestAllItems` | Connect API | Cursor-paginated list queries (IAM) |
| `proboApiMultipartRequest` | Console API | File upload mutations (multipart/form-data) |

### Pagination (`proboApiRequestAllItems`)

Caller supplies a `getConnection` function that navigates from the raw GraphQL response to the Relay connection object (must have `edges` and `pageInfo`):

```typescript
const items = await proboApiRequestAllItems.call(
    this,
    query,
    { organizationId },
    (response) => {
        const data = response?.data as IDataObject | undefined;
        const node = data?.node as IDataObject | undefined;
        return node?.myResources as IDataObject | undefined;
    },
    returnAll,
    limit,
);
```

Internal page size is 100. When `returnAll` is false, stops at `limit`.

### Update pattern

For nullable fields, empty string means "clear the value":

```typescript
if (additionalFields.description !== undefined) {
    input.description = additionalFields.description === '' ? null : additionalFields.description;
}
```

## Adding a new resource — checklist

1. **Directory** — create `nodes/Probo/actions/<resource>/` with `index.ts` and one `*.operation.ts` per operation
2. **Operations** — each file exports `description` (fields gated with `displayOptions`) and `execute` (reads params, calls GraphQL, returns `{ json, pairedItem }`)
3. **Index** — `<resource>/index.ts` defines the operation dropdown, spreads all descriptions, re-exports ops with matching value names
4. **Register** — import and add to `resources` map in `actions/index.ts`
5. **Node** — add Resource dropdown entry in `Probo.node.ts` properties
6. **Verify** — `npx n8n-node lint` must pass
