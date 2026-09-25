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

package console_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestConnectorSearchProviders(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()

	t.Run("matches a display name that is not the enum", func(t *testing.T) {
		t.Parallel()

		matched := connectorSearchProviders(registry, "Amazon")

		assert.Contains(t, matched, coredata.ConnectorProviderAWS)
		assert.NotContains(t, matched, coredata.ConnectorProviderBrex)
	})

	t.Run("matches the provider slug with spaces", func(t *testing.T) {
		t.Parallel()

		matched := connectorSearchProviders(registry, "google workspace")

		assert.Contains(t, matched, coredata.ConnectorProviderGoogleWorkspace)
	})

	t.Run("empty query matches nothing", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, connectorSearchProviders(registry, "  "))
	})
}
