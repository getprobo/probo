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

package mcp_v1

import (
	"fmt"
	"net/http"
	"strings"
)

const toolsetDocuments = "documents"

// InDocumentToolset reports whether name belongs on the documents MCP toolset.
// Compliance-portal catalog tools stay on the full server: they remove a
// portal link, they do not edit the document.
func InDocumentToolset(name string) bool {
	switch name {
	case "listOrganizations", "getOrganizationContext", "listUsers", "getUser",
		"signDocument", "cancelSignatureRequest":
		return true
	}

	if strings.Contains(name, "CompliancePortal") {
		return false
	}

	return strings.Contains(name, "Document")
}

func knownToolset(name string) bool {
	return name == toolsetDocuments
}

func toolsetFromRequest(r *http.Request) (string, error) {
	if r == nil || r.URL == nil {
		return "", nil
	}

	queryName := r.URL.Query().Get("toolset")

	var pathName string

	path := strings.TrimSuffix(r.URL.Path, "/")
	if rest, ok := strings.CutPrefix(path, "/toolsets/"); ok && rest != "" && !strings.Contains(rest, "/") {
		pathName = rest
	}

	if queryName != "" && pathName != "" && queryName != pathName {
		return "", fmt.Errorf("conflicting mcp toolset")
	}

	name := queryName
	if name == "" {
		name = pathName
	}

	if name == "" {
		return "", nil
	}

	if !knownToolset(name) {
		return "", fmt.Errorf("unknown mcp toolset")
	}

	return name, nil
}
