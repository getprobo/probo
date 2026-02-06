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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMeeting_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name        string
		input       map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "with full details",
			input: map[string]any{
				"name":    "Security Review Meeting",
				"date":    time.Now().Format(time.RFC3339Nano),
				"minutes": "Meeting notes here",
			},
			assertField: "name",
			assertValue: "Security Review Meeting",
		},
		{
			name: "without minutes",
			input: map[string]any{
				"name": "Quick Standup",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			assertField: "name",
			assertValue: "Quick Standup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateMeeting($input: CreateMeetingInput!) {
					createMeeting(input: $input) {
						meetingEdge {
							node {
								id
								name
								date
								minutes
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
			}
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				CreateMeeting struct {
					MeetingEdge struct {
						Node struct {
							ID      string  `json:"id"`
							Name    string  `json:"name"`
							Date    string  `json:"date"`
							Minutes *string `json:"minutes"`
						} `json:"node"`
					} `json:"meetingEdge"`
				} `json:"createMeeting"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateMeeting.MeetingEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, node.Name)
			}
		})
	}
}

func TestMeeting_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name": "Test Meeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "name with HTML tags",
			input: map[string]any{
				"name": "<script>alert('xss')</script>",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "HTML",
		},
		{
			name: "name with angle brackets",
			input: map[string]any{
				"name": "Test < Meeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "angle brackets",
		},
		{
			name: "name with newline",
			input: map[string]any{
				"name": "Test\nMeeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "newline",
		},
		{
			name: "name with carriage return",
			input: map[string]any{
				"name": "Test\rMeeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "carriage return",
		},
		{
			name: "name with null byte",
			input: map[string]any{
				"name": "Test\x00Meeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with tab character",
			input: map[string]any{
				"name": "Test\tMeeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with zero-width space",
			input: map[string]any{
				"name": "Test\u200BMeeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with zero-width joiner",
			input: map[string]any{
				"name": "Test\u200DMeeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with right-to-left override",
			input: map[string]any{
				"name": "Test\u202EMeeting",
				"date": time.Now().Format(time.RFC3339Nano),
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateMeeting($input: CreateMeetingInput!) {
					createMeeting(input: $input) {
						meetingEdge {
							node {
								id
							}
						}
					}
				}
			`

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}
			for k, v := range tt.input {
				input[k] = v
			}

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeeting_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name        string
		setup       func() string
		input       func(id string) map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "update name",
			setup: func() string {
				return factory.NewMeeting(owner).
					WithName("Meeting to Update").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"meetingId": id,
					"name":      "Updated Meeting Name",
				}
			},
			assertField: "name",
			assertValue: "Updated Meeting Name",
		},
		{
			name: "update minutes",
			setup: func() string {
				return factory.NewMeeting(owner).
					WithName("Minutes Test Meeting").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"meetingId": id,
					"minutes":   "Updated meeting notes",
				}
			},
			assertField: "minutes",
			assertValue: "Updated meeting notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meetingID := tt.setup()

			query := `
				mutation UpdateMeeting($input: UpdateMeetingInput!) {
					updateMeeting(input: $input) {
						meeting {
							id
							name
							minutes
						}
					}
				}
			`

			var result struct {
				UpdateMeeting struct {
					Meeting struct {
						ID      string  `json:"id"`
						Name    string  `json:"name"`
						Minutes *string `json:"minutes"`
					} `json:"meeting"`
				} `json:"updateMeeting"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(meetingID)}, &result)
			require.NoError(t, err)

			meeting := result.UpdateMeeting.Meeting
			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, meeting.Name)
			case "minutes":
				assert.NotNil(t, meeting.Minutes)
				assert.Equal(t, tt.assertValue, *meeting.Minutes)
			}
		})
	}
}

func TestMeeting_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	baseMeetingID := factory.NewMeeting(owner).WithName("Validation Test Meeting").Create()

	tests := []struct {
		name              string
		setup             func() string
		input             func(id string) map[string]any
		wantErrorContains string
	}{
		{
			name:  "invalid ID format",
			setup: func() string { return "invalid-id-format" },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "Test"}
			},
			wantErrorContains: "base64",
		},
		{
			name:  "name with HTML tags",
			setup: func() string { return baseMeetingID },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "name with angle brackets",
			setup: func() string { return baseMeetingID },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "Test < Meeting"}
			},
			wantErrorContains: "angle brackets",
		},
		{
			name:  "name with newline",
			setup: func() string { return baseMeetingID },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "Test\nMeeting"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "name with carriage return",
			setup: func() string { return baseMeetingID },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "Test\rMeeting"}
			},
			wantErrorContains: "carriage return",
		},
		{
			name:  "name with null byte",
			setup: func() string { return baseMeetingID },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "Test\x00Meeting"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with zero-width space",
			setup: func() string { return baseMeetingID },
			input: func(id string) map[string]any {
				return map[string]any{"meetingId": id, "name": "Test\u200BMeeting"}
			},
			wantErrorContains: "zero-width",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meetingID := tt.setup()

			query := `
				mutation UpdateMeeting($input: UpdateMeetingInput!) {
					updateMeeting(input: $input) {
						meeting {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(meetingID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeeting_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("delete existing meeting", func(t *testing.T) {
		meetingID := factory.NewMeeting(owner).WithName("Meeting to Delete").Create()

		query := `
			mutation DeleteMeeting($input: DeleteMeetingInput!) {
				deleteMeeting(input: $input) {
					deletedMeetingId
				}
			}
		`

		var result struct {
			DeleteMeeting struct {
				DeletedMeetingID string `json:"deletedMeetingId"`
			} `json:"deleteMeeting"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"meetingId": meetingID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, meetingID, result.DeleteMeeting.DeletedMeetingID)
	})
}

func TestMeeting_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		meetingID         string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			meetingID:         "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteMeeting($input: DeleteMeetingInput!) {
					deleteMeeting(input: $input) {
						deletedMeetingId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"meetingId": tt.meetingID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeeting_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	meetingNames := []string{"Meeting A", "Meeting B", "Meeting C"}
	for _, name := range meetingNames {
		factory.NewMeeting(owner).WithName(name).Create()
	}

	query := `
		query GetMeetings($id: ID!) {
			node(id: $id) {
				... on Organization {
					meetings(first: 10) {
						edges {
							node {
								id
								name
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Meetings struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"meetings"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Meetings.TotalCount, 3)
}

func TestMeeting_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Meeting {
						id
						name
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo",
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestMeeting_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateMeeting($input: CreateMeetingInput!) {
				createMeeting(input: $input) {
					meetingEdge {
						node {
							id
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		var result struct {
			CreateMeeting struct {
				MeetingEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"meetingEdge"`
			} `json:"createMeeting"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Timestamp Test Meeting",
				"date":           time.Now().Format(time.RFC3339Nano),
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateMeeting.MeetingEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		meetingID := factory.NewMeeting(owner).WithName("Timestamp Update Test").Create()

		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on Meeting {
						createdAt
						updatedAt
					}
				}
			}
		`

		var getResult struct {
			Node struct {
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"node"`
		}

		err := owner.Execute(getQuery, map[string]any{"id": meetingID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		updateQuery := `
			mutation UpdateMeeting($input: UpdateMeetingInput!) {
				updateMeeting(input: $input) {
					meeting {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateMeeting struct {
				Meeting struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"meeting"`
			} `json:"updateMeeting"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"meetingId": meetingID,
				"name":      "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		meeting := updateResult.UpdateMeeting.Meeting
		testutil.AssertTimestampsOnUpdate(t, meeting.CreatedAt, meeting.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestMeeting_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	meetingID := factory.NewMeeting(owner).WithName("SubResolver Test Meeting").Create()

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Meeting {
						id
						organization {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID           string `json:"id"`
				Organization struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": meetingID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})
}

func TestMeeting_SubResolvers_WithData(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	attendee1ProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	attendee2ProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	var result struct {
		CreateMeeting struct {
			MeetingEdge struct {
				Node struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					Attendees []struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"attendees"`
				} `json:"node"`
			} `json:"meetingEdge"`
		} `json:"createMeeting"`
	}

	query := `
		mutation($input: CreateMeetingInput!) {
			createMeeting(input: $input) {
				meetingEdge {
					node {
						id
						name
						attendees {
							id
							fullName
						}
					}
				}
			}
		}
	`

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Meeting With Attendees",
			"date":           time.Now().Format(time.RFC3339Nano),
			"attendeeIds":    []string{attendee1ProfileID.String(), attendee2ProfileID.String()},
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result.CreateMeeting.MeetingEdge.Node.Attendees))
}

func TestMeeting_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Run("owner can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)

			_, err := owner.Do(`
				mutation CreateMeeting($input: CreateMeetingInput!) {
					createMeeting(input: $input) {
						meetingEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"name":           "RBAC Test Meeting",
					"date":           time.Now().Format(time.RFC3339Nano),
				},
			})
			require.NoError(t, err, "owner should be able to create meeting")
		})

		t.Run("admin can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

			_, err := admin.Do(`
				mutation CreateMeeting($input: CreateMeetingInput!) {
					createMeeting(input: $input) {
						meetingEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": admin.GetOrganizationID().String(),
					"name":           "RBAC Test Meeting",
					"date":           time.Now().Format(time.RFC3339Nano),
				},
			})
			require.NoError(t, err, "admin should be able to create meeting")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

			_, err := viewer.Do(`
				mutation CreateMeeting($input: CreateMeetingInput!) {
					createMeeting(input: $input) {
						meetingEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": viewer.GetOrganizationID().String(),
					"name":           "RBAC Test Meeting",
					"date":           time.Now().Format(time.RFC3339Nano),
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create meeting")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("owner can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Update Test").Create()

			_, err := owner.Do(`
				mutation UpdateMeeting($input: UpdateMeetingInput!) {
					updateMeeting(input: $input) {
						meeting { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"meetingId": meetingID,
					"name":      "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update meeting")
		})

		t.Run("admin can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Update Test").Create()

			_, err := admin.Do(`
				mutation UpdateMeeting($input: UpdateMeetingInput!) {
					updateMeeting(input: $input) {
						meeting { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"meetingId": meetingID,
					"name":      "Updated by Admin",
				},
			})
			require.NoError(t, err, "admin should be able to update meeting")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Update Test").Create()

			_, err := viewer.Do(`
				mutation UpdateMeeting($input: UpdateMeetingInput!) {
					updateMeeting(input: $input) {
						meeting { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"meetingId": meetingID,
					"name":      "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update meeting")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Delete Test").Create()

			_, err := owner.Do(`
				mutation DeleteMeeting($input: DeleteMeetingInput!) {
					deleteMeeting(input: $input) {
						deletedMeetingId
					}
				}
			`, map[string]any{
				"input": map[string]any{"meetingId": meetingID},
			})
			require.NoError(t, err, "owner should be able to delete meeting")
		})

		t.Run("admin can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Delete Test").Create()

			_, err := admin.Do(`
				mutation DeleteMeeting($input: DeleteMeetingInput!) {
					deleteMeeting(input: $input) {
						deletedMeetingId
					}
				}
			`, map[string]any{
				"input": map[string]any{"meetingId": meetingID},
			})
			require.NoError(t, err, "admin should be able to delete meeting")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Delete Test").Create()

			_, err := viewer.Do(`
				mutation DeleteMeeting($input: DeleteMeetingInput!) {
					deleteMeeting(input: $input) {
						deletedMeetingId
					}
				}
			`, map[string]any{
				"input": map[string]any{"meetingId": meetingID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete meeting")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Run("owner can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := owner.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Meeting { id name }
					}
				}
			`, map[string]any{"id": meetingID}, &result)
			require.NoError(t, err, "owner should be able to read meeting")
			require.NotNil(t, result.Node, "owner should receive meeting data")
		})

		t.Run("admin can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := admin.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Meeting { id name }
					}
				}
			`, map[string]any{"id": meetingID}, &result)
			require.NoError(t, err, "admin should be able to read meeting")
			require.NotNil(t, result.Node, "admin should receive meeting data")
		})

		t.Run("viewer can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			meetingID := factory.NewMeeting(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := viewer.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Meeting { id name }
					}
				}
			`, map[string]any{"id": meetingID}, &result)
			require.NoError(t, err, "viewer should be able to read meeting")
			require.NotNil(t, result.Node, "viewer should receive meeting data")
		})
	})
}

func TestMeeting_MaxLength_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	longName := strings.Repeat("a", 1001)

	t.Run("create", func(t *testing.T) {
		query := `
			mutation CreateMeeting($input: CreateMeetingInput!) {
				createMeeting(input: $input) {
					meetingEdge {
						node { id }
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           longName,
				"date":           time.Now().Format(time.RFC3339Nano),
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("update", func(t *testing.T) {
		meetingID := factory.NewMeeting(owner).WithName("Max Length Test").Create()

		query := `
			mutation UpdateMeeting($input: UpdateMeetingInput!) {
				updateMeeting(input: $input) {
					meeting { id }
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"meetingId": meetingID,
				"name":      longName,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})
}

func TestMeeting_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	for i := 0; i < 5; i++ {
		factory.NewMeeting(owner).
			WithName(fmt.Sprintf("Pagination Meeting %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						meetings(first: 2) {
							edges {
								node { id name }
								cursor
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
								startCursor
								endCursor
							}
							totalCount
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Meetings struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"meetings"`
			} `json:"node"`
		}

		err := owner.Execute(
			query,
			map[string]any{
				"id": owner.GetOrganizationID().String(),
			},
			&result,
		)
		require.NoError(t, err)

		testutil.AssertFirstPage(t, len(result.Node.Meetings.Edges), result.Node.Meetings.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.Meetings.TotalCount, 5)

		testutil.AssertHasMorePages(t, result.Node.Meetings.PageInfo)
		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						meetings(first: 2, after: $after) {
							edges {
								node { id name }
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var resultAfter struct {
			Node struct {
				Meetings struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"meetings"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.Meetings.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.Meetings.Edges), resultAfter.Node.Meetings.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						meetings(last: 2) {
							edges {
								node { id name }
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Meetings struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"meetings"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.Meetings.Edges), result.Node.Meetings.PageInfo, 2, true)
	})
}

func TestMeeting_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	meetingID := factory.NewMeeting(org1Owner).WithName("Org1 Meeting").Create()

	t.Run("cannot read meeting from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Meeting {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": meetingID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "meeting")
	})

	t.Run("cannot update meeting from another organization", func(t *testing.T) {
		query := `
			mutation UpdateMeeting($input: UpdateMeetingInput!) {
				updateMeeting(input: $input) {
					meeting { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"meetingId": meetingID,
				"name":      "Hijacked Meeting",
			},
		})
		require.Error(t, err, "Should not be able to update meeting from another org")
	})

	t.Run("cannot delete meeting from another organization", func(t *testing.T) {
		query := `
			mutation DeleteMeeting($input: DeleteMeetingInput!) {
				deleteMeeting(input: $input) {
					deletedMeetingId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"meetingId": meetingID,
			},
		})
		require.Error(t, err, "Should not be able to delete meeting from another org")
	})

	t.Run("cannot list meetings from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						meetings(first: 100) {
							edges {
								node {
									id
									name
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Meetings struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"meetings"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)

		if err == nil {
			for _, edge := range result.Node.Meetings.Edges {
				assert.NotEqual(t, meetingID, edge.Node.ID, "Should not see meeting from another org")
			}
		}
	})
}

func TestMeeting_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	factory.NewMeeting(owner).WithName("AAA Order Test").Create()
	factory.NewMeeting(owner).WithName("ZZZ Order Test").Create()

	t.Run("order by created_at descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: MeetingOrder) {
				node(id: $id) {
					... on Organization {
						meetings(first: 100, orderBy: $orderBy) {
							edges {
								node {
									id
									createdAt
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Meetings struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"meetings"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
			"orderBy": map[string]any{
				"field":     "CREATED_AT",
				"direction": "DESC",
			},
		}, &result)
		require.NoError(t, err)

		times := make([]time.Time, len(result.Node.Meetings.Edges))
		for i, edge := range result.Node.Meetings.Edges {
			times[i] = edge.Node.CreatedAt
		}
		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}
