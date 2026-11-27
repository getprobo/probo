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

package console_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAsset_PublishAssetList_SnapshotsTheActiveRows(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)
	manualID := createAssetForPublish(t, owner, profileID, "Manual Laptop")
	insertDiscoveredAsset(t, owner, "Discovered Bucket")
	retiredID := createAssetForPublish(t, owner, profileID, "Retired Printer")
	deactivateAsset(t, owner, retiredID)

	published := publishAssetList(t, owner, map[string]any{
		"minor":          false,
		"organizationId": owner.GetOrganizationID().String(),
		"name":           "Quarter",
		"note":           "from the active rows",
	})

	assert.Contains(t, published.Content, "Manual Laptop")
	assert.Contains(t, published.Content, "Discovered Bucket")
	assert.NotContains(t, published.Content, "Retired Printer")

	pageNames := listAssetNames(t, owner)
	assert.Contains(t, pageNames, "Manual Laptop")
	assert.NotContains(t, pageNames, "Discovered Bucket")
	assert.NotContains(t, pageNames, "Retired Printer")

	snapshots := listAssetSnapshots(t, owner)
	require.Len(t, snapshots, 1)
	assert.Equal(t, "Quarter", snapshots[0].Name)
	assert.Equal(t, "from the active rows", snapshots[0].Note)
	assert.ElementsMatch(t, []string{"Discovered Bucket", "Manual Laptop"}, snapshotAssetNames(snapshots[0]))

	renameAsset(t, owner, manualID, "Renamed Laptop")

	reread := listAssetSnapshots(t, owner)
	require.Len(t, reread, 1)
	assert.Equal(t, snapshots[0].ID, reread[0].ID)
	assert.ElementsMatch(t, []string{"Discovered Bucket", "Manual Laptop"}, snapshotAssetNames(reread[0]))
	assert.Contains(t, documentVersionContent(t, owner, published.VersionID), "Manual Laptop")
	assert.NotContains(t, documentVersionContent(t, owner, published.VersionID), "Renamed Laptop")

	again := publishAssetList(t, owner, map[string]any{
		"minor":          false,
		"organizationId": owner.GetOrganizationID().String(),
	})
	assert.Contains(t, again.Content, "Renamed Laptop")
	assert.NotContains(t, again.Content, "Manual Laptop")

	after := listAssetSnapshots(t, owner)
	require.Len(t, after, 2)

	var named, latest *assetSnapshotView
	for i := range after {
		if after[i].ID == snapshots[0].ID {
			named = &after[i]
			continue
		}

		latest = &after[i]
	}

	require.NotNil(t, named)
	require.NotNil(t, latest)
	assert.ElementsMatch(t, []string{"Discovered Bucket", "Manual Laptop"}, snapshotAssetNames(*named))
	assert.ElementsMatch(t, []string{"Discovered Bucket", "Renamed Laptop"}, snapshotAssetNames(*latest))
}

type publishedAssetList struct {
	VersionID string
	Content   string
}

type assetSnapshotView struct {
	ID     string
	Name   string
	Note   string
	Assets []struct {
		Name string `json:"name"`
	}
}

func publishAssetList(
	t *testing.T,
	client *testutil.Client,
	input map[string]any,
) publishedAssetList {
	t.Helper()

	const query = `
		mutation($input: PublishAssetListInput!) {
			publishAssetList(input: $input) {
				documentVersionEdge {
					node {
						id
						content { prosemirror }
					}
				}
			}
		}
	`

	var result struct {
		PublishAssetList struct {
			DocumentVersionEdge struct {
				Node struct {
					ID      string `json:"id"`
					Content struct {
						Prosemirror string `json:"prosemirror"`
					} `json:"content"`
				} `json:"node"`
			} `json:"documentVersionEdge"`
		} `json:"publishAssetList"`
	}

	require.NoError(t, client.Execute(query, map[string]any{"input": input}, &result))

	version := result.PublishAssetList.DocumentVersionEdge.Node

	return publishedAssetList{
		VersionID: version.ID,
		Content:   version.Content.Prosemirror,
	}
}

func listAssetNames(t *testing.T, client *testutil.Client) []string {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					assets(first: 50) {
						edges { node { name } }
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Assets struct {
				Edges []struct {
					Node struct {
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"assets"`
		} `json:"node"`
	}

	require.NoError(t, client.Execute(query, map[string]any{
		"id": client.GetOrganizationID().String(),
	}, &result))

	names := make([]string, 0, len(result.Node.Assets.Edges))
	for _, edge := range result.Node.Assets.Edges {
		names = append(names, edge.Node.Name)
	}

	return names
}

func listAssetSnapshots(t *testing.T, client *testutil.Client) []assetSnapshotView {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					assetSnapshots(first: 10, orderBy: { field: CAPTURED_AT, direction: DESC }) {
						edges {
							node {
								id
								name
								note
								assets(first: 50) {
									edges { node { name } }
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			AssetSnapshots struct {
				Edges []struct {
					Node struct {
						ID     string  `json:"id"`
						Name   *string `json:"name"`
						Note   *string `json:"note"`
						Assets struct {
							Edges []struct {
								Node struct {
									Name string `json:"name"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"assets"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"assetSnapshots"`
		} `json:"node"`
	}

	require.NoError(t, client.Execute(query, map[string]any{
		"id": client.GetOrganizationID().String(),
	}, &result))

	snapshots := make([]assetSnapshotView, 0, len(result.Node.AssetSnapshots.Edges))
	for _, edge := range result.Node.AssetSnapshots.Edges {
		view := assetSnapshotView{ID: edge.Node.ID}
		if edge.Node.Name != nil {
			view.Name = *edge.Node.Name
		}
		if edge.Node.Note != nil {
			view.Note = *edge.Node.Note
		}
		for _, asset := range edge.Node.Assets.Edges {
			view.Assets = append(view.Assets, struct {
				Name string `json:"name"`
			}{Name: asset.Node.Name})
		}
		snapshots = append(snapshots, view)
	}

	return snapshots
}

func snapshotAssetNames(snapshot assetSnapshotView) []string {
	names := make([]string, 0, len(snapshot.Assets))
	for _, asset := range snapshot.Assets {
		names = append(names, asset.Name)
	}

	return names
}

func documentVersionContent(t *testing.T, client *testutil.Client, versionID string) string {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					content { prosemirror }
				}
			}
		}
	`

	var result struct {
		Node struct {
			Content struct {
				Prosemirror string `json:"prosemirror"`
			} `json:"content"`
		} `json:"node"`
	}

	require.NoError(t, client.Execute(query, map[string]any{"id": versionID}, &result))

	return result.Node.Content.Prosemirror
}

func deactivateAsset(t *testing.T, client *testutil.Client, assetID string) {
	t.Helper()

	const query = `
		mutation($input: UpdateAssetInput!) {
			updateAsset(input: $input) { asset { id status } }
		}
	`

	var result struct {
		UpdateAsset struct {
			Asset struct {
				Status string `json:"status"`
			} `json:"asset"`
		} `json:"updateAsset"`
	}

	require.NoError(t, client.Execute(query, map[string]any{
		"input": map[string]any{
			"id":     assetID,
			"status": "DEACTIVATED",
		},
	}, &result))
	assert.Equal(t, "DEACTIVATED", result.UpdateAsset.Asset.Status)
}

func renameAsset(t *testing.T, client *testutil.Client, assetID string, name string) {
	t.Helper()

	const query = `
		mutation($input: UpdateAssetInput!) {
			updateAsset(input: $input) { asset { name } }
		}
	`

	require.NoError(t, client.Execute(query, map[string]any{
		"input": map[string]any{
			"id":   assetID,
			"name": name,
		},
	}, &struct{}{}))
}

func insertDiscoveredAsset(t *testing.T, client *testutil.Client, name string) {
	t.Helper()

	orgID := client.GetOrganizationID()
	tenantID := orgID.TenantID()
	now := time.Now().UTC()
	assetID := gid.New(tenantID, coredata.AssetEntityType)
	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	conn := dialTestPg(t, ctx)
	defer func() { _ = conn.Close(ctx) }()

	_, err := conn.Exec(
		ctx,
		`INSERT INTO assets (
			id, tenant_id, name, organization_id, amount, asset_type,
			kind, external_id, configuration, status, connector_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 1, 'VIRTUAL',
			'object-storage', $5, '{"encrypted":true}'::jsonb, 'ACTIVE', $6,
			$7, $7
		)`,
		assetID,
		tenantID,
		name,
		orgID,
		"bucket-"+assetID.String(),
		connectorID,
		now,
	)
	require.NoError(t, err)
}
