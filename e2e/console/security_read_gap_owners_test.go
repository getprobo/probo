// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

// These tests extend the GHSA-c74x-79w6-63jh read-path regression coverage to
// every parent-authorized Profile field resolver. The advisory's confirmed
// exploit #2 disclosed another organization's person PII through
// processingActivity.dataProtectionOfficer; the same wrong-object
// authorization shape (authorizing the parent obj.ID with the child's
// ActionMembershipProfileGet, then loading the child through the scope-by-key
// Profile dataloader) also existed on asset.owner, datum.owner, finding.owner,
// obligation.owner, risk.owner, task.assignedTo, thirdParty.businessOwner and
// thirdParty.securityOwner. Each of those write paths validates the owner FK
// today, so these tests use injectCrossTenantFK to plant a foreign profile id
// directly in the row -- proving the read resolver now authorizes the actual
// child profile id and refuses cross-tenant PII independently of the write
// check (a future write regression, migration bug, or direct DB access).
package console_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestSecurity_ReadGap_AssetOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ProfileID := factory.CreateUser(org1Owner)
	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Asset Owner (read-gap probe)"})

	var createResult struct {
		CreateAsset struct {
			AssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"assetEdge"`
		} `json:"createAsset"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":  org1Owner.GetOrganizationID().String(),
			"name":            "Org1 Asset for read-gap probe",
			"amount":          1,
			"ownerId":         org1ProfileID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "Test data",
		},
	}, &createResult)
	require.NoError(t, err)

	assetID := createResult.CreateAsset.AssetEdge.Node.ID

	injectCrossTenantFK(t, "assets", "owner_profile_id", assetID, org2ProfileID)

	var readResult struct {
		Node struct {
			Owner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"owner"`
		} `json:"node"`
	}

	err = org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Asset {
					owner { id fullName }
				}
			}
		}
	`, map[string]any{"id": assetID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Owner == nil, "cross-tenant profile PII via asset.owner")
}

func TestSecurity_ReadGap_DatumOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ProfileID := factory.CreateUser(org1Owner)
	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Data Owner (read-gap probe)"})

	datumID := factory.CreateDatum(org1Owner, org1ProfileID, factory.Attrs{"name": "Org1 Datum for read-gap probe"})

	injectCrossTenantFK(t, "data", "owner_profile_id", datumID, org2ProfileID)

	var readResult struct {
		Node struct {
			Owner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"owner"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Datum {
					owner { id fullName }
				}
			}
		}
	`, map[string]any{"id": datumID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Owner == nil, "cross-tenant profile PII via datum.owner")
}

func TestSecurity_ReadGap_FindingOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ProfileID := factory.CreateUser(org1Owner)
	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Finding Owner (read-gap probe)"})

	var createResult struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org1Owner.GetOrganizationID().String(),
			"kind":           "OBSERVATION",
			"status":         "OPEN",
			"priority":       "LOW",
			"ownerId":        org1ProfileID,
		},
	}, &createResult)
	require.NoError(t, err)

	findingID := createResult.CreateFinding.FindingEdge.Node.ID

	injectCrossTenantFK(t, "findings", "owner_id", findingID, org2ProfileID)

	var readResult struct {
		Node struct {
			Owner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"owner"`
		} `json:"node"`
	}

	err = org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Finding {
					owner { id fullName }
				}
			}
		}
	`, map[string]any{"id": findingID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Owner == nil, "cross-tenant profile PII via finding.owner")
}

func TestSecurity_ReadGap_ObligationOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ProfileID := factory.CreateUser(org1Owner)
	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Obligation Owner (read-gap probe)"})

	var createResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org1Owner.GetOrganizationID().String(),
			"area":           "Data Protection",
			"source":         "GDPR Article 5",
			"requirement":    "Org1 obligation for read-gap probe",
			"ownerId":        org1ProfileID,
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createResult)
	require.NoError(t, err)

	obligationID := createResult.CreateObligation.ObligationEdge.Node.ID

	injectCrossTenantFK(t, "obligations", "owner_profile_id", obligationID, org2ProfileID)

	var readResult struct {
		Node struct {
			Owner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"owner"`
		} `json:"node"`
	}

	err = org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Obligation {
					owner { id fullName }
				}
			}
		}
	`, map[string]any{"id": obligationID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Owner == nil, "cross-tenant profile PII via obligation.owner")
}

func TestSecurity_ReadGap_RiskOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Risk Owner (read-gap probe)"})

	riskID := factory.CreateRisk(org1Owner, factory.Attrs{"name": "Org1 Risk for read-gap probe"})

	injectCrossTenantFK(t, "risks", "owner_profile_id", riskID, org2ProfileID)

	var readResult struct {
		Node struct {
			Owner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"owner"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Risk {
					owner { id fullName }
				}
			}
		}
	`, map[string]any{"id": riskID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Owner == nil, "cross-tenant profile PII via risk.owner")
}

func TestSecurity_ReadGap_TaskAssignedTo(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Assignee (read-gap probe)"})

	taskID := factory.CreateTask(org1Owner, nil, factory.Attrs{"name": "Org1 Task for read-gap probe"})

	injectCrossTenantFK(t, "tasks", "assigned_to_profile_id", taskID, org2ProfileID)

	var readResult struct {
		Node struct {
			AssignedTo *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"assignedTo"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					assignedTo { id fullName }
				}
			}
		}
	`, map[string]any{"id": taskID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.AssignedTo == nil, "cross-tenant profile PII via task.assignedTo")
}

func TestSecurity_ReadGap_ThirdPartyBusinessOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Business Owner (read-gap probe)"})

	thirdPartyID := factory.CreateThirdParty(org1Owner, factory.Attrs{"name": "Org1 ThirdParty for read-gap probe"})

	injectCrossTenantFK(t, "third_parties", "business_owner_profile_id", thirdPartyID, org2ProfileID)

	var readResult struct {
		Node struct {
			BusinessOwner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"businessOwner"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					businessOwner { id fullName }
				}
			}
		}
	`, map[string]any{"id": thirdPartyID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.BusinessOwner == nil, "cross-tenant profile PII via thirdParty.businessOwner")
}

func TestSecurity_ReadGap_ThirdPartySecurityOwner(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Security Owner (read-gap probe)"})

	thirdPartyID := factory.CreateThirdParty(org1Owner, factory.Attrs{"name": "Org1 ThirdParty for read-gap probe"})

	injectCrossTenantFK(t, "third_parties", "security_owner_profile_id", thirdPartyID, org2ProfileID)

	var readResult struct {
		Node struct {
			SecurityOwner *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"securityOwner"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					securityOwner { id fullName }
				}
			}
		}
	`, map[string]any{"id": thirdPartyID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.SecurityOwner == nil, "cross-tenant profile PII via thirdParty.securityOwner")
}
