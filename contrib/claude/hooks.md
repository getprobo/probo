# Custom hooks

Custom hooks encapsulate reusable behavior (data wiring, derived state, event logic). This guide covers **where they live**, **how they are named**, and the **mutation hook pattern** — an awaitable wrapper over Relay's `useMutation` that preserves every option and automates error feedback.

## Related guides

| Topic | Guide |
|-------|--------|
| `_lib` / `_components` placement, hoisting | [`contrib/claude/app-arborescence.md`](app-arborescence.md) |
| Mutations, store updates, connection directives | [`contrib/claude/relay.md`](relay.md) |
| Toasts, user feedback | [`contrib/claude/ui.md`](ui.md#user-feedback-toasts) |
| i18next translation keys | [`contrib/claude/i18n.md`](i18n.md) |

## Placement

- **Feature-scoped hooks** live in the feature's `_lib/` folder, next to the pages that use them (`pages/organizations/measures/_lib/useDeleteMeasure.ts`).
- **Shared hooks** used across features are hoisted to the **nearest common ancestor's** `_lib/`, and only **app-wide** primitives (used everywhere) live in the top-level `src/lib/`.
- Promote a hook **when a second feature needs it**, not preemptively — the same rule as `_components/` (see [`app-arborescence.md`](app-arborescence.md)).

```text
// Feature-scoped
pages/organizations/measures/_lib/useDeleteMeasure.ts

// Shared across a feature area
pages/organizations/_lib/useOrganizationId.ts

// App-wide primitive
src/lib/relay/useMutation.ts
```

## Shape and naming

- **One primary hook per file**; the file is camelCase and matches the hook name (`useDeleteMeasure.ts` → `useDeleteMeasure`).
- Hooks are **`function` declarations**, named `use…` (see [`react-components.md`](react-components.md#component-shape)).
- Colocate a hook's `graphql` operation in the same file.
- A hook returns either a value, or a tuple when it mirrors a React/Relay primitive (`[commit, isInFlight]`).

## Mutation hooks

All mutations go through the shared **`useMutation`** primitive. Its mechanics live in `@probo/relay` as the `createUseMutation(useNotifier)` factory, and each app binds it once in `src/lib/relay/useMutation.ts`. It wraps Relay's `useMutation` to:

1. Return an **awaitable** commit that resolves with the mutation **response** (so callers can `await` and continue only on success).
2. **Preserve every `UseMutationConfig` option** (`variables`, `connections`, `updater`, `optimisticResponse`, `onCompleted`, `onError`, …) by spreading the caller's config.
3. **Automate feedback**: on failure it notifies (via the app's injected `MutationNotifier` — Base UI toast + `formatError`) **and** rejects the promise — controllable per call through a `MutationFeedback` options object.

### Always import `useMutation` from `#/lib/relay/useMutation`

Our `useMutation` intentionally shadows `react-relay`'s. **Import it only from `#/lib/relay/useMutation`; never import `useMutation` from `react-relay` directly** (enforced in compliance-portal by a `no-restricted-imports` ESLint rule). This guarantees one consistent entrypoint with awaitable results and automatic error handling everywhere.

```ts
// Bad — raw Relay hook (no await, no auto error handling)
import { useMutation } from "react-relay";

// Good — the project primitive
import { useMutation } from "#/lib/relay/useMutation";
```

### The primitive: shared factory + app binding

The factory lives in `@probo/relay` and stays free of UI/i18n dependencies — it delegates rendering to an injected `MutationNotifier` (`createUseMutation` source: [`packages/relay/src/useMutation.ts`](../../packages/relay/src/useMutation.ts)). The app binds it once to its own toast + i18n + `formatError` stack:

```ts
// src/lib/relay/useMutation.ts — the only place feedback is wired
import { Toast } from "@base-ui/react/toast";
import { formatError, type GraphQLError } from "@probo/helpers";
import { createUseMutation, type MutationNotifier } from "@probo/relay";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";

function useMutationNotifier(): MutationNotifier {
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  return useMemo<MutationNotifier>(
    () => ({
      notifySuccess: (title) => toast.add({ title, type: "success" }),
      notifyError: (error, title) => {
        const finalTitle = title ?? t("common.error");
        toast.add({
          title: finalTitle,
          description: formatError(finalTitle, error as GraphQLError),
          type: "error",
        });
      },
    }),
    [toast, t],
  );
}

export type { MutationFeedback } from "@probo/relay";

export const useMutation = createUseMutation(useMutationNotifier);
```

### Domain mutation hook (colocated `_lib/`)

A feature hook wraps the primitive with its operation and default feedback. Name the hook after the action; name the destructured commit function after the tagged node minus `Mutation` (see [`relay.md`](relay.md#naming-convention)).

```ts
// pages/organizations/measures/_lib/useDeleteMeasure.ts
import { graphql } from "relay-runtime";
import { useTranslation } from "react-i18next";

import { useMutation } from "#/lib/relay/useMutation";
import type { DeleteMeasureMutation } from "#/__generated__/core/DeleteMeasureMutation.graphql";

const deleteMeasureMutation = graphql`
  mutation DeleteMeasureMutation($input: DeleteMeasureInput!, $connections: [ID!]!) {
    deleteMeasure(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteMeasure() {
  const { t } = useTranslation();
  return useMutation<DeleteMeasureMutation>(deleteMeasureMutation, {
    successMessage: t("measures.deleted"),
    errorToast: t("measures.deleteFailed"),
  });
}
```

### Usage — await the result

```tsx
const [deleteMeasure, isDeleting] = useDeleteMeasure();

// Default: awaits the response; on failure it toasts AND throws.
async function onConfirm() {
  await deleteMeasure({ variables: { input: { measureId }, connections: [connectionId] } });
  navigate(".."); // only runs on success
}

// Opt out of the auto-toast to handle the error yourself:
try {
  const result = await deleteMeasure({ variables }, { errorToast: false });
  // use result.deleteMeasure.deletedMeasureId …
} catch (error) {
  // custom handling
}
```

### Do / don't

```text
// Bad — legacy wrappers (removed in v2)
useMutationWithToasts(...)     // resolves to void, loses the response; v1 toast + __
useMutationWithIncrement(...)  // callback-style, not awaitable
promisifyMutation(commit)      // standalone wrapper, disconnected from the hook

// Bad — importing the raw Relay hook
import { useMutation } from "react-relay";

// Good — one primitive, awaitable, options preserved, auto error handling
import { useMutation } from "#/lib/relay/useMutation";
```

Prefer the declarative store directives (`@appendEdge`, `@deleteEdge`) and the patterns in [`relay.md`](relay.md) for *what* the mutation does to the store; this hook only governs *how* it is invoked and how its errors surface.
