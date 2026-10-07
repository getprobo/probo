// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestControlInternalControlMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a framework
	var createFrameworkResult struct {
		CreateFramework struct {
			FrameworkEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"frameworkEdge"`
		} `json:"createFramework"`
	}

	err := owner.Execute(`
		mutation($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Framework for Mapping",
		},
	}, &createFrameworkResult)
	require.NoError(t, err)

	frameworkID := createFrameworkResult.CreateFramework.FrameworkEdge.Node.ID

	// Create a control
	var createControlResult struct {
		CreateControl struct {
			ControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControl"`
	}

	err = owner.Execute(`
		mutation($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"frameworkId":   frameworkID,
			"name":          "Control for Mapping",
			"description":   "Test control for mapping",
			"sectionTitle":  "Section 1",
			"bestPractice":  true,
			"maturityLevel": "INITIAL",
		},
	}, &createControlResult)
	require.NoError(t, err)

	controlID := createControlResult.CreateControl.ControlEdge.Node.ID

	// Create an internal control
	var createInternalControlResult struct {
		CreateInternalControl struct {
			InternalControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"internalControlEdge"`
		} `json:"createInternalControl"`
	}

	err = owner.Execute(`
		mutation($input: CreateInternalControlInput!) {
			createInternalControl(input: $input) {
				internalControlEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Internal control for Mapping",
			"category":       "POLICY",
		},
	}, &createInternalControlResult)
	require.NoError(t, err)

	internalControlID := createInternalControlResult.CreateInternalControl.InternalControlEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		var result struct {
			CreateControlInternalControlMapping struct {
				ControlEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"controlEdge"`
				InternalControlEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"internalControlEdge"`
			} `json:"createControlInternalControlMapping"`
		}

		err := owner.Execute(`
			mutation($input: CreateControlInternalControlMappingInput!) {
				createControlInternalControlMapping(input: $input) {
					controlEdge {
						node {
							id
						}
					}
					internalControlEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId":         controlID,
				"internalControlId": internalControlID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, controlID, result.CreateControlInternalControlMapping.ControlEdge.Node.ID)
		assert.Equal(t, internalControlID, result.CreateControlInternalControlMapping.InternalControlEdge.Node.ID)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteControlInternalControlMappingInput!) {
				deleteControlInternalControlMapping(input: $input) {
					deletedControlId
					deletedInternalControlId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId":         controlID,
				"internalControlId": internalControlID,
			},
		})
		require.NoError(t, err)
	})
}

func TestRiskInternalControlMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a risk
	var createRiskResult struct {
		CreateRisk struct {
			RiskEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskEdge"`
		} `json:"createRisk"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskInput!) {
			createRisk(input: $input) {
				riskEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"name":               "Risk for Mapping",
			"category":           "Operational",
			"treatment":          "MITIGATED",
			"inherentLikelihood": 3,
			"inherentImpact":     3,
		},
	}, &createRiskResult)
	require.NoError(t, err)

	riskID := createRiskResult.CreateRisk.RiskEdge.Node.ID

	// Create an internal control
	var createInternalControlResult struct {
		CreateInternalControl struct {
			InternalControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"internalControlEdge"`
		} `json:"createInternalControl"`
	}

	err = owner.Execute(`
		mutation($input: CreateInternalControlInput!) {
			createInternalControl(input: $input) {
				internalControlEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Internal control for Risk Mapping",
			"category":       "TECHNICAL",
		},
	}, &createInternalControlResult)
	require.NoError(t, err)

	internalControlID := createInternalControlResult.CreateInternalControl.InternalControlEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		var result struct {
			CreateRiskInternalControlMapping struct {
				RiskEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"riskEdge"`
				InternalControlEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"internalControlEdge"`
			} `json:"createRiskInternalControlMapping"`
		}

		err := owner.Execute(`
			mutation($input: CreateRiskInternalControlMappingInput!) {
				createRiskInternalControlMapping(input: $input) {
					riskEdge {
						node {
							id
						}
					}
					internalControlEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":            riskID,
				"internalControlId": internalControlID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, riskID, result.CreateRiskInternalControlMapping.RiskEdge.Node.ID)
		assert.Equal(t, internalControlID, result.CreateRiskInternalControlMapping.InternalControlEdge.Node.ID)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteRiskInternalControlMappingInput!) {
				deleteRiskInternalControlMapping(input: $input) {
					deletedRiskId
					deletedInternalControlId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":            riskID,
				"internalControlId": internalControlID,
			},
		})
		require.NoError(t, err)
	})
}

func TestControlDocumentMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a framework and control
	var createFrameworkResult struct {
		CreateFramework struct {
			FrameworkEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"frameworkEdge"`
		} `json:"createFramework"`
	}

	err := owner.Execute(`
		mutation($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Framework for ControlDoc Mapping",
		},
	}, &createFrameworkResult)
	require.NoError(t, err)

	frameworkID := createFrameworkResult.CreateFramework.FrameworkEdge.Node.ID

	var createControlResult struct {
		CreateControl struct {
			ControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControl"`
	}

	err = owner.Execute(`
		mutation($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"frameworkId":   frameworkID,
			"name":          "Control for Document Mapping",
			"description":   "Test control",
			"sectionTitle":  "Section 1",
			"bestPractice":  true,
			"maturityLevel": "INITIAL",
		},
	}, &createControlResult)
	require.NoError(t, err)

	controlID := createControlResult.CreateControl.ControlEdge.Node.ID

	// Create a document
	var createDocumentResult struct {
		CreateDocument struct {
			DocumentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentEdge"`
		} `json:"createDocument"`
	}

	err = owner.Execute(`
		mutation($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"title":          "Document for Control Mapping",
			"content":        testutil.ProseMirrorTextDoc("Document content"),
			"documentType":   "POLICY",
			"classification": "INTERNAL",
		},
	}, &createDocumentResult)
	require.NoError(t, err)

	documentID := createDocumentResult.CreateDocument.DocumentEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: CreateControlDocumentMappingInput!) {
				createControlDocumentMapping(input: $input) {
					controlEdge {
						node {
							id
						}
					}
					documentEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId":  controlID,
				"documentId": documentID,
			},
		})
		require.NoError(t, err)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteControlDocumentMappingInput!) {
				deleteControlDocumentMapping(input: $input) {
					deletedControlId
					deletedDocumentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId":  controlID,
				"documentId": documentID,
			},
		})
		require.NoError(t, err)
	})
}

func TestControlAuditMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a framework and control
	var createFrameworkResult struct {
		CreateFramework struct {
			FrameworkEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"frameworkEdge"`
		} `json:"createFramework"`
	}

	err := owner.Execute(`
		mutation($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Framework for ControlAudit Mapping",
		},
	}, &createFrameworkResult)
	require.NoError(t, err)

	frameworkID := createFrameworkResult.CreateFramework.FrameworkEdge.Node.ID

	var createControlResult struct {
		CreateControl struct {
			ControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControl"`
	}

	err = owner.Execute(`
		mutation($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"frameworkId":   frameworkID,
			"name":          "Control for Audit Mapping",
			"description":   "Test control",
			"sectionTitle":  "Section 1",
			"bestPractice":  true,
			"maturityLevel": "INITIAL",
		},
	}, &createControlResult)
	require.NoError(t, err)

	controlID := createControlResult.CreateControl.ControlEdge.Node.ID

	// Create an audit
	var createAuditResult struct {
		CreateAudit struct {
			AuditEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"auditEdge"`
		} `json:"createAudit"`
	}

	err = owner.Execute(`
		mutation($input: CreateAuditInput!) {
			createAudit(input: $input) {
				auditEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"frameworkId":    frameworkID,
			"name":           "Audit for Control Mapping",
		},
	}, &createAuditResult)
	require.NoError(t, err)

	auditID := createAuditResult.CreateAudit.AuditEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: CreateControlAuditMappingInput!) {
				createControlAuditMapping(input: $input) {
					controlEdge {
						node {
							id
						}
					}
					auditEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"auditId":   auditID,
			},
		})
		require.NoError(t, err)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteControlAuditMappingInput!) {
				deleteControlAuditMapping(input: $input) {
					deletedControlId
					deletedAuditId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"auditId":   auditID,
			},
		})
		require.NoError(t, err)
	})
}

func TestRiskDocumentMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a risk
	var createRiskResult struct {
		CreateRisk struct {
			RiskEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskEdge"`
		} `json:"createRisk"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskInput!) {
			createRisk(input: $input) {
				riskEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"name":               "Risk for Document Mapping",
			"category":           "Operational",
			"treatment":          "MITIGATED",
			"inherentLikelihood": 3,
			"inherentImpact":     3,
		},
	}, &createRiskResult)
	require.NoError(t, err)

	riskID := createRiskResult.CreateRisk.RiskEdge.Node.ID

	// Create a document
	var createDocumentResult struct {
		CreateDocument struct {
			DocumentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentEdge"`
		} `json:"createDocument"`
	}

	err = owner.Execute(`
		mutation($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"title":          "Document for Risk Mapping",
			"content":        testutil.ProseMirrorTextDoc("Document content"),
			"documentType":   "POLICY",
			"classification": "INTERNAL",
		},
	}, &createDocumentResult)
	require.NoError(t, err)

	documentID := createDocumentResult.CreateDocument.DocumentEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: CreateRiskDocumentMappingInput!) {
				createRiskDocumentMapping(input: $input) {
					riskEdge {
						node {
							id
						}
					}
					documentEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":     riskID,
				"documentId": documentID,
			},
		})
		require.NoError(t, err)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteRiskDocumentMappingInput!) {
				deleteRiskDocumentMapping(input: $input) {
					deletedRiskId
					deletedDocumentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":     riskID,
				"documentId": documentID,
			},
		})
		require.NoError(t, err)
	})
}

func loadRiskLinkedDocumentIDs(
	t *testing.T,
	owner *testutil.Client,
	riskID string,
) []string {
	t.Helper()

	var result struct {
		Node struct {
			Documents struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"documents"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Risk {
					documents(first: 10) {
						edges {
							node {
								id
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": riskID}, &result)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Node.Documents.Edges))
	for _, edge := range result.Node.Documents.Edges {
		ids = append(ids, edge.Node.ID)
	}

	return ids
}

func TestRiskDocumentMapping_DeleteDocumentClearsRiskDocumentLink(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.NewRisk(owner).
		WithName("Risk linked to deleted document").
		Create()
	documentID := factory.NewDocument(owner).
		WithTitle("Document linked to risk").
		Create()

	_, err := owner.Do(`
		mutation($input: CreateRiskDocumentMappingInput!) {
			createRiskDocumentMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":     riskID,
			"documentId": documentID,
		},
	})
	require.NoError(t, err)

	linkedDocumentIDs := loadRiskLinkedDocumentIDs(t, owner, riskID)
	require.Contains(t, linkedDocumentIDs, documentID)

	_, err = owner.Do(`
		mutation DeleteDocument($input: DeleteDocumentInput!) {
			deleteDocument(input: $input) {
				deletedDocumentId
			}
		}
	`, map[string]any{
		"input": map[string]any{"documentId": documentID},
	})
	require.NoError(t, err)

	linkedDocumentIDs = loadRiskLinkedDocumentIDs(t, owner, riskID)
	assert.NotContains(t, linkedDocumentIDs, documentID)
	assert.Empty(t, linkedDocumentIDs)

	_, err = owner.Do(`
		mutation DeleteRisk($input: DeleteRiskInput!) {
			deleteRisk(input: $input) {
				deletedRiskId
			}
		}
	`, map[string]any{
		"input": map[string]any{"riskId": riskID},
	})
	require.NoError(t, err)
}

func TestRiskDocumentMapping_DeleteRiskRequiresUnlinkedDocuments(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.NewRisk(owner).
		WithName("Risk with active document link").
		Create()
	documentID := factory.NewDocument(owner).
		WithTitle("Document still linked to risk").
		Create()

	_, err := owner.Do(`
		mutation($input: CreateRiskDocumentMappingInput!) {
			createRiskDocumentMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":     riskID,
			"documentId": documentID,
		},
	})
	require.NoError(t, err)

	_, err = owner.Do(`
		mutation DeleteRisk($input: DeleteRiskInput!) {
			deleteRisk(input: $input) {
				deletedRiskId
			}
		}
	`, map[string]any{
		"input": map[string]any{"riskId": riskID},
	})
	require.Error(t, err)

	_, err = owner.Do(`
		mutation($input: DeleteRiskDocumentMappingInput!) {
			deleteRiskDocumentMapping(input: $input) {
				deletedRiskId
				deletedDocumentId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":     riskID,
			"documentId": documentID,
		},
	})
	require.NoError(t, err)

	_, err = owner.Do(`
		mutation DeleteRisk($input: DeleteRiskInput!) {
			deleteRisk(input: $input) {
				deletedRiskId
			}
		}
	`, map[string]any{
		"input": map[string]any{"riskId": riskID},
	})
	require.NoError(t, err)
}

func TestRiskObligationMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a risk
	var createRiskResult struct {
		CreateRisk struct {
			RiskEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskEdge"`
		} `json:"createRisk"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskInput!) {
			createRisk(input: $input) {
				riskEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"name":               "Risk for Obligation Mapping",
			"category":           "Compliance",
			"treatment":          "MITIGATED",
			"inherentLikelihood": 2,
			"inherentImpact":     4,
		},
	}, &createRiskResult)
	require.NoError(t, err)

	riskID := createRiskResult.CreateRisk.RiskEdge.Node.ID

	// Create an obligation
	profileID := factory.CreateUser(owner)

	var createObligationResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err = owner.Execute(`
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"area":           "Risk Management",
			"requirement":    "Obligation for Risk Mapping",
			"ownerId":        profileID,
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createObligationResult)
	require.NoError(t, err)

	obligationID := createObligationResult.CreateObligation.ObligationEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: CreateRiskObligationMappingInput!) {
				createRiskObligationMapping(input: $input) {
					riskEdge {
						node {
							id
						}
					}
					obligationEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":       riskID,
				"obligationId": obligationID,
			},
		})
		require.NoError(t, err)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteRiskObligationMappingInput!) {
				deleteRiskObligationMapping(input: $input) {
					deletedRiskId
					deletedObligationId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":       riskID,
				"obligationId": obligationID,
			},
		})
		require.NoError(t, err)
	})
}

func TestInternalControlDocumentMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()

	t.Run("create mapping", func(t *testing.T) {
		t.Parallel()

		documentID := factory.NewDocument(owner).Create()

		var result struct {
			CreateInternalControlDocumentMapping struct {
				InternalControlEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"internalControlEdge"`
				DocumentEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"documentEdge"`
			} `json:"createInternalControlDocumentMapping"`
		}

		err := owner.Execute(`
			mutation($input: CreateInternalControlDocumentMappingInput!) {
				createInternalControlDocumentMapping(input: $input) {
					internalControlEdge {
						node {
							id
						}
					}
					documentEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"documentId":        documentID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, internalControlID, result.CreateInternalControlDocumentMapping.InternalControlEdge.Node.ID)
		assert.Equal(t, documentID, result.CreateInternalControlDocumentMapping.DocumentEdge.Node.ID)
	})

	t.Run("delete mapping", func(t *testing.T) {
		t.Parallel()

		documentID := factory.NewDocument(owner).Create()

		// Create the mapping first
		_, err := owner.Do(`
			mutation($input: CreateInternalControlDocumentMappingInput!) {
				createInternalControlDocumentMapping(input: $input) {
					documentEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"documentId":        documentID,
			},
		})
		require.NoError(t, err)

		// Delete it
		_, err = owner.Do(`
			mutation($input: DeleteInternalControlDocumentMappingInput!) {
				deleteInternalControlDocumentMapping(input: $input) {
					deletedInternalControlId
					deletedDocumentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"documentId":        documentID,
			},
		})
		require.NoError(t, err)
	})
}

// The mapping mutations below link two independently-authored resources
// (e.g. controlId + internalControlId) together. Each is only safe because the
// underlying service loads BOTH ids in the caller's own scope before
// upserting the junction row (see e.g. ControlService.CreateInternalControlMapping);
// an attacker supplying a valid GID from another organization on either
// side must be rejected. These tests pin that invariant for every mapping
// mutation.
