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

package probo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
)

func TestCollapseTrackerPolicyThirdParties_ConflictingPrivacyPolicyURLs(t *testing.T) {
	t.Parallel()

	low := gidWithSuffix(1)
	high := gidWithSuffix(2)
	urlLow := new("https://low.example/privacy")
	urlHigh := new("https://high.example/privacy")

	lowFirst := coredata.CommonThirdParties{
		{ID: low, Name: "Hotjar", PrivacyPolicyURL: urlLow},
		{ID: high, Name: "hotjar", PrivacyPolicyURL: urlHigh},
	}
	highFirst := coredata.CommonThirdParties{
		{ID: high, Name: "hotjar", PrivacyPolicyURL: urlHigh},
		{ID: low, Name: "Hotjar", PrivacyPolicyURL: urlLow},
	}

	want := []docgen.TrackerPolicyThirdParty{
		{Name: "Hotjar", PrivacyPolicyURL: *urlLow},
	}

	assert.Equal(t, want, collapseTrackerPolicyThirdParties(lowFirst))
	assert.Equal(t, want, collapseTrackerPolicyThirdParties(highFirst))
}

func TestCollapseTrackerPolicyThirdParties_BackfillsEmptyPrivacyPolicyURL(t *testing.T) {
	t.Parallel()

	low := gidWithSuffix(1)
	high := gidWithSuffix(2)
	urlHigh := new("https://high.example/privacy")

	parties := coredata.CommonThirdParties{
		{ID: high, Name: "Hotjar", PrivacyPolicyURL: urlHigh},
		{ID: low, Name: "Hotjar"},
	}

	rows := collapseTrackerPolicyThirdParties(parties)
	require.Len(t, rows, 1)
	assert.Equal(t, *urlHigh, rows[0].PrivacyPolicyURL)
}

func gidWithSuffix(suffix byte) gid.GID {
	var id gid.GID

	id[len(id)-1] = suffix

	return id
}
