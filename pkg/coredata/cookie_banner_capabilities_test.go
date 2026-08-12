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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestCookieBannerCapabilities_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"null column", nil, true},
		{"empty bytes", []byte{}, true},
		{"object without the key", []byte(`{}`), true},
		{"unrelated key only", []byte(`{"future_capability": true}`), true},
		{"explicit true", []byte(`{"resource_reporting": true}`), true},
		{"explicit false", []byte(`{"resource_reporting": false}`), false},
		{"string payload", `{"resource_reporting": false}`, false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				var capabilities coredata.CookieBannerCapabilities
				require.NoError(t, capabilities.Scan(tt.value))
				assert.Equal(t, tt.want, capabilities.ResourceReporting)
			},
		)
	}
}

func TestCookieBannerCapabilities_ScanUnsupportedType(t *testing.T) {
	t.Parallel()

	var capabilities coredata.CookieBannerCapabilities
	require.Error(t, capabilities.Scan(42))
}

func TestCookieBannerCapabilitiesPatch_Apply(t *testing.T) {
	t.Parallel()

	t.Run(
		"nil member keeps the current value",
		func(t *testing.T) {
			t.Parallel()

			current := coredata.CookieBannerCapabilities{ResourceReporting: true}
			patch := coredata.CookieBannerCapabilitiesPatch{}

			assert.Equal(t, current, patch.Apply(current))
		},
	)

	t.Run(
		"set member overrides the current value",
		func(t *testing.T) {
			t.Parallel()

			current := coredata.CookieBannerCapabilities{ResourceReporting: true}
			patch := coredata.CookieBannerCapabilitiesPatch{ResourceReporting: new(false)}

			assert.False(t, patch.Apply(current).ResourceReporting)
		},
	)
}

func TestCompliancePortalCapabilities_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"null column", nil, true},
		{"object without the key", []byte(`{}`), true},
		{"explicit false", []byte(`{"rights_requests": false}`), false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				var capabilities coredata.CompliancePortalCapabilities
				require.NoError(t, capabilities.Scan(tt.value))
				assert.Equal(t, tt.want, capabilities.RightsRequests)
			},
		)
	}
}

func TestCompliancePortalCapabilitiesPatch_Apply(t *testing.T) {
	t.Parallel()

	current := coredata.CompliancePortalCapabilities{RightsRequests: true}

	assert.Equal(
		t,
		current,
		coredata.CompliancePortalCapabilitiesPatch{}.Apply(current),
	)
	assert.False(
		t,
		coredata.CompliancePortalCapabilitiesPatch{RightsRequests: new(false)}.Apply(current).RightsRequests,
	)
}
