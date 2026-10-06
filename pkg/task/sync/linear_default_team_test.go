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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func TestApplyLinearDefaultTeam(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	staleID := gid.New(tenantID, coredata.ConnectorEntityType)
	chosenID := gid.New(tenantID, coredata.ConnectorEntityType)
	team := &linear.Team{ID: "team-1", Name: "Platform", Key: "PLAT"}

	t.Run("stores the team on the chosen connector and clears the others", func(t *testing.T) {
		t.Parallel()

		stale := &coredata.Connector{ID: staleID}
		require.NoError(t, stale.SetSettings(&coredata.LinearSyncConnectorSettings{
			DefaultTeamID:   "old",
			DefaultTeamName: "Old",
			DefaultTeamKey:  "OLD",
		}))

		chosen := &coredata.Connector{ID: chosenID}
		connectors := []*coredata.Connector{stale, chosen}

		require.NoError(t, applyLinearDefaultTeam(connectors, chosenID, team))

		staleSettings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](stale)
		require.NoError(t, err)
		assert.Empty(t, staleSettings.DefaultTeamID)
		assert.Empty(t, stale.RawSettings)

		chosenSettings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](chosen)
		require.NoError(t, err)
		assert.Equal(t, team.ID, chosenSettings.DefaultTeamID)
		assert.Equal(t, team.Name, chosenSettings.DefaultTeamName)
		assert.Equal(t, team.Key, chosenSettings.DefaultTeamKey)
	})

	t.Run("clears every connector when no team is chosen", func(t *testing.T) {
		t.Parallel()

		first := &coredata.Connector{ID: staleID}
		require.NoError(t, first.SetSettings(&coredata.LinearSyncConnectorSettings{
			DefaultTeamID: "old",
		}))

		second := &coredata.Connector{ID: chosenID}
		require.NoError(t, second.SetSettings(&coredata.LinearSyncConnectorSettings{
			DefaultTeamID: "other",
		}))

		require.NoError(t, applyLinearDefaultTeam([]*coredata.Connector{first, second}, gid.GID{}, nil))

		for _, connector := range []*coredata.Connector{first, second} {
			settings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
			require.NoError(t, err)
			assert.Empty(t, settings.DefaultTeamID)
			assert.Empty(t, connector.RawSettings)
		}
	})
}
