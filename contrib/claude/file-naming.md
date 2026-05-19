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

## Worker files

Background worker files use the `<name>_worker.go` naming pattern. The file
contains the handler struct, `NewXxxWorker` constructor, and the `Claim`/`Process`
methods. Keep domain-specific helper methods (match, resolve, etc.) in the same
file.

```
pkg/cookiebanner/tracker_mapping_worker.go
pkg/cookiebanner/pattern_analysis_worker.go
```

## Agent files

When a package has a worker that uses an agent, the agent construction logic
goes in `<worker_prefix>_agent.go` alongside `<worker_prefix>_worker.go`. The
worker file stays focused on `Claim`/`Process` and handler methods; the agent
file owns agent construction, prompt building, constants, config, and the
`//go:embed` directive for prompt templates.

```
pkg/cookiebanner/tracker_mapping_worker.go   -- worker handler
pkg/cookiebanner/tracker_mapping_agent.go    -- agent construction + prompts
```

## Tool files

Each agent tool lives in its own `<tool_name>_tool.go` file, named after the
tool constructor function (without the `Tool` suffix). The file contains the
constructor function plus its param/result types. Avoid grouping multiple tools
in a single file.

```
pkg/cookiebanner/search_tracker_patterns_tool.go   -- searchTrackerPatternsTool
pkg/cookiebanner/search_third_parties_tool.go       -- searchThirdPartiesTool
```
