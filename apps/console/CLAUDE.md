# apps/console

React 19 + Vite + TypeScript + Relay + TailwindCSS. Port 5173.

## Commands

| Command | Purpose |
|---------|---------|
| `npm run dev` | Start dev server (port 5173) |
| `npm run build` | Production build |
| `npx relay-compiler` | Regenerate Relay artifacts |

## Routes

Defined in `src/routes.tsx` with feature-specific route files (e.g. `src/routes/assetRoutes.ts`).

- Lazy-loaded via `lazy()` from `@probo/react-lazy`
- Data loading: `loaderFromQueryLoader()` + `loadQuery()` (Relay)
- Type: all routes `satisfies AppRoute[]`
- Fallback: `PageSkeleton` or `Fallback` components during loading
- Error boundaries per route group

## Relay

Queries, fragments, and mutations are **colocated** in the component that uses them — never in separate `hooks/graph/` files.

Always use **fragments** to define data requirements. Never create custom TypeScript types for API data — let Relay generate types from fragments.

### Mutations

Use `@appendEdge` / `@deleteEdge` directives for Relay store updates:

```typescript
const createAssetMutation = graphql`
  mutation AssetCreateMutation($input: CreateAssetInput!, $connections: [ID!]!) {
    createAsset(input: $input) {
      assetEdge @appendEdge(connections: $connections) {
        node { id name }
      }
    }
  }
`;
```

## Permissions

Inline permission queries in Relay fragments:

```graphql
canCreate: permission(action: "core:asset:create")
```

## Components

- Form fields: `src/components/form/`
- Dialogs: modal components with mutation handling
- Shared UI: `@probo/ui` package
