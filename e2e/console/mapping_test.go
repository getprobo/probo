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

func TestControlMeasureMapping_CreateDelete(t *testing.T) {
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

	// Create a measure
	var createMeasureResult struct {
		CreateMeasure struct {
			MeasureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"measureEdge"`
		} `json:"createMeasure"`
	}

	err = owner.Execute(`
		mutation($input: CreateMeasureInput!) {
			createMeasure(input: $input) {
				measureEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Measure for Mapping",
			"category":       "POLICY",
		},
	}, &createMeasureResult)
	require.NoError(t, err)

	measureID := createMeasureResult.CreateMeasure.MeasureEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		var result struct {
			CreateControlMeasureMapping struct {
				ControlEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"controlEdge"`
				MeasureEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"measureEdge"`
			} `json:"createControlMeasureMapping"`
		}

		err := owner.Execute(`
			mutation($input: CreateControlMeasureMappingInput!) {
				createControlMeasureMapping(input: $input) {
					controlEdge {
						node {
							id
						}
					}
					measureEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"measureId": measureID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, controlID, result.CreateControlMeasureMapping.ControlEdge.Node.ID)
		assert.Equal(t, measureID, result.CreateControlMeasureMapping.MeasureEdge.Node.ID)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteControlMeasureMappingInput!) {
				deleteControlMeasureMapping(input: $input) {
					deletedControlId
					deletedMeasureId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"measureId": measureID,
			},
		})
		require.NoError(t, err)
	})
}

func TestRiskMeasureMapping_CreateDelete(t *testing.T) {
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

	// Create a measure
	var createMeasureResult struct {
		CreateMeasure struct {
			MeasureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"measureEdge"`
		} `json:"createMeasure"`
	}

	err = owner.Execute(`
		mutation($input: CreateMeasureInput!) {
			createMeasure(input: $input) {
				measureEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Measure for Risk Mapping",
			"category":       "TECHNICAL",
		},
	}, &createMeasureResult)
	require.NoError(t, err)

	measureID := createMeasureResult.CreateMeasure.MeasureEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		var result struct {
			CreateRiskMeasureMapping struct {
				RiskEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"riskEdge"`
				MeasureEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"measureEdge"`
			} `json:"createRiskMeasureMapping"`
		}

		err := owner.Execute(`
			mutation($input: CreateRiskMeasureMappingInput!) {
				createRiskMeasureMapping(input: $input) {
					riskEdge {
						node {
							id
						}
					}
					measureEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":    riskID,
				"measureId": measureID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, riskID, result.CreateRiskMeasureMapping.RiskEdge.Node.ID)
		assert.Equal(t, measureID, result.CreateRiskMeasureMapping.MeasureEdge.Node.ID)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteRiskMeasureMappingInput!) {
				deleteRiskMeasureMapping(input: $input) {
					deletedRiskId
					deletedMeasureId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":    riskID,
				"measureId": measureID,
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

func TestMeasureDocumentMapping_CreateDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureID := factory.NewMeasure(owner).Create()

	t.Run("create mapping", func(t *testing.T) {
		t.Parallel()

		documentID := factory.NewDocument(owner).Create()

		var result struct {
			CreateMeasureDocumentMapping struct {
				MeasureEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"measureEdge"`
				DocumentEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"documentEdge"`
			} `json:"createMeasureDocumentMapping"`
		}

		err := owner.Execute(`
			mutation($input: CreateMeasureDocumentMappingInput!) {
				createMeasureDocumentMapping(input: $input) {
					measureEdge {
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
				"measureId":  measureID,
				"documentId": documentID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, measureID, result.CreateMeasureDocumentMapping.MeasureEdge.Node.ID)
		assert.Equal(t, documentID, result.CreateMeasureDocumentMapping.DocumentEdge.Node.ID)
	})

	t.Run("delete mapping", func(t *testing.T) {
		t.Parallel()

		documentID := factory.NewDocument(owner).Create()

		// Create the mapping first
		_, err := owner.Do(`
			mutation($input: CreateMeasureDocumentMappingInput!) {
				createMeasureDocumentMapping(input: $input) {
					documentEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"measureId":  measureID,
				"documentId": documentID,
			},
		})
		require.NoError(t, err)

		// Delete it
		_, err = owner.Do(`
			mutation($input: DeleteMeasureDocumentMappingInput!) {
				deleteMeasureDocumentMapping(input: $input) {
					deletedMeasureId
					deletedDocumentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"measureId":  measureID,
				"documentId": documentID,
			},
		})
		require.NoError(t, err)
	})
}

// The mapping mutations below link two independently-authored resources
// (e.g. controlId + measureId) together. Each is only safe because the
// underlying service loads BOTH ids in the caller's own scope before
// upserting the junction row (see e.g. ControlService.CreateMeasureMapping);
// an attacker supplying a valid GID from another organization on either
// side must be rejected. These tests pin that invariant for every mapping
// mutation.

func TestControlMeasureMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)
	org2MeasureID := factory.NewMeasure(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateControlMeasureMappingInput!) {
			createControlMeasureMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId": controlID,
			"measureId": org2MeasureID,
		},
	})
	require.Error(t, err, "must not link a control to a measure from another organization")
}

func TestRiskMeasureMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.CreateRisk(org1Owner)
	org2MeasureID := factory.NewMeasure(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateRiskMeasureMappingInput!) {
			createRiskMeasureMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":    riskID,
			"measureId": org2MeasureID,
		},
	})
	require.Error(t, err, "must not link a risk to a measure from another organization")
}

func TestControlDocumentMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)
	org2DocumentID := factory.NewDocument(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateControlDocumentMappingInput!) {
			createControlDocumentMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId":  controlID,
			"documentId": org2DocumentID,
		},
	})
	require.Error(t, err, "must not link a control to a document from another organization")
}

func TestControlAuditMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)
	org2FrameworkID := factory.CreateFramework(org2Owner)
	org2AuditID := factory.CreateAudit(org2Owner, org2FrameworkID)

	_, err := org1Owner.Do(`
		mutation($input: CreateControlAuditMappingInput!) {
			createControlAuditMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId": controlID,
			"auditId":   org2AuditID,
		},
	})
	require.Error(t, err, "must not link a control to an audit from another organization")
}

func TestRiskDocumentMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.CreateRisk(org1Owner)
	org2DocumentID := factory.NewDocument(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateRiskDocumentMappingInput!) {
			createRiskDocumentMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":     riskID,
			"documentId": org2DocumentID,
		},
	})
	require.Error(t, err, "must not link a risk to a document from another organization")
}

func TestRiskObligationMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.CreateRisk(org1Owner)

	org2ProfileID := factory.CreateUser(org2Owner)

	var createObligationResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := org2Owner.Execute(`
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge {
					node { id }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org2Owner.GetOrganizationID().String(),
			"area":           "Risk Management",
			"requirement":    "Org2 Obligation for Mapping Isolation",
			"ownerId":        org2ProfileID,
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createObligationResult)
	require.NoError(t, err)

	org2ObligationID := createObligationResult.CreateObligation.ObligationEdge.Node.ID

	_, err = org1Owner.Do(`
		mutation($input: CreateRiskObligationMappingInput!) {
			createRiskObligationMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":       riskID,
			"obligationId": org2ObligationID,
		},
	})
	require.Error(t, err, "must not link a risk to an obligation from another organization")
}

func TestMeasureDocumentMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	measureID := factory.NewMeasure(org1Owner).Create()
	org2DocumentID := factory.NewDocument(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateMeasureDocumentMappingInput!) {
			createMeasureDocumentMapping(input: $input) {
				measureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"measureId":  measureID,
			"documentId": org2DocumentID,
		},
	})
	require.Error(t, err, "must not link a measure to a document from another organization")
}
