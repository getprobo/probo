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

package provider

import (
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

// ResolveInitialAccount returns the account present from the start, before any
// account discovered or enabled later. A provider with no such field returns
// empty strings. Settings that cannot be decoded return an error.
func (r *Registration) ResolveInitialAccount(c *coredata.Connector) (externalID string, name string, err error) {
	return r.InitialAccountFunc(c)
}

func emptyInitialAccount(*coredata.Connector) (string, string, error) {
	return "", "", nil
}

// initialAccount reads one settings field and uses it as both the external
// account id and the account name.
func initialAccount[T any](field func(T) string) func(*coredata.Connector) (string, string, error) {
	return func(c *coredata.Connector) (string, string, error) {
		if c == nil {
			return "", "", nil
		}

		settings, err := coredata.ConnectorSettings[T](c)
		if err != nil {
			return "", "", fmt.Errorf("cannot read connector settings: %w", err)
		}

		id := field(settings)

		return id, id, nil
	}
}
