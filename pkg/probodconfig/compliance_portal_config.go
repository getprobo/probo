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

package probodconfig

import "fmt"

type CompliancePortalTLSMode string

const (
	CompliancePortalTLSModeDirect   CompliancePortalTLSMode = "direct"
	CompliancePortalTLSModeExternal CompliancePortalTLSMode = "external"
)

func ParseCompliancePortalTLSMode(value string) (CompliancePortalTLSMode, error) {
	mode := CompliancePortalTLSMode(value)

	switch mode {
	case "":
		return CompliancePortalTLSModeDirect, nil
	case CompliancePortalTLSModeDirect, CompliancePortalTLSModeExternal:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported compliance portal TLS mode %q", value)
	}
}

func (m CompliancePortalTLSMode) IsExternal() bool {
	return m == CompliancePortalTLSModeExternal
}

func (m CompliancePortalTLSMode) Validate() error {
	switch m {
	case "", CompliancePortalTLSModeDirect, CompliancePortalTLSModeExternal:
		return nil
	default:
		return fmt.Errorf("unsupported compliance portal TLS mode %q", m)
	}
}
