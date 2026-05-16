# Probo — TypeScript Frontend — @probo/helpers, @probo/hooks, @probo/i18n

The three "shared utilities" packages.

## @probo/helpers

Pure utility functions for string formatting, date/time, DOM, error formatting. **Universal
convention:** every helper returning user-facing text takes `__: Translator` as its **first
argument** (see [conventions.md § 5](../conventions.md#5-translator-injection)). This keeps the
package free of React-hook dependencies and unit-testable.

### Key files

- `packages/helpers/src/format/formatDate.ts`
- `packages/helpers/src/format/formatError.ts` — translates server error codes / messages to a
  localized user message.
- `packages/helpers/src/sprintf.ts`
- `packages/helpers/src/dom/faviconUrl.ts`
- `packages/helpers/src/index.ts` — **barrel export**; new files must be re-exported here.

### Known gap

The package's `tsconfig.json` does **not** enable `strict` mode. New files should still be
written as if strict were on. See [conventions.md § 11](../conventions.md#11-tsconfigper-workspace-no-root).

## @probo/hooks

**Nine** general-purpose hooks: `usePageTitle`, `useFavicon`, `useToggle`, `useList`, plus a
handful of others. Each hook is single-purpose and has no peer-dep state library — they use
`useState` / `useEffect` directly.

### Key files

- `packages/hooks/src/usePageTitle.ts` — sets `document.title`.
- `packages/hooks/src/useFavicon.ts` — swaps the favicon.
- `packages/hooks/src/useToggle.ts`, `packages/hooks/src/useList.ts` — generic state hooks.

### How to extend

Add `<HookName>.ts` and re-export from `packages/hooks/src/index.ts`. Hooks are good Vitest
candidates (see [testing.md § 5](../testing.md#5-hook--component-test-pattern)) — none have tests
today.

## @probo/i18n

Custom **zero-dependency** i18n library. Exports `TranslatorProvider`, `useTranslate`, and the
`Translator` type. **Currently dormant**:

- The loaders provided to `TranslatorProvider` return `{}`.
- Language is hard-coded `"en"` in app entry points.
- `__("My string")` is effectively `(s) => s`.

Continue wrapping every user-facing string in `__(...)` — when real translations land, this
infrastructure becomes live without code changes. **Don't strip `__` calls** because they look
like no-ops. See [pitfalls.md § 14](../pitfalls.md#14-packagesi18n-loaders-return---i18n-is-dormant).

## Top pitfalls

1. **Forgetting the helpers barrel update** —
   [pitfalls.md § 13](../pitfalls.md#13-forgetting-to-update-the-probohelpers-barrel-indexts).
2. **Removing dormant `__` calls** in app code —
   [pitfalls.md § 14](../pitfalls.md#14-packagesi18n-loaders-return---i18n-is-dormant).
