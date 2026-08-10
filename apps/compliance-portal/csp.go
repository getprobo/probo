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

package complianceportalstatics

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"go.probo.inc/probo/pkg/baseurl"
)

//go:embed content-security-policy.txt.tmpl
var contentSecurityPolicyTmplContent string

var contentSecurityPolicyTmpl = template.Must(
	template.New("content-security-policy").Parse(contentSecurityPolicyTmplContent),
)

type contentSecurityPolicyData struct {
	AppOrigin string
}

// ContentSecurityPolicy returns the compliance-portal CSP with AppOrigin
// substituted (scheme://host of PROBOD_BASE_URL / file download origin).
func ContentSecurityPolicy(appOrigin string) (string, error) {
	origin, err := baseurl.CSPOrigin(appOrigin)
	if err != nil {
		return "", fmt.Errorf("invalid CSP app origin: %w", err)
	}

	var buf bytes.Buffer

	err = contentSecurityPolicyTmpl.Execute(
		&buf,
		contentSecurityPolicyData{AppOrigin: origin},
	)
	if err != nil {
		return "", fmt.Errorf("cannot render content security policy: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}
