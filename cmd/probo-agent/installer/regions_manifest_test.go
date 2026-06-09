// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
