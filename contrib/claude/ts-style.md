# TypeScript Style

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
