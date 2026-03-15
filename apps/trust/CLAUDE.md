# apps/trust

React 19 + Vite + TypeScript + Relay + TailwindCSS. Port 5174.

Same frontend stack as `apps/console/` — see its CLAUDE.md for Relay query patterns, mutation hooks, and component conventions.

## Trust-specific differences

- Public-facing trust center app (not an internal dashboard)
- Path-prefix routing: `/trust/{slug}` for Probo-hosted, `/` for custom domains
- Routes: `/overview`, `/documents`, `/subprocessors`, `/updates`
- Auth flow: `/connect`, `/verify-magic-link`, `/full-name`
- Content routes (`/overview`, `/documents`, `/subprocessors`, `/updates`) wrapped in `MainLayout`
- Auth routes (`/connect`, `/verify-magic-link`, `/full-name`) wrapped in `AuthLayout`
- All route groups use `RootErrorBoundary`
