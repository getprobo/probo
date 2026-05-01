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

package cloudaccount_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudaccount "go.probo.inc/probo/pkg/cmd/cloud-account"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

// TestCloudAccount_Root_RegistersAllSubcommands smoke-tests that the root
// cobra command parses correctly and registers every subcommand the plan
// calls for. This is a guardrail against accidental removal in refactors.
func TestCloudAccount_Root_RegistersAllSubcommands(t *testing.T) {
	t.Parallel()

	io, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: io}
	root := cloudaccount.NewCmdCloudAccount(f)

	require.NotNil(t, root)
	assert.Equal(t, "cloud-account <command>", root.Use)
	assert.Contains(t, root.Aliases, "ca")

	expected := map[string]bool{
		"list":           false,
		"get":            false,
		"create":         false,
		"verify":         false,
		"rotate":         false,
		"delete":         false,
		"install-assets": false,
	}

	for _, sub := range root.Commands() {
		// Use is "verb <args>" or just "verb".
		name := sub.Name()
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for verb, found := range expected {
		assert.True(t, found, "subcommand %q not registered on cloud-account root", verb)
	}
}
