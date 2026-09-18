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

package tasksync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func TestMergeLinearTeams(t *testing.T) {
	t.Parallel()

	teams := mergeLinearTeams(
		[][]linear.Team{
			{
				{ID: "team-a", Name: "Alpha", Key: "A"},
				{ID: "team-b", Name: "Beta", Key: "B"},
			},
			{
				{ID: "team-a", Name: "Alpha duplicate", Key: "A"},
				{ID: "team-c", Name: "Gamma", Key: "C"},
			},
		},
	)

	require.Len(t, teams, 3)
	assert.Equal(t, "team-a", teams[0].ID)
	assert.Equal(t, "Alpha", teams[0].Name)
	assert.Equal(t, "team-b", teams[1].ID)
	assert.Equal(t, "team-c", teams[2].ID)
}

func TestLinearTeamExists(t *testing.T) {
	t.Parallel()

	teams := []linear.Team{
		{ID: "team-a", Name: "Alpha"},
		{ID: "team-b", Name: "Beta"},
	}

	assert.True(t, linearTeamExists(teams, "team-b"))
	assert.False(t, linearTeamExists(teams, "team-z"))
	assert.False(t, linearTeamExists(nil, "team-a"))
}
