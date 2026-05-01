# Probo — TypeScript Frontend — @probo/ui

The shared component library (**285 files**) consumed by `apps/console` and `apps/trust`. Folder
layout under `packages/ui/src/`: **`atoms/`** (Button, Input, Field, Badge, Avatar, Icon, Label,
Slot, Skeleton…) and **`molecules/`** (Card, Dialog, Dropdown, Select, Combobox, Tabs, Table,
PageHeader, Toast, Confirm, Popover, Tooltip…). Styling is **`tailwind-variants` `tv()`**, sole
API. Compound components export **flat named members** (`DialogRoot`, `DialogShell`, `DialogTitle`,
`DialogSkeleton`) — never namespace-style `Dialog.Root`. Radix UI backs Dialog/Dropdown/Select/
Tabs/Label/Popover; Ariakit backs Combobox; `cmdk` powers the command palette.

The custom `Slot` (`packages/ui/src/atoms/Slot.tsx`) implements the `asChild` pattern by merging
props/refs onto a single child — used by Button, Link, etc.

## Key files

- `packages/ui/src/atoms/Slot.tsx` — custom `asChild` Slot.
- `packages/ui/src/atoms/Button/` — modern compound shape (variants in `variants.ts`).
- `packages/ui/src/atoms/Button/Button.tsx` — **legacy** `export const + arrow` holdout; pattern
  to **avoid** copying.
- `packages/ui/src/molecules/Toast/` — `useToast` consumed by the apps for user feedback.
- `packages/ui/src/molecules/Confirm/` — `useConfirm` for destructive actions.

## How to extend

1. Pick the right folder (`atoms/` for primitives, `molecules/` for compositions).
2. Create `<Component>.tsx` (function declaration), `variants.ts` (`tv()` variants),
   `<Component>.stories.tsx` (CSF 3), and a co-located skeleton if the component has loading state.
3. Export from the package barrel (`packages/ui/src/index.ts`).
4. Skeleton must NOT import the `*Root` — both pull from `variants.ts`.

## Top pitfalls

1. **Legacy `Button.tsx` shape** — do not copy `export const Button = forwardRef(...)`; use
   `export function`. See [pitfalls.md § 9](../pitfalls.md#9-packagesui-legacy-buttontsx-uses-export-const--arrow).
2. **Skeleton importing Root** — see [pitfalls.md § 10](../pitfalls.md#10-skeleton-importing-root).
