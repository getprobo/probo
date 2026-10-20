// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package filetemplate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
)

const sampleBody = `{{define "body"}}
{{template "masthead" .Layout}}
{{template "kicker" .Note}}
{{template "title" (titleBlock .Layout.Title "Line <b>")}}
<dl class="meta">
{{template "metaEntry" (meta "Owner" .Note)}}
</dl>
{{template "section" (heading ">" "Checks")}}
{{template "status" "Pass"}}
{{template "status" "Fail"}}
{{template "status" "Unknown"}}
{{template "status" "Not applicable"}}
{{template "status" "<script>"}}
{{template "callout" (callout "Empty" "Nothing <i>here</i>")}}
{{end}}`

type sampleDocument struct {
	Layout Layout
	Note   string
}

func TestRender_SharedPartials(t *testing.T) {
	t.Parallel()

	parsed := MustParse("sample", sampleBody)
	html, err := Render(parsed, sampleDocument{
		Layout: Layout{
			Title:            `Report <script>`,
			OrganizationName: `Acme <script>`,
			LogoSrc:          template.URL("data:image/png;base64,aaaa"),
			Classification:   "CONFIDENTIAL",
		},
		Note: `Ada <b>Lovelace</b>`,
	})
	require.NoError(t, err)

	body := string(html)
	assert.Contains(t, body, "<title>Report &lt;script&gt;</title>")
	assert.Contains(t, body, "Acme &lt;script&gt;")
	assert.Contains(t, body, `class="org-logo"`)
	assert.Contains(t, body, `src="data:image/png;base64,aaaa"`)
	assert.Contains(t, body, "Ada &lt;b&gt;Lovelace&lt;/b&gt;")
	assert.Contains(t, body, "Line &lt;b&gt;")
	assert.Contains(t, body, `<span class="index">&gt;</span>`)
	assert.Contains(t, body, "Checks")
	assert.Contains(t, body, `class="mark mark-pass"`)
	assert.Contains(t, body, `aria-label="Pass"`)
	assert.Contains(t, body, `class="mark mark-fail"`)
	assert.Contains(t, body, `aria-label="Fail"`)
	assert.Contains(t, body, `class="mark mark-unknown"`)
	assert.Contains(t, body, `aria-label="Unknown"`)
	assert.Contains(t, body, "N/A")
	assert.NotContains(t, body, "Not applicable")
	assert.Contains(t, body, "&lt;script&gt;")
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, "<b>")
	assert.NotContains(t, body, "<i>")
	assert.Contains(t, body, "Nothing &lt;i&gt;here&lt;/i&gt;")
	assert.Contains(t, body, "col.inset-left")
	assert.Contains(t, body, "width: 5%")
	assert.Contains(t, body, "width: 15%")
	assert.NotContains(t, body, "Powered by")
}

func TestRender_OmitsEmptyLogo(t *testing.T) {
	t.Parallel()

	parsed := MustParse("sample", sampleBody)
	html, err := Render(parsed, sampleDocument{
		Layout: Layout{
			Title:            "Report",
			OrganizationName: "Acme",
		},
		Note: "Note",
	})
	require.NoError(t, err)

	assert.NotContains(t, string(html), `class="org-logo"`)
}

func TestParse_InvalidBody(t *testing.T) {
	t.Parallel()

	_, err := Parse("broken", `{{define "body"}}{{end`)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot parse file template body")
}

func TestFooter(t *testing.T) {
	t.Parallel()

	footer := Footer(`CONFIDENTIAL <script>`)

	assert.Contains(t, footer, `viewBox="0 0 90 24"`)
	assert.Contains(t, footer, "CONFIDENTIAL &lt;script&gt;")
	assert.NotContains(t, footer, "<script>")
	assert.Contains(t, footer, "background-color:#e5484d")
	assert.Contains(t, footer, "-webkit-print-color-adjust:exact")
	assert.Contains(t, footer, "print-color-adjust:exact")
	assert.Contains(t, footer, "color:#fff")
	assert.Contains(t, footer, `<span class="pageNumber"></span>`)
	assert.Contains(t, footer, `<span class="totalPages"></span>`)
	assert.NotContains(t, footer, "Powered by")
	assert.Equal(t, 1, strings.Count(footer, "CONFIDENTIAL"))
}
