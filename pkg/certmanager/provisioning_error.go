// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package certmanager

import (
	"errors"
	"strings"
)

const (
	ProvisioningErrorDNSCNAME         = "DNS_CNAME"
	ProvisioningErrorDNSCAA           = "DNS_CAA"
	ProvisioningErrorACMERateLimited  = "ACME_RATE_LIMITED"
	ProvisioningErrorACMEInvalidOrder = "ACME_INVALID_ORDER"
	ProvisioningErrorACMETemporary    = "ACME_TEMPORARY"
	ProvisioningErrorACMEFailed       = "ACME_FAILED"
)

func classifyProvisioningError(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, ErrACMERateLimited) {
		return ProvisioningErrorACMERateLimited
	}

	if errors.Is(err, ErrOrderInvalid) {
		return ProvisioningErrorACMEInvalidOrder
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "cname"):
		return ProvisioningErrorDNSCNAME
	case strings.Contains(msg, "caa record"):
		return ProvisioningErrorDNSCAA
	case strings.Contains(msg, "status: invalid"), strings.Contains(msg, "order is in unexpected status \"invalid\""):
		return ProvisioningErrorACMEInvalidOrder
	default:
		return ProvisioningErrorACMETemporary
	}
}

func provisioningErrorCodePtr(code string) *string {
	if code == "" {
		return nil
	}

	return &code
}
