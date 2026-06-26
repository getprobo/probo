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

## Naming operations and fragments

Every operation (query, mutation, subscription) and fragment carries its **module-name prefix**: `<ModuleName><Type>` for operations, `<ModuleName>_<localName>` for fragments, where `<ModuleName>` is the file's basename.

```tsx
// PosterHovercard.tsx
query PosterHovercardQuery { ... }                 // operation: <ModuleName><Type>
fragment PosterHovercard_poster on Poster { ... }  // fragment:  <ModuleName>_<localName>
```

Relay 21 dropped the **compiler** requirement to prefix names with the filename for non-Haste projects (now opt-in via `enforce_module_name_prefix_for_non_haste`). We keep the prefix as mandatory house style:

- **Operations still require it regardless** — the `relay/graphql-naming` rule (from `eslint-plugin-relay`'s `ts-recommended`) reports any operation whose name doesn't start with the module name. Dropping it only for fragments would split the convention.
- **Uniqueness** — Relay still requires globally-unique operation/fragment names per project; the module prefix is the collision-free scheme that guarantees it.
- **Discoverability** — generated artifact filenames and the `$key` / `$data` types derive from the name, so `PosterHovercard_poster$key` points straight back to its source module.

We set `enforce_module_name_prefix_for_non_haste` in `relay.config.json` so the **compiler** guarantees the convention for fragments too — the lint rule only covers operations and legacy fragment containers, not hooks-based fragments.

`<localName>` is the data the key feeds (the prop minus its `Key` suffix), never a redundant `Fragment` word: a `contactKey` prop reads a `ContactListItem_contact` fragment.

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
  fragment ContactListItem_contact on ThirdPartyContact {
    id
    fullName
    email
    phone
    role
    createdAt
    updatedAt
    canUpdate: permission(action: "core:thirdParty-contact:update")
    canDelete: permission(action: "core:thirdParty-contact:delete")
  }
`;

function ContactListItem(props: { contactKey: ContactListItem_contact$key }) {
  const contact = useFragment(contactFragment, props.contactKey);
  // ...
}
```

Select `permission(action:)` fields (aliased `canUpdate` / `canDelete`) in the fragment of the component that renders the action, and gate the UI on the resulting boolean. See [`contrib/claude/permissions.md`](permissions.md).

### Required fields (`@required`)

GraphQL schemas mark many fields nullable defensively, but at a given call site you usually **expect** a value to be there. When a field is nullable in the schema but the component cannot meaningfully render without it, annotate it with `@required` so the **generated type is non-null**. This keeps typing honest and consistent: callers stop threading `?.` / `?? ""` / non-null `!` assertions through code that always expects data, and a genuinely-missing value surfaces as a real signal instead of silently rendering an empty UI.

Pick the action by what should happen when the value is actually absent at runtime:

- **`@required(action: THROW)`** — the value is an invariant for this view (e.g. a page's root entity, the `currentTrustCenter` a portal is built around). A null throws on read and propagates to the nearest error boundary (see [`error-handling.md`](error-handling.md)). The field becomes non-null in the type.
- **`@required(action: LOG)`** — a missing value should degrade gracefully rather than crash: the null **bubbles up** to the nearest `@required` ancestor (or makes the fragment/field data null), and Relay logs it. Use when the surrounding UI can render a sensible fallback.
- **`@required(action: NONE)`** — bubble nullability without logging; rarely needed.

```graphql
# Good — the view is built around this entity; THROW makes it non-null
currentTrustCenter @required(action: THROW) {
  organization {
    name          # already String! in the schema — no @required needed
    logo { downloadUrl }   # legitimately optional — leave nullable
  }
}
```

```tsx
// Good — non-null typing falls out of @required; no defensive chaining
const { organization } = data.currentTrustCenter;
const name = organization.name;
const logoUrl = organization.logo?.downloadUrl ?? undefined; // logo stays optional
```

Do **not** reach for `@required` to silence nullability on fields that are *genuinely* optional (an avatar, a logo, a description that may be empty). Those keep their nullable type and get a real empty/fallback state. Likewise, never select a field, mark it `@required(action: THROW)`, and rely on the throw as control flow for an expected-empty case — that is an error path, not a branch. And there is no need to annotate fields the schema already declares non-null (`String!`, `Organization!`).

### Refetchable fragments

For lists that support sorting and pagination, use `@refetchable` with `@argumentDefinitions`:

```tsx
const thirdPartyContactsFragment = graphql`
  fragment ThirdPartyContactsTabFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyContactsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyContactOrder", defaultValue: null }
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
    ) @connection(key: "ThirdPartyContactsTabFragment_contacts") {
      __id
      edges {
        node {
          ...ThirdPartyContactsTabFragment_contact
        }
      }
    }
  }
`;

const [data, refetch] = useRefetchableFragment(thirdPartyContactsFragment, thirdParty);
const connectionId = data.contacts.__id;
```

## Pagination

Use `usePaginationFragment` for cursor-based Relay pagination:

```tsx
const pagination = usePaginationFragment(paginatedThirdPartiesFragment, data.node);
const thirdParties = pagination.data.thirdParties?.edges.map(edge => edge.node);
const connectionId = pagination.data.thirdParties.__id;
```

The `@connection(key: "...", filters: [...])` directive on the fragment tells Relay how to manage the paginated list in the store. The `filters` array controls which variables affect the connection identity.

`SortableTable` is the standard component for rendering paginated, sortable lists — it receives `pagination` (with `loadNext`, `hasNext`, `isLoadingNext`) and a `refetch` callback for sorting.

## Mutations

Every mutation **must** update the Relay store so the UI reflects changes immediately — never rely on a page reload. Use `@appendEdge`/`@prependEdge` for creates, `@deleteEdge` for deletes, node `id` returns for in-place updates, and `updater` functions for complex multi-connection operations.

### `useMutation`

The project mutation primitive — awaitable, preserves every `UseMutationConfig` option, and routes success/error feedback through an injected notifier.

> Import `useMutation` from `#/lib/relay/useMutation`, never from `react-relay`. In compliance-portal this is enforced by a `no-restricted-imports` ESLint rule. See [`contrib/claude/hooks.md`](hooks.md#mutation-hooks).

#### Shared hook, app binding

The mechanics live in `@probo/relay` as `createUseMutation(useNotifier)` — a factory that wraps `react-relay`'s `useMutation` (promise wrapping, `onCompleted`/`onError` dispatch, `errorToast` semantics) but knows nothing about toasts or i18n. Each app binds it once to its own feedback stack via a `MutationNotifier` and re-exports the result as the canonical `useMutation`:

```tsx
// apps/compliance-portal/src/lib/relay/useMutation.ts — the only place feedback is wired
import { createUseMutation, type MutationNotifier } from "@probo/relay";

function useMutationNotifier(): MutationNotifier {
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  return useMemo<MutationNotifier>(() => ({
    notifySuccess: (title) => toast.add({ title, type: "success" }),
    notifyError: (error, title) => {
      const finalTitle = title ?? t("common.error");
      toast.add({ title: finalTitle, description: formatError(finalTitle, error as GraphQLError), type: "error" });
    },
  }), [toast, t]);
}

export const useMutation = createUseMutation(useMutationNotifier);
```

This keeps `@probo/relay` free of UI and i18n dependencies (the toast system, `react-i18next`, and `formatError` stay in the app), while the awaitable behavior is shared. Pass `MutationFeedback` (`successMessage`, `errorToast`) to control notifications without writing `onCompleted`/`onError` by hand.

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
const [deleteThirdParty] = useMutation<ThirdPartyGraphDeleteMutation>(deleteThirdPartyMutation);
```

For mutations with user feedback, queue a toast with Base UI's toast manager (`Toast.useToastManager()`; see [`ui.md`](ui.md#user-feedback-toasts)) from the `onCompleted` / `onError` callbacks:

```tsx
const toast = Toast.useToastManager();
const [createObligation, isCreating] = useMutation<CreateObligationMutation>(createObligationMutation);

const onSubmit = (input: ObligationInput) => {
  createObligation({
    variables: {
      input,
      connections: [connectionId],
    },
    onCompleted() {
      toast.add({
        title: t("obligations.created"),
        type: "success",
      });
    },
    onError(error) {
      toast.add({
        title: t("common.error"),
        description: formatError(t("obligations.createFailed"), error as GraphQLError),
        type: "error",
      });
    },
  });
};
```

### `useMutationWithToasts` (deprecated)

**Do not use.** Use `useMutation` and queue feedback with Base UI's toast manager (see [`ui.md`](ui.md#user-feedback-toasts)).

### `promisifyMutation` (deprecated)

**Do not use.** Use `useMutation` with `onCompleted`/`onError` callbacks instead of wrapping in a promise.

### Store update directives

Relay directives handle connection updates automatically — no manual store manipulation needed.

#### Connection setup

Any connection that a mutation will add to or remove from **must** have a `@connection` directive. If the mutation needs the connection ID in the same fragment, expose `__id`; otherwise derive it with `ConnectionHandler.getConnectionID`:

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
  mutation CreateThirdPartyMutation($input: CreateThirdPartyInput!, $connections: [ID!]!) {
    createThirdParty(input: $input) {
      thirdPartyEdge @prependEdge(connections: $connections) {
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
  mutation DeleteThirdPartyMutation($input: DeleteThirdPartyInput!, $connections: [ID!]!) {
    deleteThirdParty(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

// Update in-place (Relay matches by id — no directive needed)
const updateMutation = graphql`
  mutation UpdateContactMutation($input: UpdateThirdPartyContactInput!) {
    updateThirdPartyContact(input: $input) {
      thirdPartyContact {
        ...ThirdPartyContactsTabFragment_contact
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
const [deleteThirdParty] = useMutation<DeleteThirdPartyMutation>(deleteThirdPartyMutation);

return () => {
  confirm(
    () =>
      new Promise<void>((resolve) => {
        deleteThirdParty({
          variables: {
            input: { thirdPartyId: thirdParty.id! },
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
pages/organizations/third-parties/
  ThirdPartiesPage.tsx                    # query + pagination fragment
  _components/
    CreateContactDialog.tsx          # create mutation
    EditContactDialog.tsx            # update mutation
    ThirdPartyContactListItem.tsx    # connection-item fragment
  contacts/
    ThirdPartyContactsPage.tsx           # refetchable fragment + list section
    ThirdPartyContactsSection.tsx        # section fragment
```

Component-specific operations (queries, fragments, mutations) are defined inline in the component file that uses them. Shared sub-components live in `_components/` next to the page (scoped to the nearest common ancestor).