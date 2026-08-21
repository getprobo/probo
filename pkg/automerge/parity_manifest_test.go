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

package automerge_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	parityManifest struct {
		SchemaVersion int                   `json:"schemaVersion"`
		Sources       parityManifestSources `json:"sources"`
		Tests         []parityManifestTest  `json:"tests"`
	}

	parityManifestSources struct {
		Rust       parityManifestSource `json:"rust"`
		JavaScript parityManifestSource `json:"javascript"`
	}

	parityManifestSource struct {
		Version          string `json:"version"`
		GitCommit        string `json:"gitCommit"`
		CrateChecksum    string `json:"crateChecksum"`
		PackageIntegrity string `json:"npmIntegrity"`
	}

	parityManifestTest struct {
		ID             string   `json:"id"`
		Source         string   `json:"source"`
		File           string   `json:"file"`
		Line           int      `json:"line"`
		Name           string   `json:"name"`
		Classification string   `json:"classification"`
		Requirement    string   `json:"requirement"`
		LocalTests     []string `json:"localTests"`
		Rationale      string   `json:"rationale"`
	}
)

func TestUpstreamParityManifest(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/upstream-parity.json")
	require.NoError(t, err)

	var manifest parityManifest
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, 1, manifest.SchemaVersion)
	assert.Equal(t, "0.10.0", manifest.Sources.Rust.Version)
	assert.Equal(
		t,
		"a4f584c86358dd07f83f36708573e1c8d1bd8161",
		manifest.Sources.Rust.GitCommit,
	)
	assert.Equal(
		t,
		"09b78abcbba93428b9465b26cb2816a5b4654cce507f099a84a8c1b311cb3633",
		manifest.Sources.Rust.CrateChecksum,
	)
	assert.Equal(t, "3.4.0", manifest.Sources.JavaScript.Version)
	assert.Equal(
		t,
		"f8b0911dc9d86265dd62934b7dc782571e3a7fcb",
		manifest.Sources.JavaScript.GitCommit,
	)
	assert.Equal(
		t,
		"sha512-THmghtTNGGt2xsI0pM3o1i3PM8oZKcYFgOj25FOzW7l6e94SQOivNtCwy6xc0I8hVJsQSSotoBNs+yk/9hM2dg==",
		manifest.Sources.JavaScript.PackageIntegrity,
	)

	seen := make(map[string]struct{}, len(manifest.Tests))
	pending := make([]string, 0)
	interopPending := make([]string, 0)
	sourceCounts := make(map[string]int)

	for _, test := range manifest.Tests {
		require.NotEmpty(t, test.ID)
		require.NotEmpty(t, test.Source)
		require.NotEmpty(t, test.File)
		require.Positive(t, test.Line)
		require.NotEmpty(t, test.Name)

		_, duplicate := seen[test.ID]
		assert.False(t, duplicate, "duplicate parity test %q", test.ID)
		seen[test.ID] = struct{}{}
		sourceCounts[test.Source]++

		switch test.Classification {
		case "covered":
			assert.NotEmpty(t, test.LocalTests, "covered test %q has no mapping", test.ID)
		case "language-specific":
			assert.NotEmpty(
				t,
				test.Rationale,
				"language-specific test %q has no rationale",
				test.ID,
			)
		case "pending":
			pending = append(pending, test.ID)
		default:
			assert.Fail(
				t,
				"invalid parity classification",
				"test %q has classification %q",
				test.ID,
				test.Classification,
			)
		}

		switch test.Requirement {
		case "interop-required", "api-convenience":
			if test.Classification == "language-specific" {
				assert.Fail(
					t,
					"invalid parity requirement",
					"language-specific test %q is marked %q",
					test.ID,
					test.Requirement,
				)
			}
		case "language-specific":
			assert.Equal(t, "language-specific", test.Classification)
		default:
			assert.Fail(
				t,
				"invalid parity requirement",
				"test %q has requirement %q",
				test.ID,
				test.Requirement,
			)
		}

		if test.Classification == "pending" &&
			test.Requirement == "interop-required" {
			interopPending = append(interopPending, test.ID)
		}
	}

	assert.Equal(t, 361, sourceCounts["rust"])
	assert.Equal(t, 16, sourceCounts["rust-doc"])
	assert.Equal(t, 319, sourceCounts["javascript"])
	assert.Equal(t, 16, sourceCounts["javascript-packaging"])

	if os.Getenv("AUTOMERGE_REQUIRE_FULL_PARITY") == "1" {
		const previewLimit = 10

		preview := pending
		if len(preview) > previewLimit {
			preview = preview[:previewLimit]
		}

		if len(pending) > 0 {
			t.Fatalf(
				"%d upstream tests remain unmapped; first %d: %v",
				len(pending),
				len(preview),
				preview,
			)
		}
	}

	if os.Getenv("AUTOMERGE_REQUIRE_FULL_INTEROP") == "1" &&
		len(interopPending) > 0 {
		const previewLimit = 10

		preview := interopPending
		if len(preview) > previewLimit {
			preview = preview[:previewLimit]
		}

		t.Fatalf(
			"%d interoperability tests remain unmapped; first %d: %v",
			len(interopPending),
			len(preview),
			preview,
		)
	}
}
