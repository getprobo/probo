// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			"frameworkId":  frameworkID,
			"name":         "Control for Mapping",
			"description":  "Test control for mapping",
			"sectionTitle": "Section 1",
			"bestPractice": true,
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
			"frameworkId":  frameworkID,
			"name":         "Control for Document Mapping",
			"description":  "Test control",
			"sectionTitle": "Section 1",
			"bestPractice": true,
		},
	}, &createControlResult)
	require.NoError(t, err)
	controlID := createControlResult.CreateControl.ControlEdge.Node.ID

	// Create a document
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
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
			"content":        "Document content",
			"approverIds":    []string{profileID.String()},
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
			"frameworkId":  frameworkID,
			"name":         "Control for Audit Mapping",
			"description":  "Test control",
			"sectionTitle": "Section 1",
			"bestPractice": true,
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

func TestControlSnapshotMapping_CreateDelete(t *testing.T) {
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
			"name":           "Framework for ControlSnapshot Mapping",
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
			"frameworkId":  frameworkID,
			"name":         "Control for Snapshot Mapping",
			"description":  "Test control",
			"sectionTitle": "Section 1",
			"bestPractice": true,
		},
	}, &createControlResult)
	require.NoError(t, err)
	controlID := createControlResult.CreateControl.ControlEdge.Node.ID

	// Create a snapshot
	var createSnapshotResult struct {
		CreateSnapshot struct {
			SnapshotEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"snapshotEdge"`
		} `json:"createSnapshot"`
	}
	err = owner.Execute(`
		mutation($input: CreateSnapshotInput!) {
			createSnapshot(input: $input) {
				snapshotEdge {
					node {
						id
					}
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Snapshot for Control Mapping",
			"type":           "RISKS",
		},
	}, &createSnapshotResult)
	require.NoError(t, err)
	snapshotID := createSnapshotResult.CreateSnapshot.SnapshotEdge.Node.ID

	t.Run("create mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: CreateControlSnapshotMappingInput!) {
				createControlSnapshotMapping(input: $input) {
					controlEdge {
						node {
							id
						}
					}
					snapshotEdge {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId":  controlID,
				"snapshotId": snapshotID,
			},
		})
		require.NoError(t, err)
	})

	t.Run("delete mapping", func(t *testing.T) {
		_, err := owner.Do(`
			mutation($input: DeleteControlSnapshotMappingInput!) {
				deleteControlSnapshotMapping(input: $input) {
					deletedControlId
					deletedSnapshotId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId":  controlID,
				"snapshotId": snapshotID,
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
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
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
			"content":        "Document content",
			"approverIds":    []string{profileID.String()},
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
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
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
			"ownerId":        profileID.String(),
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
