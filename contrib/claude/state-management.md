# Client state management

Probo frontends have several places state can live. Choosing the wrong one is the most common source of avoidable complexity: data drilled through props, local state that should have been in the URL, or a global store holding what is really server data. This guide is a **decision order** — start at the top and stop at the first option that fits.

## Related guides

| Topic | Guide |
|-------|--------|
| Server data (queries, fragments, mutations, store) | [`contrib/claude/relay.md`](relay.md) |
| URL/search params, navigation | [`contrib/claude/routing.md`](routing.md) |
| Props are configuration, not data | [`contrib/claude/react-components.md`](react-components.md#props-are-for-configuration-and-composition-not-data) |

## Decision order

1. **Server data → Relay.** Anything fetched from the API lives in the Relay store, read via `useFragment` / `usePreloadedQuery`, and mutated with `useMutation` that updates the store. Never copy server data into `useState` to "manage" it. (See [`relay.md`](relay.md).)
2. **Shareable / linkable / reload-surviving UI state → the URL.** Active tab, filters, sort, search query, pagination cursor, selected id in a master-detail view. Use `useSearchParams` / route params (see [`routing.md`](routing.md)). If a teammate should be able to paste the link and see the same view, it belongs here.
3. **Local component state → `useState` / `useReducer`.** Ephemeral, view-only state scoped to one component or a small subtree: an input's draft value, whether a menu is open, a hover flag. Keep it as local as possible.
4. **Shared ephemeral state across a subtree → React context.** When a handful of nearby components need the same ephemeral state and lifting to a common parent is clean (e.g. a wizard step, a selection set within a list). Context carries the state + setters; it does **not** carry server data.
5. **App-wide ephemeral client state → `zustand`.** Only for genuinely global, non-server, non-URL state that many unrelated parts of the app read/write: theme / dark-mode preference, a command palette, global toasts. Reach for it last.

```mermaid
flowchart TD
  start[Need to hold some state] --> server{From the API?}
  server -->|Yes| relay[Relay store]
  server -->|No| url{"Shareable / linkable / survive reload?"}
  url -->|Yes| search[URL: useSearchParams or route params]
  url -->|No| scope{Used by one component or a small subtree?}
  scope -->|One subtree| local[useState / useReducer]
  scope -->|A few nearby components| ctx[React context]
  scope -->|Many unrelated places| zustand[zustand store]
```

## Do / don't

### Server data stays in Relay

```tsx
// Bad — copying fetched data into local state to "edit" it
const data = usePreloadedQuery(query, queryRef);
const [measures, setMeasures] = useState(data.measures);

// Good — read from Relay; mutate through useMutation (store updates flow back)
const data = usePreloadedQuery(query, queryRef);
```

### Tab / filter belong in the URL

```tsx
// Bad — active tab in local state; lost on reload, not linkable
const [tab, setTab] = useState<"open" | "closed">("open");

// Good — tab in the URL
const [params, setParams] = useSearchParams();
const tab = params.get("tab") ?? "open";
```

### Don't reach for a global store by default

```tsx
// Bad — a zustand store for state two sibling components share
useFilterStore();   // global singleton for a local concern

// Good — lift to the nearest common parent (state + context if the tree is deep)
```

`zustand` is available (it's a `@probo/ui` dependency) — use it deliberately for app-wide ephemeral state, not as a shortcut around prop-passing or the URL.

## Filtering a list

A filterable list (a toolbar of selects + a search box driving a server-side
query) combines several of the rules above. Three points matter, in order:

1. **The filter values live in the URL** (rule 2): category, status, sort, and
   the search term go in `useSearchParams`. Expose them through a hook that is
   **pure** — it reads the params and returns values + setters, with **no
   `useState` and no `useEffect`**. A pure hook is safe to call from every
   component that needs the filters (page, loader, toolbar, empty state); they
   all read one source of truth.

   The failure mode: a URL-filter hook that keeps its own `useState` mirror
   **and** an effect writing that mirror back to the URL. Every caller runs its
   own copy of that effect, but only one (the input) updates the mirror — the
   rest keep a **stale** mirror and fight the real writer, flipping the list
   between filtered and unfiltered in an infinite loop.

2. **The debounced search input is the one piece of local state** — and it lives
   in a **separate, single-owner hook** mounted in exactly one component (the
   toolbar), the single writer of the search param. It holds the immediate input
   value and commits it to the URL after a debounce. Guard the URL→input sync
   with a `ref` tracking your own writes, so a debounced commit isn't echoed
   back onto the input (which would drop in-flight keystrokes); only *external*
   changes (a Clear button, back/forward) sync.

3. **Refetch on change inside a transition** so the loading state is scoped to
   the results, not the whole page — see
   [`relay.md`](relay.md#refetch-inside-a-transition).

```ts
// Bad — URL-filter hook with a per-instance mirror + write-back effect.
// Called by the page, loader, toolbar and empty state → four fighting writers.
export function useThirdPartyFilters() {
  const [params, setParams] = useSearchParams();
  const query = params.get("q") ?? "";
  const [input, setInput] = useState(query);
  useEffect(() => {
    if (input !== query) setParams(/* write q=input */, { replace: true });
  }, [input, query, setParams]);
  return { query, input, setInput /* … */ };
}
```

```ts
// Good — pure URL-state hook (no state/effects); safe to call anywhere
export function useThirdPartyFilters() {
  const [params, setParams] = useSearchParams();
  const setParam = useCallback((k: string, v: string) => {
    setParams((prev) => {
      const next = new URLSearchParams(prev);
      if (v) next.set(k, v); else next.delete(k);
      return next;
    }, { replace: true });
  }, [setParams]);
  return {
    query: params.get("q") ?? "",
    category: params.get("category") ?? "",
    setQuery: (v: string) => setParam("q", v),
    setCategory: (v: string) => setParam("category", v),
    clear: () => setParams({}, { replace: true }),
  };
}

// Good — debounced search input in its own single-owner hook (used by the toolbar)
export function useThirdPartySearch(): [string, (v: string) => void] {
  const { query, setQuery } = useThirdPartyFilters();
  const [input, setInput] = useState(query);
  const lastCommitted = useRef(query);
  useEffect(() => {                          // debounce input → URL
    if (input === query) return;
    const h = setTimeout(() => { lastCommitted.current = input; setQuery(input); }, 300);
    return () => clearTimeout(h);
  }, [input, query, setQuery]);
  useEffect(() => {                          // sync URL → input for external changes only
    if (query !== lastCommitted.current) { lastCommitted.current = query; setInput(query); }
  }, [query]);
  return [input, setInput];
}
```

## Anti-patterns to avoid

- **Prop-drilling data** that a child could read from Relay (`useFragment`) or the router (`useParams`) itself.
- **Mirroring** a controlled value (e.g. a dialog's `open`, a fragment field) into `useState` + `useEffect` to keep them in sync — pass the source through instead.
- **Global store as a cache** for server data — that's Relay's job.
- **Outlet context for domain data** — forbidden; see [`routing.md`](routing.md).
