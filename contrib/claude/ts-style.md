# TypeScript Style

## Related guides

| Topic | Guide |
|-------|--------|
| GraphQL data (the default for app data) | [`contrib/claude/relay.md`](relay.md) |
| Forms and file inputs | [`contrib/claude/forms.md`](forms.md) |

## URL and query parameter construction

**Never** build URLs with template literals, string concatenation, or string formatting. Always use the `URL` and `URLSearchParams` APIs.

- Use `new URL()` to construct or parse full URLs.
- Use `URLSearchParams` to build query strings.
- Use `.pathname`, `.searchParams`, and other `URL` properties to modify parts of a URL safely.
- Use `encodeURIComponent` for dynamic path segments.

```typescript
// Bad — template literal
const endpoint = `https://api.example.com/users/${userId}?active=${active}`;

// Bad — string concatenation
const endpoint = "https://api.example.com/orgs/" + orgId + "/members";

// Bad — query params via string concat
const qs = "?domain=" + domain + "&limit=100";

// Good — URL object
const url = new URL("https://api.example.com");
url.pathname = `/users/${encodeURIComponent(userId)}`;
url.searchParams.set("active", String(active));

// Good — URLSearchParams for query strings
const params = new URLSearchParams();
params.set("domain", domain);
params.set("limit", "100");
const qs = params.toString();
```

## Non-Relay HTTP

Application data is GraphQL via Relay (see [`contrib/claude/relay.md`](relay.md)) — that is the default and covers almost everything. Reach for `fetch` only for the cases GraphQL does not handle: binary uploads/downloads, REST endpoints exposed by the backend, and health/probe calls.

Rules:

- Build the endpoint with `URL` (see above). Derive the host/path from app config (e.g. `import.meta.env.VITE_API_URL`, a `pathPrefix` helper), never a hardcoded string.
- Use `fetch` directly; do not add an HTTP client dependency for a handful of calls.
- Always check `response.ok` and throw a typed error on failure so it can be caught (`try`/`catch` in the handler — see [`contrib/claude/error-handling.md`](error-handling.md)) and surfaced via a toast.
- Send credentials/auth the same way the Relay environment does (e.g. `credentials: "include"`); do not invent a second auth scheme.

```ts
// Good — URL-built endpoint, ok check, typed failure
const url = new URL(buildEndpoint());
url.pathname = `/api/console/v1/documents/${encodeURIComponent(documentId)}/download`;

const response = await fetch(url, { credentials: "include" });
if (!response.ok) {
  throw new Error(`Download failed: ${response.status}`);
}
const blob = await response.blob();
```

### File upload / download

- **Uploads:** collect the file with `react-dropzone` (already a `@probo/ui` dependency) or a native `<input type="file">`, then send a `FormData` body via `fetch` (or the GraphQL upload mechanism if the schema exposes one). Validate type/size client-side with the `@probo/helpers` `fileAccept` helpers before sending; the server still validates.
- **Downloads:** prefer a direct link to a backend URL when the endpoint streams a file; use `fetch` + `blob` only when you must read the bytes (e.g. to rename or post-process). Revoke any `URL.createObjectURL` you create.

```ts
// Bad — template-literal URL, no ok check, ad-hoc auth header
const res = await fetch(`${base}/upload?id=${id}`, {
  headers: { Authorization: "Bearer " + token },
  body: file,
});
```
