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
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const internalControlControlFields = `
	id
	code
	state
	controlType
	nature
	operatingFrequency { mode interval event }
	evidenceCadence
	testingCadence
	nextEvidenceDue
	nextTestDue
	implementationStatus
	owner { id }
	reviewer { id }
`

type internalControlControlProfile struct {
	ID string `json:"id"`
}

type internalControlOperatingFrequency struct {
	Mode     string  `json:"mode"`
	Interval *string `json:"interval"`
	Event    *string `json:"event"`
}

type internalControlControlNode struct {
	ID                   string                             `json:"id"`
	Code                 *string                            `json:"code"`
	State                string                             `json:"state"`
	ControlType          *string                            `json:"controlType"`
	Nature               *string                            `json:"nature"`
	OperatingFrequency   *internalControlOperatingFrequency `json:"operatingFrequency"`
	EvidenceCadence      *string                            `json:"evidenceCadence"`
	TestingCadence       *string                            `json:"testingCadence"`
	NextEvidenceDue      *time.Time                         `json:"nextEvidenceDue"`
	NextTestDue          *time.Time                         `json:"nextTestDue"`
	ImplementationStatus string                             `json:"implementationStatus"`
	Owner                *internalControlControlProfile     `json:"owner"`
	Reviewer             *internalControlControlProfile     `json:"reviewer"`
}

func TestMeasure_InternalControlEnums(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	reviewer := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	code := strings.ReplaceAll(factory.SafeName("IC"), " ", "-")

	created := createInternalControlControl(t, owner, map[string]any{
		"name":        "Access reviews",
		"category":    "ACCESS",
		"code":        code,
		"controlType": "PREVENTIVE",
		"nature":      "MANUAL",
		"operatingFrequency": map[string]any{
			"mode":     "PERIODIC",
			"interval": "P3M",
		},
		"evidenceCadence":      "P1M",
		"testingCadence":       "P3M",
		"implementationStatus": "OPERATING",
		"ownerId":              owner.GetProfileID().String(),
		"reviewerId":           reviewer.GetProfileID().String(),
	})

	assert.Equal(t, code, derefString(t, created.Code))
	assert.Equal(t, "IMPLEMENTED", created.State)
	assert.Equal(t, "PREVENTIVE", derefString(t, created.ControlType))
	assert.Equal(t, "MANUAL", derefString(t, created.Nature))
	require.NotNil(t, created.OperatingFrequency)
	assert.Equal(t, "PERIODIC", created.OperatingFrequency.Mode)
	assert.Equal(t, "P3M", derefString(t, created.OperatingFrequency.Interval))
	assert.Nil(t, created.OperatingFrequency.Event)
	assert.Equal(t, "P1M", derefString(t, created.EvidenceCadence))
	assert.Equal(t, "P3M", derefString(t, created.TestingCadence))
	require.NotNil(t, created.NextEvidenceDue)
	require.NotNil(t, created.NextTestDue)
	assert.Equal(t, "OPERATING", created.ImplementationStatus)
	require.NotNil(t, created.Owner)
	require.NotNil(t, created.Reviewer)
	assert.Equal(t, owner.GetProfileID().String(), created.Owner.ID)
	assert.Equal(t, reviewer.GetProfileID().String(), created.Reviewer.ID)

	reloaded := getInternalControlControl(t, owner, created.ID)
	assert.Equal(t, created.ID, reloaded.ID)
	assert.Equal(t, "OPERATING", reloaded.ImplementationStatus)
	assertSameInstant(t, created.NextEvidenceDue, reloaded.NextEvidenceDue)
	assertSameInstant(t, created.NextTestDue, reloaded.NextTestDue)

	renamed := updateInternalControlControl(t, owner, map[string]any{
		"id":   created.ID,
		"name": "Access reviews renamed",
	})
	assert.Equal(t, "OPERATING", renamed.ImplementationStatus)
	assert.Equal(t, "IMPLEMENTED", renamed.State)
	assertSameInstant(t, created.NextEvidenceDue, renamed.NextEvidenceDue)

	repeated := updateInternalControlControl(t, owner, map[string]any{
		"id":                   created.ID,
		"implementationStatus": "OPERATING",
	})
	assert.Equal(t, "OPERATING", repeated.ImplementationStatus)

	eventDriven := updateInternalControlControl(t, owner, map[string]any{
		"id": created.ID,
		"operatingFrequency": map[string]any{
			"mode":  "EVENT",
			"event": "when someone leaves",
		},
	})
	require.NotNil(t, eventDriven.OperatingFrequency)
	assert.Equal(t, "EVENT", eventDriven.OperatingFrequency.Mode)
	assert.Nil(t, eventDriven.OperatingFrequency.Interval)
	assert.Equal(t, "when someone leaves", derefString(t, eventDriven.OperatingFrequency.Event))
	assertSameInstant(t, created.NextEvidenceDue, eventDriven.NextEvidenceDue)
	assertSameInstant(t, created.NextTestDue, eventDriven.NextTestDue)

	continuous := updateInternalControlControl(t, owner, map[string]any{
		"id": created.ID,
		"operatingFrequency": map[string]any{
			"mode": "CONTINUOUS",
		},
	})
	require.NotNil(t, continuous.OperatingFrequency)
	assert.Equal(t, "CONTINUOUS", continuous.OperatingFrequency.Mode)
	assert.Nil(t, continuous.OperatingFrequency.Interval)
	assert.Nil(t, continuous.OperatingFrequency.Event)
	assertSameInstant(t, created.NextEvidenceDue, continuous.NextEvidenceDue)

	sameCadence := updateInternalControlControl(t, owner, map[string]any{
		"id":              created.ID,
		"evidenceCadence": "P1M",
	})
	assertSameInstant(t, created.NextEvidenceDue, sameCadence.NextEvidenceDue)

	movedCadence := updateInternalControlControl(t, owner, map[string]any{
		"id":              created.ID,
		"evidenceCadence": "P2M",
	})
	require.NotNil(t, movedCadence.NextEvidenceDue)
	assert.False(t, movedCadence.NextEvidenceDue.Equal(*created.NextEvidenceDue))

	cleared := updateInternalControlControl(t, owner, map[string]any{
		"id":              created.ID,
		"evidenceCadence": nil,
	})
	assert.Nil(t, cleared.EvidenceCadence)
	assert.Nil(t, cleared.NextEvidenceDue)
	assertSameInstant(t, created.NextTestDue, cleared.NextTestDue)
}

func TestMeasure_InternalControlDefaults(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	created := createInternalControlControl(t, owner, map[string]any{
		"name":     "Draft control",
		"category": "ACCESS",
	})

	assert.Nil(t, created.Code)
	assert.Nil(t, created.ControlType)
	assert.Nil(t, created.Nature)
	assert.Nil(t, created.OperatingFrequency)
	assert.Nil(t, created.EvidenceCadence)
	assert.Nil(t, created.TestingCadence)
	assert.Nil(t, created.NextEvidenceDue)
	assert.Nil(t, created.NextTestDue)
	assert.Nil(t, created.Owner)
	assert.Nil(t, created.Reviewer)
	assert.Equal(t, "NOT_STARTED", created.State)
	assert.Equal(t, "NOT_IMPLEMENTED", created.ImplementationStatus)
}

func TestMeasure_NonPositiveCadenceHasNoSchedule(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	created := createInternalControlControl(t, owner, map[string]any{
		"name":            "Unscheduled control",
		"category":        "ACCESS",
		"evidenceCadence": "-P1M",
		"testingCadence":  "P0D",
	})

	assert.Nil(t, created.EvidenceCadence)
	assert.Nil(t, created.TestingCadence)
	assert.Nil(t, created.NextEvidenceDue)
	assert.Nil(t, created.NextTestDue)
}

func TestMeasure_InternalControlValidation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name    string
		input   map[string]any
		message string
	}{
		{
			name: "periodic frequency requires a duration",
			input: map[string]any{
				"operatingFrequency": map[string]any{"mode": "PERIODIC"},
			},
			message: "periodic frequency requires a duration",
		},
		{
			name: "periodic frequency has no event",
			input: map[string]any{
				"operatingFrequency": map[string]any{
					"mode":     "PERIODIC",
					"interval": "P3M",
					"event":    "when someone leaves",
				},
			},
			message: "periodic frequency has no event",
		},
		{
			name: "continuous frequency has no interval",
			input: map[string]any{
				"operatingFrequency": map[string]any{
					"mode":     "CONTINUOUS",
					"interval": "P3M",
				},
			},
			message: "continuous frequency has no interval or event",
		},
		{
			name: "event frequency has no interval",
			input: map[string]any{
				"operatingFrequency": map[string]any{
					"mode":     "EVENT",
					"interval": "P3M",
				},
			},
			message: "event frequency has no interval",
		},
		{
			name: "periodic interval must be positive",
			input: map[string]any{
				"operatingFrequency": map[string]any{
					"mode":     "PERIODIC",
					"interval": "-P1M",
				},
			},
			message: "must be a positive duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("Control"),
				"category":       "ACCESS",
			}
			maps.Copy(input, tt.input)

			err := owner.Execute(createInternalControlControlQuery, map[string]any{"input": input}, &struct{}{})
			testutil.RequireErrorCode(t, err, "INVALID")
			assert.ErrorContains(t, err, tt.message)
		})
	}

	t.Run("owner and reviewer must differ", func(t *testing.T) {
		t.Parallel()

		profileID := owner.GetProfileID().String()
		err := owner.Execute(createInternalControlControlQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("Control"),
				"category":       "ACCESS",
				"ownerId":        profileID,
				"reviewerId":     profileID,
			},
		}, &struct{}{})
		testutil.RequireErrorCode(t, err, "INVALID")
		assert.ErrorContains(t, err, "must be a different person from the owner")
	})

	t.Run("code is unique in the organization", func(t *testing.T) {
		t.Parallel()

		code := strings.ReplaceAll(factory.SafeName("IC"), " ", "-")
		createInternalControlControl(t, owner, map[string]any{
			"name":     factory.SafeName("Control"),
			"category": "ACCESS",
			"code":     code,
		})

		err := owner.Execute(createInternalControlControlQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("Control"),
				"category":       "ACCESS",
				"code":           code,
			},
		}, &struct{}{})
		testutil.RequireConflictError(t, err)
	})

	t.Run("unknown control type is rejected", func(t *testing.T) {
		t.Parallel()

		err := owner.Execute(createInternalControlControlQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("Control"),
				"category":       "ACCESS",
				"controlType":    "AUTOMATED",
			},
		}, &struct{}{})
		require.Error(t, err)
		assert.ErrorContains(t, err, "AUTOMATED")
	})
}

const createInternalControlControlQuery = `
	mutation CreateInternalControl($input: CreateInternalControlInput!) {
		createInternalControl(input: $input) {
			internalControlEdge {
				node { ` + internalControlControlFields + ` }
			}
		}
	}
`

const updateInternalControlControlQuery = `
	mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
		updateInternalControl(input: $input) {
			internalControl { ` + internalControlControlFields + ` }
		}
	}
`

const getInternalControlControlQuery = `
	query GetInternalControl($id: ID!) {
		node(id: $id) {
			... on InternalControl { ` + internalControlControlFields + ` }
		}
	}
`

func createInternalControlControl(t *testing.T, owner *testutil.Client, fields map[string]any) internalControlControlNode {
	t.Helper()

	input := map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
	}
	maps.Copy(input, fields)

	var result struct {
		CreateInternalControl struct {
			InternalControlEdge struct {
				Node internalControlControlNode `json:"node"`
			} `json:"internalControlEdge"`
		} `json:"createInternalControl"`
	}

	err := owner.Execute(createInternalControlControlQuery, map[string]any{"input": input}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.CreateInternalControl.InternalControlEdge.Node.ID)

	return result.CreateInternalControl.InternalControlEdge.Node
}

func updateInternalControlControl(t *testing.T, owner *testutil.Client, input map[string]any) internalControlControlNode {
	t.Helper()

	var result struct {
		UpdateInternalControl struct {
			InternalControl internalControlControlNode `json:"internalControl"`
		} `json:"updateInternalControl"`
	}

	err := owner.Execute(updateInternalControlControlQuery, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.UpdateInternalControl.InternalControl
}

func getInternalControlControl(t *testing.T, owner *testutil.Client, id string) internalControlControlNode {
	t.Helper()

	var result struct {
		Node internalControlControlNode `json:"node"`
	}

	err := owner.Execute(getInternalControlControlQuery, map[string]any{"id": id}, &result)
	require.NoError(t, err)
	require.Equal(t, id, result.Node.ID)

	return result.Node
}

func derefString(t *testing.T, value *string) string {
	t.Helper()
	require.NotNil(t, value)

	return *value
}

func assertSameInstant(t *testing.T, expected, actual *time.Time) {
	t.Helper()
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	assert.True(t, expected.Truncate(time.Microsecond).Equal(actual.Truncate(time.Microsecond)))
}
