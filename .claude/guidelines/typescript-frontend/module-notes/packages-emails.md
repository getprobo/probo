# Probo — TypeScript Frontend — @probo/emails

**14 React Email templates** authored as TSX, pre-rendered at build time into Go template files
that the backend embeds via `//go:embed`. The package straddles the TS/Go boundary: TS produces
the artifacts, Go consumes them.

```
packages/emails/
  src/
    EmailLayout.tsx           ← branding tokens (colors, header logo)
    <Template>.tsx            ← React Email JSX with Go template placeholders
  scripts/
    build.ts                  ← tsx + render() → dist/<Template>.html.tmpl
                                                 dist/<Template>.txt.tmpl
  dist/                       ← generated; consumed by pkg/mailer via //go:embed
```

**Placeholders are Go template strings inside JSX literals**:

```tsx
<p>Hello {`{{ .UserFullName }}`},</p>
<p>Click <a href={`{{ .ActionURL }}`}>this link</a> to continue.</p>
```

There is **no TS-side type-check** of placeholder names — typos surface only at Go runtime when
`text/template` complains. Cross-check placeholder names against the data struct in `pkg/mailer`.

## Key files

- `packages/emails/src/EmailLayout.tsx` — single source of truth for email branding (colors,
  fonts, header logo). Update both this file **and** the app Tailwind theme when brand colors
  change.
- `packages/emails/scripts/build.ts` — the build entry; runs via `tsx`.

## How to extend

1. Add `packages/emails/src/<NewTemplate>.tsx` using `EmailLayout` and React Email primitives.
2. Use `{`{{ .FieldName }}`}` placeholders matching the Go data struct in `pkg/mailer`.
3. Run the package build (`tsx scripts/build.ts`, or `make build WITH_APPS=1`) — verify
   `dist/<NewTemplate>.html.tmpl` and `dist/<NewTemplate>.txt.tmpl` are generated.
4. Wire the new template into `pkg/mailer` (Go side) — add it to `//go:embed` and the dispatch
   map.
5. The four-surface API rule applies if the email is triggered by a new public operation — see
   [shared.md § 3](../shared.md#3-the-four-surface-api-rule).

## Top pitfalls

1. **Stale `dist/` at `go build` time** — see
   [pitfalls.md § 15](../pitfalls.md#15-packagesemails-dist-stale-at-go-build-time).
2. **Placeholder typo only fails at Go runtime** — see
   [pitfalls.md § 16](../pitfalls.md#16-email-template-placeholders-typo-only-surfaces-at-go-runtime).
