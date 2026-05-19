# File naming conventions

## Template files

Template files use the extension pattern `<name>.<output-ext>.tmpl`:

```
pkg/trust/sitemap.xml.tmpl
pkg/server/mailactions/templates/page.html.tmpl
pkg/cookiebanner/prompts/tracker_identification.txt.tmpl
pkg/probo/templates/risk_list.json.tmpl
compose/keycloak/probo-realm.json.tmpl
pkg/esign/certificate.html.tmpl
```

The `<output-ext>` indicates the format produced after rendering (`.xml`, `.html`, `.txt`, `.json`). The final `.tmpl` suffix marks the file as a template requiring substitution before use.

### Dynamic values over hardcoded lists

Never hardcode enum values or source-of-truth lists in template text. Use a
placeholder and substitute at runtime so the template stays in sync with the
code:

```
Use one of: {{.Categories}}.
```

For single-placeholder templates, `strings.Replace` is sufficient. For multiple
placeholders, use `text/template`.
