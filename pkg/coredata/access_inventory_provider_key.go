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
	"encoding"
	"fmt"
)

// AccessInventoryProviderKey identifies a sync cadence bucket. Connector-
// backed sources use the ConnectorProvider string; CSV and built-in Probo
// memberships use the synthetic keys below (they have no connector_provider).
type AccessInventoryProviderKey string

const (
	AccessInventoryProviderKeyCSV              AccessInventoryProviderKey = "CSV"
	AccessInventoryProviderKeyProboMemberships AccessInventoryProviderKey = "PROBO_MEMBERSHIPS"
)

var (
	_ fmt.Stringer             = AccessInventoryProviderKey("")
	_ encoding.TextMarshaler   = AccessInventoryProviderKey("")
	_ encoding.TextUnmarshaler = (*AccessInventoryProviderKey)(nil)
)

func AccessInventoryProviderKeyFromConnector(provider ConnectorProvider) AccessInventoryProviderKey {
	return AccessInventoryProviderKey(provider)
}

// AccessInventoryProviderKeyForSource picks the cadence key for a live source.
// Connector-backed sources win; otherwise CSV when csv data is present; else
// the built-in Probo memberships driver.
func AccessInventoryProviderKeyForSource(
	connectorProvider *ConnectorProvider,
	hasCSVData bool,
) AccessInventoryProviderKey {
	if connectorProvider != nil {
		return AccessInventoryProviderKeyFromConnector(*connectorProvider)
	}

	if hasCSVData {
		return AccessInventoryProviderKeyCSV
	}

	return AccessInventoryProviderKeyProboMemberships
}

func (k AccessInventoryProviderKey) IsValid() bool {
	switch k {
	case
		AccessInventoryProviderKeyCSV,
		AccessInventoryProviderKeyProboMemberships:
		return true
	}

	return ConnectorProvider(k).IsValid()
}

func (k AccessInventoryProviderKey) String() string {
	return string(k)
}

func (k AccessInventoryProviderKey) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *AccessInventoryProviderKey) UnmarshalText(text []byte) error {
	val := AccessInventoryProviderKey(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessInventoryProviderKey value: %q", string(text))
	}

	*k = val

	return nil
}
