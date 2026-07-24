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

package installer_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/deviceagent"
)

type regionsManifest struct {
	Regions []regionEntry `json:"regions"`
}

type regionEntry struct {
	ID        string  `json:"id"`
	ServerURL *string `json:"server_url"`
}

func TestRegionsManifestMatchesGoConstants(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("regions.json")
	require.NoError(t, err)

	var manifest regionsManifest
	require.NoError(t, json.Unmarshal(data, &manifest))

	urls := map[string]string{}

	for _, region := range manifest.Regions {
		if region.ServerURL != nil {
			urls[region.ID] = *region.ServerURL
		}
	}

	assert.Equal(t, deviceagent.USConsoleURL, urls["us"])
	assert.Equal(t, deviceagent.EUConsoleURL, urls["eu"])
	assert.NotContains(t, urls, "self_hosted")
}
