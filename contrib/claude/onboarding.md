# Organization onboarding wizard (console)

Unlisted wizard at `/organizations/:organizationId/onboarding` for peer / new
organization setup. Not linked from sidebar or nav; shareable URL plus redirect
after **Create organization**.

## Steps

| # | Step ID | Purpose | Done (GraphQL) |
|---|---------|---------|----------------|
| 1 | `scim` | IdP provisioning via SCIM | `organization.scimConfiguration` exists |
| 2 | `accessReview` | Access review connectors | ≥1 `accessReviewSources` edge |
| 3 | `agent` | Probo agent / device fleet | ≥1 `devices` edge |
| 4 | `mcp` | MCP setup instructions | Never auto; **Continue** always |
| 5 | `congrats` | Completion | Reach step 5 in flow |

## UX

- **Learn** panel: title, purpose, bullets, SVG/CSS illustration (`prefers-reduced-motion`).
- **Do** panel: embedded existing flows (SCIM connectors, add access review source, create device).
- **Do this later** on steps 1–4: session-only deferral; advances to next step.
- **Continue** on MCP without checkbox.
- **Congrats**: done vs still-to-do from integration state; optional link to console tasks.

## Persistence

- No backend onboarding model.
- **Done** = live integration data (refetch after mutations).
- **Deferred** = in-memory React context only (no `localStorage`); lost on tab close or refresh.

## Navigation

- Post-create: navigate to `/organizations/:id/onboarding`.
- Step in URL: `?step=scim|accessReview|agent|mcp|congrats` (browser back).
- Initial step: first incomplete integration step, or **congrats** if all complete.
- `?welcome=1` after create: start at step 1 even if empty (optional; default first gap).

## Access

- Authenticated org member required.
- **Employee** → redirect to `employee` portal.
- **Auditor** → redirect to `measures`.
- Missing permissions: show learn panel + deep link to settings; still allow **Do this later**.

## Layout

- `hideSidebar` org layout (focused wizard), same pattern as employee layout.

## Deep links (finish later)

- SCIM → `/organizations/:id/settings/scim`
- Access review → `/organizations/:id/access-reviews/sources`
- Devices → `/organizations/:id/devices`
- MCP → instance `${origin}` + `/mcp/v1`, skills / IDE docs

## i18n

- Namespace `organizations/onboarding` → `pages/organizations/onboarding/_locales/*.json`
