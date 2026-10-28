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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	// Pasting a full title used to be a tsquery syntax error on every list.
	punctuationSearchQuery = `POL-004: Access Control Policy (v1.0) & Bilt's "rules" a\b !x | y <-> z   `

	policySearchTitle = "POL-004: Access Control Policy (v1.0)"
	otherSearchTitle  = "Other doc"
)

func TestListSearch_AcceptsPunctuation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()
	frameworkID := factory.NewFramework(owner).
		WithName(factory.SafeName("Framework")).
		Create()

	factory.NewDocument(owner).WithTitle("Punctuation probe").Create()
	factory.NewControl(owner, frameworkID).
		WithName("Punctuation probe").
		WithSectionTitle("zzz").
		WithDescription("zzz").
		Create()
	factory.NewRisk(owner).WithName("Punctuation probe").Create()
	factory.NewInternalControl(owner).WithName("Punctuation probe").Create()

	profileID := factory.CreateUser(owner, factory.Attrs{"fullName": "Punctuation Probe"})
	createDevice(t, owner, organizationID, &profileID)

	searches := []struct {
		name   string
		search func(*testing.T, string) searchConnection
	}{
		{"documents", func(t *testing.T, query string) searchConnection {
			return searchDocuments(t, owner, organizationID, query)
		}},
		{"controls", func(t *testing.T, query string) searchConnection {
			return searchControls(t, owner, organizationID, query)
		}},
		{"risks", func(t *testing.T, query string) searchConnection {
			return searchRisks(t, owner, organizationID, query)
		}},
		{"internal controls", func(t *testing.T, query string) searchConnection {
			return searchInternalControls(t, owner, organizationID, query)
		}},
		{"devices", func(t *testing.T, query string) searchConnection {
			return searchDevices(t, owner, organizationID, query)
		}},
	}

	for _, search := range searches {
		t.Run(search.name, func(t *testing.T) {
			t.Parallel()

			search.search(t, punctuationSearchQuery)
		})
	}
}

func TestListSearch_MatchesWordPrefixes(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()
	frameworkID := factory.NewFramework(owner).
		WithName(factory.SafeName("Framework")).
		Create()

	policyDocumentID := factory.NewDocument(owner).WithTitle(policySearchTitle).Create()
	otherDocumentID := factory.NewDocument(owner).WithTitle(otherSearchTitle).Create()

	policyControlID := factory.NewControl(owner, frameworkID).
		WithName(policySearchTitle).
		WithSectionTitle("zza").
		WithDescription("zza").
		Create()
	otherControlID := factory.NewControl(owner, frameworkID).
		WithName(otherSearchTitle).
		WithSectionTitle("zzb").
		WithDescription("zzb").
		Create()

	referenceRiskID := factory.NewRisk(owner).WithName("Unrelated register").Create()
	policyRiskID := factory.NewRisk(owner).WithName(policySearchTitle).Create()
	otherRiskID := factory.NewRisk(owner).WithName(otherSearchTitle).Create()
	require.Equal(t, "RSK-001", riskReferenceID(t, owner, referenceRiskID))

	policyInternalControlID := factory.NewInternalControl(owner).
		WithName(policySearchTitle).
		Create()
	otherInternalControlID := factory.NewInternalControl(owner).
		WithName(otherSearchTitle).
		Create()

	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, policyRiskID, analysisID)
	planID := factory.CreateTreatmentPlan(owner, policyRiskID, analysisID)
	factory.LinkTreatmentPlanInternalControl(owner, planID, policyInternalControlID)
	factory.LinkTreatmentPlanInternalControl(owner, planID, otherInternalControlID)

	asOf := time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano)

	policyOwnerID := factory.CreateUser(owner, factory.Attrs{"fullName": policySearchTitle})
	otherOwnerID := factory.CreateUser(owner, factory.Attrs{"fullName": otherSearchTitle})
	policyDeviceID := createDevice(t, owner, organizationID, &policyOwnerID).CreateDevice.Device.ID
	otherDeviceID := createDevice(t, owner, organizationID, &otherOwnerID).CreateDevice.Device.ID

	// Documents, controls, and risks index '-' and '+' as spaces, so the
	// digits in POL-004 and RSK-001 are searchable. Internal controls, their
	// history, and device owner names keep the hyphen.
	lists := []listSearchList{
		{
			name:        "documents",
			policyID:    policyDocumentID,
			otherID:     otherDocumentID,
			allIDs:      []string{policyDocumentID, otherDocumentID},
			matchDigits: true,
			search: func(t *testing.T, query string) searchConnection {
				return searchDocuments(t, owner, organizationID, query)
			},
		},
		{
			name:        "controls",
			policyID:    policyControlID,
			otherID:     otherControlID,
			allIDs:      []string{policyControlID, otherControlID},
			matchDigits: true,
			search: func(t *testing.T, query string) searchConnection {
				return searchControls(t, owner, organizationID, query)
			},
		},
		{
			name:        "risks",
			policyID:    policyRiskID,
			otherID:     otherRiskID,
			allIDs:      []string{referenceRiskID, policyRiskID, otherRiskID},
			referenceID: referenceRiskID,
			matchDigits: true,
			search: func(t *testing.T, query string) searchConnection {
				return searchRisks(t, owner, organizationID, query)
			},
		},
		{
			name:     "internal controls",
			policyID: policyInternalControlID,
			otherID:  otherInternalControlID,
			allIDs:   []string{policyInternalControlID, otherInternalControlID},
			search: func(t *testing.T, query string) searchConnection {
				return searchInternalControls(t, owner, organizationID, query)
			},
		},
		{
			name:     "internal control history",
			policyID: policyInternalControlID,
			otherID:  otherInternalControlID,
			allIDs:   []string{policyInternalControlID, otherInternalControlID},
			search: func(t *testing.T, query string) searchConnection {
				return searchInternalControlsAsOf(t, owner, planID, asOf, query)
			},
		},
		{
			name:     "devices",
			policyID: policyDeviceID,
			otherID:  otherDeviceID,
			allIDs:   []string{policyDeviceID, otherDeviceID},
			search: func(t *testing.T, query string) searchConnection {
				return searchDevices(t, owner, organizationID, query)
			},
		},
	}

	queries := []string{
		"acc con",
		"control access",
		"POL-004:",
		"v1.0",
		"doc",
		"olic",
		"& (",
		"004",
		"001",
	}

	for _, list := range lists {
		for _, query := range queries {
			t.Run(list.name+"/"+query, func(t *testing.T) {
				t.Parallel()

				assertSearchHits(t, list.search(t, query), list.want(t, query))
			})
		}
	}
}

func TestListSearch_MatchesWordsAcrossControlColumns(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()
	frameworkID := factory.NewFramework(owner).
		WithName(factory.SafeName("Framework")).
		Create()

	accessID := factory.NewControl(owner, frameworkID).
		WithSectionTitle("zza").
		WithName("Access").
		WithDescription("Review notes").
		Create()
	factory.NewControl(owner, frameworkID).
		WithSectionTitle("zzb").
		WithName("Other doc").
		WithDescription("zzb").
		Create()
	backupID := factory.NewControl(owner, frameworkID).
		WithSectionTitle("AC-2").
		WithName("Backup").
		WithDescription("zzc").
		Create()

	tests := []struct {
		query string
		want  []string
	}{
		{"access review", []string{accessID}},
		{"2", []string{backupID}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			t.Parallel()

			assertSearchHits(t, searchControls(t, owner, organizationID, tt.query), tt.want)
		})
	}
}

type searchConnection struct {
	TotalCount int `json:"totalCount"`
	Edges      []struct {
		Node struct {
			ID string `json:"id"`
		} `json:"node"`
	} `json:"edges"`
}

type listSearchList struct {
	name        string
	policyID    string
	otherID     string
	allIDs      []string
	referenceID string
	matchDigits bool
	search      func(t *testing.T, query string) searchConnection
}

func (l listSearchList) want(t *testing.T, query string) []string {
	t.Helper()

	switch query {
	case "acc con", "control access", "POL-004:", "v1.0":
		return []string{l.policyID}
	case "doc":
		return []string{l.otherID}
	case "olic":
		return []string{}
	case "& (":
		return l.allIDs
	case "004":
		if l.matchDigits {
			return []string{l.policyID}
		}

		return []string{}
	case "001":
		if l.referenceID != "" {
			return []string{l.referenceID}
		}

		return []string{}
	default:
		t.Fatalf("unexpected search query %q", query)

		return nil
	}
}

func assertSearchHits(t *testing.T, got searchConnection, want []string) {
	t.Helper()

	ids := make([]string, len(got.Edges))
	for i, edge := range got.Edges {
		ids[i] = edge.Node.ID
	}

	assert.ElementsMatch(t, want, ids)
	assert.Equal(t, len(want), got.TotalCount)
}

const listDocumentsByQuery = `
	query($id: ID!, $query: String) {
		node(id: $id) {
			... on Organization {
				documents(first: 20, filter: { query: $query }) {
					totalCount
					edges { node { id } }
				}
			}
		}
	}
`

const listControlsByQuery = `
	query($id: ID!, $query: String) {
		node(id: $id) {
			... on Organization {
				controls(first: 20, filter: { query: $query }) {
					totalCount
					edges { node { id } }
				}
			}
		}
	}
`

const listRisksByQuery = `
	query($id: ID!, $query: String) {
		node(id: $id) {
			... on Organization {
				risks(first: 20, filter: { query: $query }) {
					totalCount
					edges { node { id } }
				}
			}
		}
	}
`

const listInternalControlsByQuery = `
	query($id: ID!, $query: String) {
		node(id: $id) {
			... on Organization {
				internalControls(first: 20, filter: { query: $query }) {
					totalCount
					edges { node { id } }
				}
			}
		}
	}
`

const listDevicesByQuery = `
	query($id: ID!, $query: String) {
		node(id: $id) {
			... on Organization {
				devices(first: 20, filter: { query: $query }) {
					totalCount
					edges { node { id } }
				}
			}
		}
	}
`

const listInternalControlsAsOfByQuery = `
	query($id: ID!, $query: String, $asOf: Datetime) {
		node(id: $id) {
			... on TreatmentPlan {
				internalControls(first: 20, asOf: $asOf, filter: { query: $query }) {
					totalCount
					edges { node { id } }
				}
			}
		}
	}
`

const riskReferenceIDQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Risk { referenceId }
		}
	}
`

func searchDocuments(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	query string,
) searchConnection {
	t.Helper()

	var result struct {
		Node struct {
			Documents searchConnection `json:"documents"`
		} `json:"node"`
	}
	client.MustExecute(listDocumentsByQuery, map[string]any{
		"id":    organizationID,
		"query": query,
	}, &result)

	return result.Node.Documents
}

func searchControls(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	query string,
) searchConnection {
	t.Helper()

	var result struct {
		Node struct {
			Controls searchConnection `json:"controls"`
		} `json:"node"`
	}
	client.MustExecute(listControlsByQuery, map[string]any{
		"id":    organizationID,
		"query": query,
	}, &result)

	return result.Node.Controls
}

func searchRisks(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	query string,
) searchConnection {
	t.Helper()

	var result struct {
		Node struct {
			Risks searchConnection `json:"risks"`
		} `json:"node"`
	}
	client.MustExecute(listRisksByQuery, map[string]any{
		"id":    organizationID,
		"query": query,
	}, &result)

	return result.Node.Risks
}

func searchInternalControls(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	query string,
) searchConnection {
	t.Helper()

	var result struct {
		Node struct {
			InternalControls searchConnection `json:"internalControls"`
		} `json:"node"`
	}
	client.MustExecute(listInternalControlsByQuery, map[string]any{
		"id":    organizationID,
		"query": query,
	}, &result)

	return result.Node.InternalControls
}

func searchDevices(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	query string,
) searchConnection {
	t.Helper()

	var result struct {
		Node struct {
			Devices searchConnection `json:"devices"`
		} `json:"node"`
	}
	client.MustExecute(listDevicesByQuery, map[string]any{
		"id":    organizationID,
		"query": query,
	}, &result)

	return result.Node.Devices
}

func searchInternalControlsAsOf(
	t *testing.T,
	client *testutil.Client,
	treatmentPlanID string,
	asOf string,
	query string,
) searchConnection {
	t.Helper()

	var result struct {
		Node struct {
			InternalControls searchConnection `json:"internalControls"`
		} `json:"node"`
	}
	client.MustExecute(listInternalControlsAsOfByQuery, map[string]any{
		"id":    treatmentPlanID,
		"query": query,
		"asOf":  asOf,
	}, &result)

	return result.Node.InternalControls
}

func riskReferenceID(t *testing.T, client *testutil.Client, riskID string) string {
	t.Helper()

	var result struct {
		Node struct {
			ReferenceID string `json:"referenceId"`
		} `json:"node"`
	}
	client.MustExecute(riskReferenceIDQuery, map[string]any{"id": riskID}, &result)
	require.NotEmpty(t, result.Node.ReferenceID)

	return result.Node.ReferenceID
}
