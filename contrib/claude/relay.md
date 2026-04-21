# Relay (Frontend GraphQL Client)

The console app uses [Relay](https://relay.dev/) as its GraphQL client. All GraphQL operations are defined inline with the `graphql` template tag from `relay-runtime` — there are no separate `.graphql` files on the frontend.

## Environments

Two Relay environments connect to two separate GraphQL APIs:


| Environment       | Endpoint                  | Purpose                   |
| ----------------- | ------------------------- | ------------------------- |
| `coreEnvironment` | `/api/console/v1/graphql` | Main application data     |
| `iamEnvironment`  | `/api/connect/v1/graphql` | Authentication / identity |


Configured in `apps/console/src/environments.ts`. Each has its own store with 1-minute query cache expiration.

## Relay compiler

Config lives in `relay.config.json` at the repo root with three projects (`core`, `iam`, `trust`) mapped to different source directories and schemas. Each project uses `schema` pointing to `base.graphql` and `schemaExtensions` pointing to the `graphql/` directory containing the per-entity schema files. Generated files go into `__generated__/` directories.

```sh
make relay  # merge split schemas + clean + compile
```

Custom scalar mappings: `Datetime → string`, `GID → string`, `CursorKey → string`, `Duration → string`, `BigInt → number`, `EmailAddr → string`.

## Colocated queries

Queries are defined inline in the file that uses them. Route-level queries are preloaded in a dedicated `*PageLoader` component before the page renders.

### Route definition

Routes only declare `path`, `Fallback`, and `Component` pointing to a lazy-loaded loader component — no Relay logic in the route itself:

```tsx
// In route file (e.g. findingRoutes.ts)
import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const findingRoutes = [
  {
    path: "findings",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/findings/FindingsPageLoader"),
    ),
  },
] satisfies AppRoute[];
```

### Loader component

The loader component owns the Relay query lifecycle — it calls `useQueryLoader` + `useEffect` to preload, renders a skeleton while waiting, then wraps the real page in `Suspense`:

```tsx
// FindingsPageLoader.tsx
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { FindingsPageListQuery } from "#/__generated__/core/FindingsPageListQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import FindingsPage, { findingsPageQuery } from "./FindingsPage";

export default function FindingsPageLoader() {
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<FindingsPageListQuery>(findingsPageQuery);

  useEffect(() => {
    loadQuery({
      organizationId,
      snapshotId: snapshotId ?? null,
    });
  }, [loadQuery, organizationId, snapshotId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <FindingsPage queryRef={queryRef} />
    </Suspense>
  );
}
```

### Page component

The page receives `queryRef` as a prop and reads data with `usePreloadedQuery`:

```tsx
// FindingsPage.tsx
export const findingsPageQuery = graphql`
  query FindingsPageListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...FindingsPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

interface FindingsPageProps {
  queryRef: PreloadedQuery<FindingsPageListQuery>;
};

export default function FindingsPage({ queryRef }: FindingsPageProps) {
  const data = usePreloadedQuery(findingsPageQuery, queryRef);
  // ...
}
```

### `loaderFromQueryLoader` / `withQueryRef` (deprecated)

**Do not use.** Use a `*PageLoader` component with `useQueryLoader` as shown above instead.

## Interaction-triggered queries

When a user interaction (hover, click, open dialog) needs data beyond what the initial page query loaded, use a secondary query with `useQueryLoader` + `usePreloadedQuery`. This starts fetching in the event handler — before the target component renders — so the network request and component rendering overlap instead of running sequentially.

The parent component owns the query lifecycle with `useQueryLoader`, triggers the fetch in the event handler, and passes the query ref down:

```tsx
import { Suspense } from "react";
import { useQueryLoader } from "react-relay";

import type { PosterHovercardQuery as HovercardQueryType } from "#/__generated__/core/PosterHovercardQuery.graphql";

import PosterHovercard, { posterHovercardQuery } from "./PosterHovercard";

function PosterByline({ poster }: Props) {
  const data = useFragment(posterBylineFragment, poster);
  const [hovercardQueryRef, loadHovercardQuery] =
    useQueryLoader<HovercardQueryType>(posterHovercardQuery);

  function onBeginHover() {
    loadHovercardQuery({ posterId: data.id });
  }

  return (
    <HoverTrigger onBeginHover={onBeginHover}>
      {hovercardQueryRef && (
        <Suspense fallback={<Spinner />}>
          <PosterHovercard queryRef={hovercardQueryRef} />
        </Suspense>
      )}
    </HoverTrigger>
  );
}
```

The child component reads data with `usePreloadedQuery`:

```tsx
import { graphql, usePreloadedQuery } from "react-relay";
import type { PreloadedQuery } from "react-relay";
import type { PosterHovercardQuery } from "#/__generated__/core/PosterHovercardQuery.graphql";

export const posterHovercardQuery = graphql`
  query PosterHovercardQuery($posterId: ID!) {
    node(id: $posterId) {
      ... on Poster {
        ...PosterHovercardBodyFragment
      }
    }
  }
`;

interface PosterHovercardProps {
  queryRef: PreloadedQuery<PosterHovercardQuery>;
}

export default function PosterHovercard({ queryRef }: PosterHovercardProps) {
  const data = usePreloadedQuery(posterHovercardQuery, queryRef);
  // ...
}
```

**Do not use `useLazyLoadQuery`** — it defers the fetch until the component renders, adding unnecessary latency. Always prefer `useQueryLoader` + `usePreloadedQuery` so the network request starts in the event handler.

## Fragments

Fragments colocate data requirements with the component that reads them:

```tsx
const contactFragment = graphql`
  fragment ContactRow_contactFragment on VendorContact {
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

function ContactRow(props: { contactKey: ContactRow_contactFragment$key }) {
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

Every mutation **must** update the Relay store so the UI reflects changes immediately — never rely on a page reload. Use `@appendEdge`/`@prependEdge` for creates, `@deleteEdge` for deletes, node `id` returns for in-place updates, and `updater` functions for complex multi-connection operations.

### `useMutation`

Direct Relay hook for simple cases.

#### Naming convention

Name the destructured result of `useMutation` after the **graphql tagged-template variable**, dropping the `Mutation` suffix:

| Tagged node variable | Commit function | In-flight boolean |
|----------------------|-----------------|-------------------|
| `createCookieBannerMutation` | `createCookieBanner` | `isCreating` or `isCreatingCookieBanner` |
| `updateBannerMutation` | `updateBanner` | `isUpdating` |
| `deleteCategoryMutation` | `deleteCategory` | `isDeleting` |
| `activateMutation` | `activate` | `isActivating` |

**Never** use generic names like `commitMutation`, `commit`, or `isInFlight`.

```tsx
// Bad
const [commitMutation, isInFlight] = useMutation<Mutation>(createCookieBannerMutation);
commitMutation({ variables: { ... } });

// Good
const [createCookieBanner, isCreating] = useMutation<Mutation>(createCookieBannerMutation);
createCookieBanner({ variables: { ... } });
```

#### Examples

```tsx
const [deleteVendor] = useMutation<VendorGraphDeleteMutation>(deleteVendorMutation);
```

For mutations with user feedback, combine with `useToast` and use `onCompleted`/`onError` callbacks:

```tsx
const { toast } = useToast();
const [createObligation, isCreating] = useMutation<CreateObligationMutation>(createObligationMutation);

const onSubmit = (formData: FormData) => {
  createObligation({
    variables: {
      input: { ...formData },
      connections: [connectionId],
    },
    onCompleted() {
      toast({
        title: __("Success"),
        description: __("Obligation created successfully"),
        variant: "success",
      });
    },
    onError(error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to create obligation"), error as GraphQLError),
        variant: "error",
      });
    },
  });
};
```

### `useMutationWithToasts` (deprecated)

**Do not use.** Use `useMutation` combined with `useToast` instead.

### `promisifyMutation` (deprecated)

**Do not use.** Use `useMutation` with `onCompleted`/`onError` callbacks instead of wrapping in a promise.

### Store update directives

Relay directives handle connection updates automatically — no manual store manipulation needed.

#### Connection setup

Any connection that a mutation will add to or remove from **must** have a `@connection` directive and expose `__id`:

```tsx
const fragment = graphql`
  fragment CategorySectionFragment on CookieCategory {
    id
    cookies(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
      @connection(key: "CategorySection_cookies", filters: [])
      @required(action: THROW) {
      __id
      edges {
        node {
          id
          ...EditCookieRowFragment
        }
      }
    }
  }
`;

const category = useFragment(fragment, categoryKey);
const connectionId = category.cookies.__id;
```

#### `filters` on `@connection`

By default Relay treats every non-pagination argument (`first`, `last`, `after`, `before` are excluded) as a **filter** and encodes its value into the connection's store identity. This means `ConnectionHandler.getConnection(record, key)` will fail to find the connection unless the exact same filter values are passed as a third argument.

**Use `filters: []`** when the connection has fixed arguments (e.g. a hardcoded `orderBy`) and there is only ever one instance of the connection per parent node. This is the common case — it keeps `ConnectionHandler.getConnection` and `getConnectionID` simple:

```graphql
# Good — single fixed ordering, no filtered variants
cookies(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
  @connection(key: "CategorySection_cookies", filters: [])
```

**List specific filter arguments** when the same connection is rendered with different filter values and you need Relay to maintain separate lists in the store (e.g. a table with user-selectable sorting or status filters):

```graphql
# Good — user can change the status filter, each variant is a separate list
tasks(first: 50, status: $status, orderBy: $order)
  @connection(key: "TaskList_tasks", filters: ["status"])
```

When `filters` includes an argument, `ConnectionHandler.getConnection` requires matching filter values:

```tsx
const conn = ConnectionHandler.getConnection(record, "TaskList_tasks", {
  status: "OPEN",
});
```

**Never omit `filters`** — rely on the explicit list rather than Relay's default (all non-pagination args), which silently breaks `ConnectionHandler` lookups and `updater` functions.

#### Connection ID from outside the subtree

When the mutation is triggered from a component that doesn't have access to the connection's `__id` (e.g. a sibling's child rather than a direct descendant), derive the connection ID with `ConnectionHandler.getConnectionID`:

```tsx
import { ConnectionHandler } from "relay-runtime";

const connectionId = ConnectionHandler.getConnectionID(
  parentNodeId,           // the store ID of the node that owns the connection
  "CategorySection_cookies", // the @connection key
);
```

This is useful for dialogs, drawers, or other components rendered outside the subtree that reads the connection.

#### Directive examples

```tsx
// Add new edge to a connection
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

// Update in-place (Relay matches by id — no directive needed)
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

#### Fragment spreads in create mutations

When a create mutation returns a new edge, its `node` selection **must** include all fragment spreads used by the list that renders it. This ensures the store has every field the UI needs to render the new item without a refetch:

```tsx
// Bad — missing fragment spread, child components will have missing data
cookieEdge @appendEdge(connections: $connections) {
  node { id name duration description }
}

// Good — spreads the same fragment the list uses to render each item
cookieEdge @appendEdge(connections: $connections) {
  node { id name duration description ...EditCookieRowFragment }
}
```

#### `updater` for complex store changes

When a single mutation affects multiple connections (e.g. moving an item between two lists) and the server payload doesn't return both an edge and a deletedId, use an `updater` function with `ConnectionHandler`:

```tsx
import { ConnectionHandler } from "relay-runtime";

moveCookie({
  variables: { input: { cookieId, targetCookieCategoryId: targetId } },
  updater(store) {
    const source = store.get(sourceCategoryId);
    if (source) {
      const sourceConn = ConnectionHandler.getConnection(source, "CategorySection_cookies");
      if (sourceConn) ConnectionHandler.deleteNode(sourceConn, cookieId);
    }

    const target = store.get(targetId);
    if (target) {
      const targetConn = ConnectionHandler.getConnection(target, "CategorySection_cookies");
      if (targetConn) {
        const node = store.get(cookieId);
        if (node) {
          const edge = ConnectionHandler.createEdge(store, targetConn, node, "CookieEdge");
          ConnectionHandler.insertEdgeAfter(targetConn, edge);
        }
      }
    }
  },
});
```

Prefer declarative directives (`@appendEdge`, `@deleteEdge`) whenever possible; only fall back to `updater` when the operation cannot be expressed with directives alone.

### `useConfirm` for destructive actions

Destructive mutations (delete) are wrapped with a confirmation dialog:

```tsx
const confirm = useConfirm();
const [deleteVendor] = useMutation<DeleteVendorMutation>(deleteVendorMutation);

return () => {
  confirm(
    () =>
      new Promise<void>((resolve) => {
        deleteVendor({
          variables: {
            input: { vendorId: vendor.id! },
            connections: [connectionId],
          },
          onCompleted() {
            resolve();
          },
          onError() {
            resolve();
          },
        });
      }),
    { message: "Confirm deletion..." },
  );
};
```

## File organization

GraphQL operations are colocated with the components that use them. See [`contrib/claude/app-arborescence.md`](app-arborescence.md) for the full folder layout.

```
pages/organizations/vendors/
  VendorsPage.tsx                    # query + pagination fragment
  _components/
    CreateContactDialog.tsx          # create mutation
    EditContactDialog.tsx            # update mutation
  tabs/
    VendorContactsTab.tsx            # refetchable fragment + item fragment
    VendorComplianceTab.tsx
```

Component-specific operations (queries, fragments, mutations) are defined inline in the component file that uses them. Shared sub-components live in `_components/` next to the page (scoped to the nearest common ancestor).