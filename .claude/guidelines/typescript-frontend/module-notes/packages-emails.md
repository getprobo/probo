# Probo -- TypeScript Frontend -- packages/emails

> Module-specific notes for `packages/emails` (`@probo/emails`)
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md)

## Purpose

React Email templates for all transactional emails sent by Probo. Templates are authored in TSX using `@react-email/components`, compiled to Go-template HTML files by a build script, and consumed by the Go mailer via `//go:embed dist` and `html/template` rendering.

## Hybrid Architecture

This package spans both stacks -- TSX templates on the TypeScript side, Go `Presenter` on the Go side:

```
Developer writes .tsx template
    |
npm run build (scripts/build.ts)
    |
    v
dist/<name>.html.tmpl + dist/<name>.txt.tmpl
    |
go:embed dist (emails.go)
    |
    v
Go Presenter.Render*() executes templates at send time
```

## Adding a New Email Template

Changes required in **four places**:

1. **New `.tsx` file** in `packages/emails/src/` -- wrap content in `<EmailLayout>`, embed Go template variables as JSX string literals
2. **New `.txt` file** in `packages/emails/templates/` -- plain-text fallback with the same Go template variables
3. **Entry in `scripts/build.ts`** `TemplateConfig` array -- maps slug to component
4. **New `Render*` method** in `packages/emails/emails.go` -- builds template data, executes both templates

Missing any step causes a build failure or runtime panic.

## Go Template Variables in JSX

The fundamental encoding trick: Go template syntax is embedded as JSX string literals so it passes through React rendering unchanged.

```tsx
// From packages/emails/src/Invitation.tsx
<Text style={bodyText}>
  {"{{.RecipientFullName}}"}, you have been invited to join {"{{.OrganizationName}}"}.
</Text>
<Button href={"{{.InvitationURL}}"} style={button}>
  Accept Invitation
</Button>
```

Go range loops:

```tsx
{"{{range .FileNames}}"}
<Text>{"{{.}}"}</Text>
{"{{end}}"}
```

**Pitfall**: Writing `{{.Foo}}` bare in JSX will be treated as JSX expression syntax and cause a TypeScript parse error. Always wrap in `{'{{.Foo}}'}` or `` {`{{.Foo}}`} ``.

## EmailLayout

`packages/emails/src/components/EmailLayout.tsx` is the shared wrapper providing:
- Consistent container and header logo
- Greeting line (`Hi {{.RecipientFullName}},`)
- Footer with company address
- "Powered By Probo" branding
- Exported CSS-in-JS style constants (`button`, `bodyText`, `footerText`)

All email templates compose inside `<EmailLayout>`.

## Template Components Have No Props

Templates accept no props. The build script renders each as a zero-argument function: `render(() => ConfirmEmail())`. All dynamic values are Go template placeholders embedded as literal strings.

## Build Dependency

The `dist/` directory must be built before the Go binary is compiled. `go:embed dist` is evaluated at Go compile time. `make build` orchestrates this, but running `go build` directly will fail if `dist/` is missing.

## Key Files

| File | Purpose |
|------|---------|
| `packages/emails/src/Invitation.tsx` | Canonical short template example |
| `packages/emails/src/components/EmailLayout.tsx` | Shared layout + style constants |
| `packages/emails/scripts/build.ts` | Build pipeline: TSX to `.html.tmpl` |
| `packages/emails/emails.go` | Go Presenter: template execution, asset URL resolution |
| `packages/emails/templates/invitation.txt` | Plain-text fallback example |
