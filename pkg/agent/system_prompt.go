// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package agent

import (
	"bytes"
	"text/template"
)

type (
	systemPromptData struct {
		Instructions string
		Handoffs     []systemPromptHandoff
	}

	systemPromptHandoff struct {
		Name        string
		Description string
	}
)

var systemPromptTmpl = template.Must(template.New("system_prompt").Parse(
	`{{- .Instructions -}}
{{- if .Handoffs }}

## Handoffs
You can transfer the conversation to a more specialized agent when appropriate:
{{ range .Handoffs -}}
- {{ .Name }}{{ with .Description }}: {{ . }}{{ end }}
{{ end -}}
{{- end -}}
`))

func buildSystemPrompt(data systemPromptData) string {
	var buf bytes.Buffer
	_ = systemPromptTmpl.Execute(&buf, data)

	return buf.String()
}
