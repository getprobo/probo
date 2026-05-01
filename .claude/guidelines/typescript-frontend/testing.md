# Probo — TypeScript Frontend — Testing

> Cross-cutting CI gates and the four-surface API rule live in [shared.md](../shared.md).
> Frontend testing is significantly **less mature** than backend testing — much of this file
> documents gaps as much as positive guidance.

---

## 1. Frameworks & Tooling

| Tool | Where used |
| --- | --- |
| **Vitest** | Unit / integration runner for all packages and both apps. `vitest` is in the workspace toolchain root. |
| **@testing-library/react** | Component testing in `@probo/ui` and (where present) the apps. |
| **Storybook 8** | `@probo/ui` interactive component catalog and visual stories. |
| **Playwright** | NOT used for frontend. End-to-end is Go-driven (`e2e/console`, `e2e/mcp`); see [shared.md § 7](../shared.md#7-ci--quality-gates). |

`make test` runs **Go** unit tests only. There is no `make test-frontend` target. To run frontend
tests today, drop into the workspace and run `npm test` (Vitest) or
`npx turbo run test --filter=@probo/ui`.

---

## 2. Current State — Coverage Gaps

> Universal observation across `apps/console`, `apps/trust`, and most `packages/*`: tests are
> **scarce or absent** despite Vitest being configured.

| Module | Test status |
| --- | --- |
| `apps/console` (437 files) | No test files found. |
| `apps/trust` (50 files) | No test files found. |
| `packages/ui` | Partial Storybook coverage; Vitest tests for some atoms only. |
| `packages/relay` | Some unit tests on the error-classifier path. |
| `packages/helpers` | No tests despite Vitest configured. |
| `packages/hooks` | No tests despite Vitest configured. |
| `packages/i18n` | No tests; library is dormant anyway. |
| `packages/prosemirror` | No tests despite Vitest configured. |
| `packages/emails` | No tests; templates are validated only by the build step. |
| `packages/n8n-node` | `npx n8n-node lint` is the only gate; no unit tests. |
| `packages/cookie-banner` | No tests. |
| `packages/react-lazy` | No tests. |
| `packages/vendors` | Static data; no tests. |

**For new code, prefer to write a Vitest test** — even a trivial smoke test starts the trend in
the right direction, and the runner is already wired.

> See [shared.md § 13 #11](../shared.md#13-code-review-enforced-standards) for the
> security-package 100% coverage rule (Go-side: OAuth2/OIDC/PKCE). The frontend has no equivalent
> mandatory coverage at the moment, but reviewers may flag missing tests for security-sensitive
> flows (e.g. NDA flow, magic-link redirect handling in `apps/trust`).

---

## 3. Test File Organization

When you do write tests, **co-locate** them with the source:

```
packages/helpers/src/format/formatDate.ts
packages/helpers/src/format/formatDate.test.ts
```

Use `.test.ts(x)` (Vitest's default glob). Avoid `__tests__/` subfolders for new code.

For `@probo/ui` Storybook stories, name `<Component>.stories.tsx` next to the component:

```
packages/ui/src/atoms/Button/Button.tsx
packages/ui/src/atoms/Button/Button.stories.tsx
```

Stories use **CSF 3** format:

```tsx
import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "./Button";

const meta = { component: Button } satisfies Meta<typeof Button>;
export default meta;

type Story = StoryObj<typeof meta>;

export const Primary: Story = { args: { variant: "primary", children: "Save" } };
export const Disabled: Story = { args: { variant: "primary", disabled: true, children: "Save" } };
```

> **Reviewer expectation:** new `@probo/ui` components should ship with a Storybook story. (Drawn
> from PR-mining patterns around the UI package.)

---

## 4. Test Naming Conventions

```ts
import { describe, it, expect } from "vitest";

describe("formatError", () => {
  it("returns the translated message for a known GraphQL code", () => { ... });

  it("falls back to a generic internal error message for unknown codes", () => { ... });
});
```

- Top-level `describe` named after the unit under test (function, hook, component).
- `it("…")` reads as a sentence completing *"It …"*. No `should` prefix.
- One assertion per concept; multiple `expect()` calls inside a single `it` is fine.

For React components testing user behaviour, follow Testing Library's user-centric query order
(`getByRole` → `getByLabelText` → `getByText` → `getByTestId` only as last resort).

---

## 5. Hook & Component Test Pattern

```tsx
// packages/hooks/src/useToggle.test.ts
import { renderHook, act } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { useToggle } from "./useToggle";

describe("useToggle", () => {
  it("flips the boolean on toggle()", () => {
    const { result } = renderHook(() => useToggle(false));
    expect(result.current[0]).toBe(false);
    act(() => result.current[1]());
    expect(result.current[0]).toBe(true);
  });
});
```

For Relay components, wrap in a `RelayEnvironmentProvider` with a
`createMockEnvironment()` from `relay-test-utils`. (No fixtures exist in the codebase yet — when
the first test lands, factor the wrapper into a shared `test-utils.tsx`.)

---

## 6. Coverage Expectations

There is no per-package coverage threshold enforced today.
- CI does not collect frontend coverage.
- The Go side enforces `-cover` and PR reviewers ask for tests on security-sensitive code.
- Treat critical user flows (auth callback, NDA, mutation rollback paths) as the
  highest-value targets when adding tests.

---

## 7. Lint as the De-Facto Test

`make lint` runs:

- `eslint` via reviewdog on `apps/console`, `apps/trust`, `packages/ui`, `packages/eslint-config`.
- `npx n8n-node lint` on `packages/n8n-node`.

There is **no root `eslint.config.*`**; every workspace carries its own config (extending
`@probo/eslint-config`). `@probo/eslint-plugin-relay-types` enforces the "use Relay-generated
types" rule programmatically — see [conventions.md § 4](./conventions.md#4-typesuse-relay-generated-dont-redeclare).

---

## 8. End-to-End

Frontend behavior is exercised through the Go-driven `e2e/console/` (43 files) and `e2e/mcp/`
(22 files) suites. Add e2e coverage when you introduce a new console flow — see
[`contrib/claude/e2e.md`](../../../contrib/claude/e2e.md) and
[shared.md § 7](../shared.md#7-ci--quality-gates) for how the suite runs (Lima sandbox,
Docker Compose stack, `CGO_ENABLED=1 -race`).
