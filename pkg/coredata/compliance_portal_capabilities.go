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

package coredata

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type CompliancePortalCapabilities struct {
	RightsRequests bool `json:"rights_requests"`
}

func DefaultCompliancePortalCapabilities() CompliancePortalCapabilities {
	return CompliancePortalCapabilities{
		RightsRequests: true,
	}
}

func (c CompliancePortalCapabilities) Value() (driver.Value, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal compliance portal capabilities: %w", err)
	}

	return data, nil
}

func (c *CompliancePortalCapabilities) Scan(value any) error {
	if value == nil {
		*c = DefaultCompliancePortalCapabilities()

		return nil
	}

	var data []byte

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan compliance portal capabilities: unsupported type %T", value)
	}

	if len(data) == 0 {
		*c = DefaultCompliancePortalCapabilities()

		return nil
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("cannot unmarshal compliance portal capabilities: %w", err)
	}

	return nil
}
