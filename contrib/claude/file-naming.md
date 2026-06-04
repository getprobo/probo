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

A file whose **sole purpose** is to construct and operate an agent uses the
`<name>_agent.go` suffix. The agent file owns agent construction, prompt
building, the `//go:embed` directive for prompt templates, the typed result,
the agent-specific config, and any agent-only constants (timeout, confidence
threshold). Callers (workers, services) hold a `*agent.Agent` field and import
the file's `Build…Agent` constructor.

This applies in two shapes:

1. **Paired with a worker** (most common). The worker file
   `<worker_prefix>_worker.go` stays focused on `Claim`/`Process`, and the
   agent it uses lives in `<worker_prefix>_agent.go` next to it.

   ```
   pkg/cookiebanner/tracker_mapping_worker.go   -- worker handler
   pkg/cookiebanner/tracker_mapping_agent.go    -- agent construction + prompts
   ```

2. **Standalone, called from elsewhere.** When the agent is consumed by a
   different package (or by multiple packages — e.g. an agent that operates on
   a domain entity, used by several feature workers), it lives in the package
   that owns the domain, named `<purpose>_agent.go`.

   ```
   pkg/thirdparty/disambiguation_agent.go   -- catalog→org ThirdParty matcher
   pkg/vetting/sub_agent.go                  -- generic vetting sub-agent
   ```

A file is NOT renamed to `_agent.go` when the agent is incidental to a service
that does substantially more than agent orchestration (e.g. CRUD, caching,
auth). In that case the file keeps its service name and the agent is built
inline:

```
pkg/evidenceassessor/assessment.go            -- single-file evidence assessment service
pkg/vetting/assessment.go                     -- third-party assessment service
```

If the agent construction grows past a few dozen lines or sprouts its own
prompt embed / typed result / config struct, extract it into a sibling
`<purpose>_agent.go`.

## Tool files

Each agent tool lives in its own `<tool_name>_tool.go` file, named after the
tool constructor function (without the `Tool` suffix). The file contains the
constructor function plus its param/result types. Avoid grouping multiple tools
in a single file.

```
pkg/cookiebanner/search_tracker_patterns_tool.go   -- searchTrackerPatternsTool
pkg/cookiebanner/search_third_parties_tool.go       -- searchThirdPartiesTool
```
