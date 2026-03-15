# Relay (Frontend GraphQL Client)

The console app uses [Relay](https://relay.dev/) as its GraphQL client. All GraphQL operations are defined inline with the `graphql` template tag from `relay-runtime` — there are no separate `.graphql` files on the frontend.

## Environments

Two Relay environments connect to two separate GraphQL APIs:

| Environment | Endpoint | Purpose |
|-------------|----------|---------|
| `coreEnvironment` | `/api/console/v1/graphql` | Main application data |
| `iamEnvironment` | `/api/connect/v1/graphql` | Authentication / identity |

Configured in `apps/console/src/environments.ts`. Each has its own store with 1-minute query cache expiration.

## Relay compiler

Config lives in `apps/console/relay.config.json` with two projects (`core`, `iam`) mapped to different source directories and schemas. Generated files go into `__generated__/` directories.

```sh
npm run relay          # clean + compile
npm run relay-compile  # compile only
```

Custom scalar mappings: `Datetime → string`, `GID → string`, `CursorKey → string`, `Duration → string`, `BigInt → number`, `EmailAddr → string`.

## Colocated queries

Queries are defined inline in the file that uses them. Route-level queries are preloaded in the router loader before the component renders:

```tsx
// In route definition
{
  path: "vendors",
  loader: loaderFromQueryLoader(({ organizationId }) =>
    loadQuery<VendorGraphListQuery>(coreEnvironment, vendorsQuery, {
      organizationId,
      snapshotId: null,
    }),
  ),
  Component: withQueryRef(
    lazy(() => import("#/pages/organizations/vendors/VendorsPage")),
  ),
}

// In the component
export default function VendorsPage(props: Props) {
  const data = usePreloadedQuery(vendorsQuery, props.queryRef);
  // ...
}
```

- `loaderFromQueryLoader` — converts a query loader into a React Router loader, returns `{ queryRef, dispose }`
- `withQueryRef` — extracts `queryRef` from loader data and handles cleanup on unmount

For queries that need to run after render (e.g. select dropdowns), use `useLazyLoadQuery` with `fetchPolicy: "network-only"`.

## Fragments

Fragments colocate data requirements with the component that reads them:

```tsx
const contactFragment = graphql`
  fragment VendorContactsTabFragment_contact on VendorContact {
    id
    fullName
    email
    phone
    role
    createdAt
    updatedAt
    canUpdate: permission(action: "core:vendor-contact:update")
    canDelete: permission(action: "core:vendor-contact:delete")
  }
`;

function ContactRow(props: { contactKey: VendorContactsTabFragment_contact$key }) {
  const contact = useFragment(contactFragment, props.contactKey);
  // ...
}
```

### Refetchable fragments

For lists that support sorting and pagination, use `@refetchable` with `@argumentDefinitions`:

```tsx
const vendorContactsFragment = graphql`
  fragment VendorContactsTabFragment on Vendor
  @refetchable(queryName: "VendorContactsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorContactOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    contacts(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorContactsTabFragment_contacts") {
      __id
      edges {
        node {
          ...VendorContactsTabFragment_contact
        }
      }
    }
  }
`;

const [data, refetch] = useRefetchableFragment(vendorContactsFragment, vendor);
const connectionId = data.contacts.__id;
```

## Pagination

Use `usePaginationFragment` for cursor-based Relay pagination:

```tsx
const pagination = usePaginationFragment(paginatedVendorsFragment, data.node);
const vendors = pagination.data.vendors?.edges.map(edge => edge.node);
const connectionId = pagination.data.vendors.__id;
```

The `@connection(key: "...", filters: [...])` directive on the fragment tells Relay how to manage the paginated list in the store. The `filters` array controls which variables affect the connection identity.

`SortableTable` is the standard component for rendering paginated, sortable lists — it receives `pagination` (with `loadNext`, `hasNext`, `isLoadingNext`) and a `refetch` callback for sorting.

## Mutations

### `useMutation`

Direct Relay hook for simple cases:

```tsx
const [mutate] = useMutation<VendorGraphDeleteMutation>(deleteVendorMutation);
```

### `useMutationWithToasts`

Custom wrapper that adds toast notifications on success/error:

```tsx
const [createContact, isLoading] = useMutationWithToasts(
  createContactMutation,
  {
    successMessage: __("Contact created successfully."),
    errorMessage: __("Failed to create contact"),
  },
);

await createContact({
  variables: {
    input: { vendorId, ...cleanData },
    connections: [connectionId],
  },
  onSuccess: () => {
    dialogRef.current?.close();
    reset();
  },
});
```

### Store update directives

Relay directives handle connection updates automatically — no manual store manipulation needed:

```tsx
// Add new edge to the beginning of a connection
const createMutation = graphql`
  mutation CreateVendorMutation($input: CreateVendorInput!, $connections: [ID!]!) {
    createVendor(input: $input) {
      vendorEdge @prependEdge(connections: $connections) {
        node {
          id
          name
        }
      }
    }
  }
`;

// Remove an edge from a connection
const deleteMutation = graphql`
  mutation DeleteVendorMutation($input: DeleteVendorInput!, $connections: [ID!]!) {
    deleteVendor(input: $input) {
      deletedVendorId @deleteEdge(connections: $connections)
    }
  }
`;

// Update in-place via fragment spread (no directive needed)
const updateMutation = graphql`
  mutation UpdateContactMutation($input: UpdateVendorContactInput!) {
    updateVendorContact(input: $input) {
      vendorContact {
        ...VendorContactsTabFragment_contact
      }
    }
  }
`;
```

The `connections` variable is obtained from the `__id` field on the connection in the parent query/fragment.

### `useConfirm` for destructive actions

Destructive mutations (delete) are wrapped with a confirmation dialog:

```tsx
const confirm = useConfirm();

return () => {
  confirm(
    () =>
      promisifyMutation(mutate)({
        variables: {
          input: { vendorId: vendor.id! },
          connections: [connectionId],
        },
      }),
    { message: "Confirm deletion..." },
  );
};
```

## File organization

GraphQL operations are colocated with the components that use them:

```
pages/organizations/vendors/
  VendorsPage.tsx                    # query + pagination fragment
  tabs/
    VendorContactsTab.tsx            # refetchable fragment + item fragment
    VendorComplianceTab.tsx
  dialogs/
    CreateContactDialog.tsx          # create mutation
    EditContactDialog.tsx            # update mutation

hooks/graph/
  VendorGraph.ts                     # shared queries, mutations, hooks
```

Shared queries and mutation hooks (used by multiple components) live in `hooks/graph/*.ts`. Component-specific operations are defined inline in the component file.
