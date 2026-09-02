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

// Package factory provides Rails-like test data factories using gofakeit.
package factory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
)

func SafeName(prefix string) string {
	return fmt.Sprintf("%s %s", prefix, gofakeit.LetterN(8))
}

func SafeEmail() string {
	return fmt.Sprintf("%s@example.com", strings.ToLower(gofakeit.LetterN(12)))
}

type Attrs map[string]any

func (a Attrs) get(key string, defaultVal any) any {
	if a == nil {
		return defaultVal
	}

	if v, ok := a[key]; ok {
		return v
	}

	return defaultVal
}

func (a Attrs) getString(key string, defaultVal string) string {
	if v, ok := a.get(key, defaultVal).(string); ok {
		return v
	}

	return defaultVal
}

func (a Attrs) getStringPtr(key string) *string {
	if a == nil {
		return nil
	}

	if v, ok := a[key]; ok {
		if s, ok := v.(string); ok {
			return &s
		}
	}

	return nil
}

func (a Attrs) getInt(key string, defaultVal int) int {
	if a == nil {
		return defaultVal
	}

	if v, ok := a[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}

	return defaultVal
}

func (a Attrs) getIntPtr(key string) *int {
	if a == nil {
		return nil
	}

	v, ok := a[key]
	if !ok {
		return nil
	}

	switch val := v.(type) {
	case int:
		return &val
	case int64:
		i := int(val)
		return &i
	case float64:
		i := int(val)
		return &i
	default:
		return nil
	}
}

func (a Attrs) getBool(key string, defaultVal bool) bool {
	if a == nil {
		return defaultVal
	}

	if v, ok := a[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}

	return defaultVal
}

func CreateUser(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateUserInput!) {
			createUser(input: $input) {
				profileEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":           c.GetOrganizationID(),
		"emailAddress":             a.getString("emailAddress", SafeEmail()),
		"fullName":                 a.getString("fullName", SafeName("User")),
		"role":                     a.getString("role", "EMPLOYEE"),
		"kind":                     "EMPLOYEE",
		"additionalEmailAddresses": []string{},
	}
	if position := a.getStringPtr("position"); position != nil {
		input["position"] = *position
	}

	var result struct {
		CreateUser struct {
			ProfileEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"profileEdge"`
		} `json:"createUser"`
	}

	err := c.ExecuteConnect(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createUser mutation failed")

	return result.CreateUser.ProfileEdge.Node.ID
}

func InviteUser(c *testutil.Client, profileID string) string {
	c.T.Helper()

	const query = `
		mutation($input: InviteUserInput!) {
			inviteUser(input: $input) {
				invitationEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		InviteUser struct {
			InvitationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"invitationEdge"`
		} `json:"inviteUser"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId": c.GetOrganizationID().String(),
			"profileId":      profileID,
		},
	}, &result)
	require.NoError(c.T, err, "inviteUser mutation failed")

	return result.InviteUser.InvitationEdge.Node.ID
}

func CreateThirdParty(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("ThirdParty")),
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	if url := a.getStringPtr("websiteUrl"); url != nil {
		input["websiteUrl"] = *url
	}

	if cat := a.getStringPtr("category"); cat != nil {
		input["category"] = *cat
	}

	if v, ok := a["administratorIds"]; ok {
		input["administratorIds"] = v
	}

	var result struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createThirdParty mutation failed")

	return result.CreateThirdParty.ThirdPartyEdge.Node.ID
}

// InjectCrossTenantThirdPartyAdministrator bypasses the application and writes a
// third_party_administrators row for a foreign profile, for read-gap security tests.
func InjectCrossTenantThirdPartyAdministrator(t *testing.T, thirdPartyID, foreignProfileID string) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `
INSERT INTO third_party_administrators (
	third_party_id,
	administrator_profile_id,
	tenant_id,
	organization_id,
	created_at,
	updated_at
)
SELECT
	id,
	$1,
	tenant_id,
	organization_id,
	NOW(),
	NOW()
FROM third_parties
WHERE id = $2
ON CONFLICT DO NOTHING
`, foreignProfileID, thirdPartyID)

		return err
	})
	require.NoError(t, err, "test setup: cannot inject cross-tenant third party administrator")
}

func RequireFileSoftDeleted(t *testing.T, fileID string) {
	t.Helper()

	var deletedAt *time.Time

	err := test.PGClient(t).WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				`
SELECT deleted_at
FROM files
WHERE id = $1
`,
				fileID,
			).Scan(&deletedAt)
		},
	)
	require.NoError(t, err, "cannot load file deleted_at")
	require.NotNil(t, deletedAt, "file %s must be soft-deleted", fileID)
}

func CreateFramework(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("Framework")),
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	var result struct {
		CreateFramework struct {
			FrameworkEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"frameworkEdge"`
		} `json:"createFramework"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createFramework mutation failed")

	return result.CreateFramework.FrameworkEdge.Node.ID
}

func CreateControl(c *testutil.Client, frameworkID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"frameworkId":   frameworkID,
		"name":          a.getString("name", SafeName("Control")),
		"description":   a.getString("description", "Test control description"),
		"sectionTitle":  a.getString("sectionTitle", fmt.Sprintf("Section %s", gofakeit.LetterN(3))),
		"bestPractice":  a.getBool("bestPractice", true),
		"maturityLevel": a.getString("maturityLevel", "INITIAL"),
	}

	if justification := a.getStringPtr("notImplementedJustification"); justification != nil {
		input["notImplementedJustification"] = *justification
	}

	var result struct {
		CreateControl struct {
			ControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControl"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createControl mutation failed")

	return result.CreateControl.ControlEdge.Node.ID
}

func CreateMeasure(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateMeasureInput!) {
			createMeasure(input: $input) {
				measureEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("Measure")),
		"category":       a.getString("category", "POLICY"),
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	var result struct {
		CreateMeasure struct {
			MeasureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"measureEdge"`
		} `json:"createMeasure"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createMeasure mutation failed")

	return result.CreateMeasure.MeasureEdge.Node.ID
}

func CreateTask(c *testutil.Client, measureID *string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("Task")),
		"priority":       a.getString("priority", "MEDIUM"),
	}
	if measureID != nil {
		input["measureId"] = *measureID
	}

	if content := a.getStringPtr("content"); content != nil {
		input["content"] = ProseMirrorPlainText(*content)
	}

	var result struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createTask mutation failed")

	return result.CreateTask.TaskEdge.Node.ID
}

func CreateRisk(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskInput!) {
			createRisk(input: $input) {
				riskEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":     c.GetOrganizationID().String(),
		"name":               a.getString("name", SafeName("Risk")),
		"category":           a.getString("category", "SECURITY"),
		"treatment":          a.getString("treatment", "MITIGATED"),
		"inherentLikelihood": a.getInt("inherentLikelihood", 2),
		"inherentImpact":     a.getInt("inherentImpact", 2),
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	var result struct {
		CreateRisk struct {
			RiskEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskEdge"`
		} `json:"createRisk"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRisk mutation failed")

	return result.CreateRisk.RiskEdge.Node.ID
}

type ThirdPartyBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewThirdParty(c *testutil.Client) *ThirdPartyBuilder {
	return &ThirdPartyBuilder{client: c, attrs: Attrs{}}
}

func (b *ThirdPartyBuilder) WithName(name string) *ThirdPartyBuilder {
	b.attrs["name"] = name
	return b
}

func (b *ThirdPartyBuilder) WithDescription(desc string) *ThirdPartyBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *ThirdPartyBuilder) WithWebsiteUrl(url string) *ThirdPartyBuilder {
	b.attrs["websiteUrl"] = url
	return b
}

func (b *ThirdPartyBuilder) WithCategory(category string) *ThirdPartyBuilder {
	b.attrs["category"] = category
	return b
}

func (b *ThirdPartyBuilder) Create() string {
	return CreateThirdParty(b.client, b.attrs)
}

type FrameworkBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewFramework(c *testutil.Client) *FrameworkBuilder {
	return &FrameworkBuilder{client: c, attrs: Attrs{}}
}

func (b *FrameworkBuilder) WithName(name string) *FrameworkBuilder {
	b.attrs["name"] = name
	return b
}

func (b *FrameworkBuilder) WithDescription(desc string) *FrameworkBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *FrameworkBuilder) Create() string {
	return CreateFramework(b.client, b.attrs)
}

type ControlBuilder struct {
	client      *testutil.Client
	frameworkID string
	attrs       Attrs
}

func NewControl(c *testutil.Client, frameworkID string) *ControlBuilder {
	return &ControlBuilder{client: c, frameworkID: frameworkID, attrs: Attrs{}}
}

func (b *ControlBuilder) WithName(name string) *ControlBuilder {
	b.attrs["name"] = name
	return b
}

func (b *ControlBuilder) WithDescription(desc string) *ControlBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *ControlBuilder) WithSectionTitle(title string) *ControlBuilder {
	b.attrs["sectionTitle"] = title
	return b
}

func (b *ControlBuilder) WithStatus(status string) *ControlBuilder {
	b.attrs["status"] = status
	return b
}

func (b *ControlBuilder) WithBestPractice(bestPractice bool) *ControlBuilder {
	b.attrs["bestPractice"] = bestPractice
	return b
}

func (b *ControlBuilder) WithMaturityLevel(maturityLevel string) *ControlBuilder {
	b.attrs["maturityLevel"] = maturityLevel
	return b
}

func (b *ControlBuilder) WithNotImplementedJustification(justification string) *ControlBuilder {
	b.attrs["notImplementedJustification"] = justification
	return b
}

func (b *ControlBuilder) Create() string {
	return CreateControl(b.client, b.frameworkID, b.attrs)
}

type MeasureBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewMeasure(c *testutil.Client) *MeasureBuilder {
	return &MeasureBuilder{client: c, attrs: Attrs{}}
}

func (b *MeasureBuilder) WithName(name string) *MeasureBuilder {
	b.attrs["name"] = name
	return b
}

func (b *MeasureBuilder) WithDescription(desc string) *MeasureBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *MeasureBuilder) WithCategory(category string) *MeasureBuilder {
	b.attrs["category"] = category
	return b
}

func (b *MeasureBuilder) Create() string {
	return CreateMeasure(b.client, b.attrs)
}

type TaskBuilder struct {
	client    *testutil.Client
	measureID *string
	attrs     Attrs
}

func NewTask(c *testutil.Client, measureID string) *TaskBuilder {
	return &TaskBuilder{client: c, measureID: &measureID, attrs: Attrs{}}
}

func NewTaskWithoutMeasure(c *testutil.Client) *TaskBuilder {
	return &TaskBuilder{client: c, measureID: nil, attrs: Attrs{}}
}

func (b *TaskBuilder) WithName(name string) *TaskBuilder {
	b.attrs["name"] = name
	return b
}

func (b *TaskBuilder) WithContent(content string) *TaskBuilder {
	b.attrs["content"] = content
	return b
}

func (b *TaskBuilder) Create() string {
	return CreateTask(b.client, b.measureID, b.attrs)
}

func CreateTaskComment(c *testutil.Client, taskID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateTaskCommentInput!) {
			createTaskComment(input: $input) {
				taskCommentEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"taskId":  taskID,
		"content": ProseMirrorPlainText(a.getString("content", SafeName("Comment"))),
	}

	if ownerID, ok := a["ownerId"]; ok {
		input["ownerId"] = ownerID
	}

	var result struct {
		CreateTaskComment struct {
			TaskCommentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"taskCommentEdge"`
		} `json:"createTaskComment"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createTaskComment mutation failed")

	return result.CreateTaskComment.TaskCommentEdge.Node.ID
}

type TaskCommentBuilder struct {
	client *testutil.Client
	taskID string
	attrs  Attrs
}

func NewTaskComment(c *testutil.Client, taskID string) *TaskCommentBuilder {
	return &TaskCommentBuilder{client: c, taskID: taskID, attrs: Attrs{}}
}

func (b *TaskCommentBuilder) WithContent(content string) *TaskCommentBuilder {
	b.attrs["content"] = content
	return b
}

func (b *TaskCommentBuilder) WithOwnerID(ownerID string) *TaskCommentBuilder {
	b.attrs["ownerId"] = ownerID
	return b
}

func (b *TaskCommentBuilder) Create() string {
	return CreateTaskComment(b.client, b.taskID, b.attrs)
}

type RiskBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewRisk(c *testutil.Client) *RiskBuilder {
	return &RiskBuilder{client: c, attrs: Attrs{}}
}

func (b *RiskBuilder) WithName(name string) *RiskBuilder {
	b.attrs["name"] = name
	return b
}

func (b *RiskBuilder) WithDescription(desc string) *RiskBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *RiskBuilder) WithCategory(category string) *RiskBuilder {
	b.attrs["category"] = category
	return b
}

func (b *RiskBuilder) WithTreatment(treatment string) *RiskBuilder {
	b.attrs["treatment"] = treatment
	return b
}

func (b *RiskBuilder) WithLikelihood(likelihood int) *RiskBuilder {
	b.attrs["inherentLikelihood"] = likelihood
	return b
}

func (b *RiskBuilder) WithImpact(impact int) *RiskBuilder {
	b.attrs["inherentImpact"] = impact
	return b
}

func (b *RiskBuilder) Create() string {
	return CreateRisk(b.client, b.attrs)
}

func CreateAudit(c *testutil.Client, frameworkID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateAuditInput!) {
			createAudit(input: $input) {
				auditEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"frameworkId":    frameworkID,
		"name":           a.getString("name", SafeName("Audit")),
	}
	if state := a.getStringPtr("state"); state != nil {
		input["state"] = *state
	}

	if firm := a.getStringPtr("firm"); firm != nil {
		input["firm"] = *firm
	}

	var result struct {
		CreateAudit struct {
			AuditEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"auditEdge"`
		} `json:"createAudit"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createAudit mutation failed")

	return result.CreateAudit.AuditEdge.Node.ID
}

type AuditBuilder struct {
	client      *testutil.Client
	frameworkID string
	attrs       Attrs
}

func NewAudit(c *testutil.Client, frameworkID string) *AuditBuilder {
	return &AuditBuilder{client: c, frameworkID: frameworkID, attrs: Attrs{}}
}

func (b *AuditBuilder) WithName(name string) *AuditBuilder {
	b.attrs["name"] = name
	return b
}

func (b *AuditBuilder) WithState(state string) *AuditBuilder {
	b.attrs["state"] = state
	return b
}

func (b *AuditBuilder) WithFirm(firm string) *AuditBuilder {
	b.attrs["firm"] = firm
	return b
}

func (b *AuditBuilder) Create() string {
	return CreateAudit(b.client, b.frameworkID, b.attrs)
}

func CreateDatum(c *testutil.Client, ownerID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateDatumInput!) {
			createDatum(input: $input) {
				datumEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":     c.GetOrganizationID().String(),
		"ownerId":            ownerID,
		"name":               a.getString("name", SafeName("Datum")),
		"dataClassification": a.getString("dataClassification", "INTERNAL"),
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	var result struct {
		CreateDatum struct {
			DatumEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"datumEdge"`
		} `json:"createDatum"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createDatum mutation failed")

	return result.CreateDatum.DatumEdge.Node.ID
}

type DatumBuilder struct {
	client  *testutil.Client
	ownerID string
	attrs   Attrs
}

func NewDatum(c *testutil.Client, ownerID string) *DatumBuilder {
	return &DatumBuilder{client: c, ownerID: ownerID, attrs: Attrs{}}
}

func (b *DatumBuilder) WithName(name string) *DatumBuilder {
	b.attrs["name"] = name
	return b
}

func (b *DatumBuilder) WithDescription(desc string) *DatumBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *DatumBuilder) WithDataClassification(classification string) *DatumBuilder {
	b.attrs["dataClassification"] = classification
	return b
}

func (b *DatumBuilder) Create() string {
	return CreateDatum(b.client, b.ownerID, b.attrs)
}

type DocumentBuilder struct {
	client    *testutil.Client
	attrs     Attrs
	versionID string
}

func NewDocument(c *testutil.Client) *DocumentBuilder {
	return &DocumentBuilder{client: c, attrs: Attrs{}}
}

func (b *DocumentBuilder) WithTitle(title string) *DocumentBuilder {
	b.attrs["title"] = title
	return b
}

func (b *DocumentBuilder) WithContent(content string) *DocumentBuilder {
	b.attrs["content"] = content
	return b
}

func (b *DocumentBuilder) WithDocumentType(docType string) *DocumentBuilder {
	b.attrs["documentType"] = docType
	return b
}

func (b *DocumentBuilder) WithClassification(classification string) *DocumentBuilder {
	b.attrs["classification"] = classification
	return b
}

func (b *DocumentBuilder) VersionID() string {
	return b.versionID
}

func (b *DocumentBuilder) Create() string {
	b.client.T.Helper()

	a := b.attrs

	const query = `
		mutation($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node { id }
				}
				documentVersionEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": b.client.GetOrganizationID().String(),
		"title":          a.getString("title", SafeName("Document")),
		"documentType":   a.getString("documentType", "POLICY"),
		"classification": a.getString("classification", "INTERNAL"),
	}

	var result struct {
		CreateDocument struct {
			DocumentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentEdge"`
			DocumentVersionEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentVersionEdge"`
		} `json:"createDocument"`
	}

	err := b.client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(b.client.T, err, "createDocument mutation failed")

	b.versionID = result.CreateDocument.DocumentVersionEdge.Node.ID

	return result.CreateDocument.DocumentEdge.Node.ID
}

func CreateProcessingActivity(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateProcessingActivityInput!) {
			createProcessingActivity(input: $input) {
				processingActivityEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":                       c.GetOrganizationID().String(),
		"name":                                 a.getString("name", SafeName("ProcessingActivity")),
		"specialOrCriminalData":                a.getString("specialOrCriminalData", "NO"),
		"lawfulBasis":                          a.getString("lawfulBasis", "CONSENT"),
		"internationalTransfers":               a.getBool("internationalTransfers", false),
		"dataProtectionImpactAssessmentNeeded": a.getString("dataProtectionImpactAssessmentNeeded", "NOT_NEEDED"),
		"transferImpactAssessmentNeeded":       a.getString("transferImpactAssessmentNeeded", "NOT_NEEDED"),
		"role":                                 a.getString("role", "CONTROLLER"),
	}
	if purpose := a.getStringPtr("purpose"); purpose != nil {
		input["purpose"] = *purpose
	}

	var result struct {
		CreateProcessingActivity struct {
			ProcessingActivityEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"processingActivityEdge"`
		} `json:"createProcessingActivity"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createProcessingActivity mutation failed")

	return result.CreateProcessingActivity.ProcessingActivityEdge.Node.ID
}

type ProcessingActivityBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewProcessingActivity(c *testutil.Client) *ProcessingActivityBuilder {
	return &ProcessingActivityBuilder{client: c, attrs: Attrs{}}
}

func (b *ProcessingActivityBuilder) WithName(name string) *ProcessingActivityBuilder {
	b.attrs["name"] = name
	return b
}

func (b *ProcessingActivityBuilder) WithPurpose(purpose string) *ProcessingActivityBuilder {
	b.attrs["purpose"] = purpose
	return b
}

func (b *ProcessingActivityBuilder) WithLawfulBasis(basis string) *ProcessingActivityBuilder {
	b.attrs["lawfulBasis"] = basis
	return b
}

func (b *ProcessingActivityBuilder) WithInternationalTransfers(transfers bool) *ProcessingActivityBuilder {
	b.attrs["internationalTransfers"] = transfers
	return b
}

func (b *ProcessingActivityBuilder) WithSpecialOrCriminalData(value string) *ProcessingActivityBuilder {
	b.attrs["specialOrCriminalData"] = value
	return b
}

func (b *ProcessingActivityBuilder) Create() string {
	return CreateProcessingActivity(b.client, b.attrs)
}

func CreateAccessReviewSource(c *testutil.Client, organizationID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateAccessReviewSourceInput!) {
			createAccessReviewSource(input: $input) {
				accessReviewSourceEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": organizationID,
		"name":           a.getString("name", SafeName("AccessReviewSource")),
	}
	if csvData := a.getStringPtr("csvData"); csvData != nil {
		input["csvData"] = *csvData
	}

	if connectorID := a.getStringPtr("connectorId"); connectorID != nil {
		input["connectorId"] = *connectorID
	}

	var result struct {
		CreateAccessReviewSource struct {
			AccessReviewSourceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"accessReviewSourceEdge"`
		} `json:"createAccessReviewSource"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createAccessReviewSource mutation failed")

	return result.CreateAccessReviewSource.AccessReviewSourceEdge.Node.ID
}

type AccessReviewSourceBuilder struct {
	client         *testutil.Client
	organizationID string
	attrs          Attrs
}

func NewAccessReviewSource(c *testutil.Client, organizationID string) *AccessReviewSourceBuilder {
	return &AccessReviewSourceBuilder{client: c, organizationID: organizationID, attrs: Attrs{}}
}

func (b *AccessReviewSourceBuilder) WithName(name string) *AccessReviewSourceBuilder {
	b.attrs["name"] = name
	return b
}

func (b *AccessReviewSourceBuilder) WithCsvData(csvData string) *AccessReviewSourceBuilder {
	b.attrs["csvData"] = csvData
	return b
}

func (b *AccessReviewSourceBuilder) Create() string {
	return CreateAccessReviewSource(b.client, b.organizationID, b.attrs)
}

func CreateAccessReviewCampaign(c *testutil.Client, organizationID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateAccessReviewCampaignInput!) {
			createAccessReviewCampaign(input: $input) {
				accessReviewCampaignEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": organizationID,
		"name":           a.getString("name", SafeName("Campaign")),
	}

	if v, ok := a["accessReviewSourceIds"]; ok {
		input["accessReviewSourceIds"] = v
	}

	var result struct {
		CreateAccessReviewCampaign struct {
			AccessReviewCampaignEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"accessReviewCampaignEdge"`
		} `json:"createAccessReviewCampaign"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createAccessReviewCampaign mutation failed")

	return result.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node.ID
}

type AccessReviewCampaignBuilder struct {
	client         *testutil.Client
	organizationID string
	attrs          Attrs
}

func NewAccessReviewCampaign(c *testutil.Client, organizationID string) *AccessReviewCampaignBuilder {
	return &AccessReviewCampaignBuilder{client: c, organizationID: organizationID, attrs: Attrs{}}
}

func (b *AccessReviewCampaignBuilder) WithName(name string) *AccessReviewCampaignBuilder {
	b.attrs["name"] = name
	return b
}

func (b *AccessReviewCampaignBuilder) WithAccessReviewSourceIDs(ids []string) *AccessReviewCampaignBuilder {
	b.attrs["accessReviewSourceIds"] = ids
	return b
}

func (b *AccessReviewCampaignBuilder) Create() string {
	return CreateAccessReviewCampaign(b.client, b.organizationID, b.attrs)
}

type StatementOfApplicabilityBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewStatementOfApplicability(c *testutil.Client) *StatementOfApplicabilityBuilder {
	return &StatementOfApplicabilityBuilder{client: c, attrs: Attrs{}}
}

func (b *StatementOfApplicabilityBuilder) WithName(name string) *StatementOfApplicabilityBuilder {
	b.attrs["name"] = name
	return b
}

func (b *StatementOfApplicabilityBuilder) Create() string {
	b.client.T.Helper()

	a := b.attrs

	const query = `
		mutation($input: CreateStatementOfApplicabilityInput!) {
			createStatementOfApplicability(input: $input) {
				statementOfApplicabilityEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": b.client.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("SOA")),
	}

	var result struct {
		CreateStatementOfApplicability struct {
			StatementOfApplicabilityEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"statementOfApplicabilityEdge"`
		} `json:"createStatementOfApplicability"`
	}

	err := b.client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(b.client.T, err, "createStatementOfApplicability mutation failed")

	return result.CreateStatementOfApplicability.StatementOfApplicabilityEdge.Node.ID
}

func CreateApplicabilityStatement(c *testutil.Client, soaID, controlID string, applicability bool, justification *string) string {
	c.T.Helper()

	const query = `
		mutation($input: CreateApplicabilityStatementInput!) {
			createApplicabilityStatement(input: $input) {
				applicabilityStatementEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"statementOfApplicabilityId": soaID,
		"controlId":                  controlID,
		"applicability":              applicability,
	}

	if justification != nil {
		input["justification"] = *justification
	}

	var result struct {
		CreateApplicabilityStatement struct {
			ApplicabilityStatementEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"applicabilityStatementEdge"`
		} `json:"createApplicabilityStatement"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createApplicabilityStatement mutation failed")

	return result.CreateApplicabilityStatement.ApplicabilityStatementEdge.Node.ID
}

type OAuth2ClientResult struct {
	ClientID     string
	ClientSecret string
}

func CreateOAuth2Client(c *testutil.Client, attrs Attrs) OAuth2ClientResult {
	input := map[string]any{
		"organization_id": c.GetOrganizationID().String(),
		"client_name":     SafeName("OAuth2 Client"),
		"visibility":      "private",
		"redirect_uris":   []string{"http://localhost:9999/callback"},
		"grant_types": []string{
			"authorization_code",
			"refresh_token",
		},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_basic",
		"scopes":                     "openid email profile offline_access",
	}

	maps.Copy(input, attrs)

	resp, raw, err := testutil.OAuth2RegisterClient(c, input)
	require.NoError(c.T, err, "OAuth2 client registration failed")
	require.NotNil(c.T, resp, "OAuth2 client registration returned nil (status=%d body=%s)", raw.StatusCode, string(raw.Body))

	return OAuth2ClientResult{
		ClientID:     resp.ClientID,
		ClientSecret: resp.ClientSecret,
	}
}

func CreateOAuth2ClientWithAPIScopes(c *testutil.Client, scopes string, attrs Attrs) OAuth2ClientResult {
	input := map[string]any{
		"organization_id": c.GetOrganizationID().String(),
		"client_name":     SafeName("OAuth2 API Client"),
		"visibility":      "private",
		"redirect_uris":   []string{"http://localhost:9999/callback"},
		"grant_types": []string{
			"authorization_code",
			"refresh_token",
		},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_basic",
		"scopes":                     scopes,
	}

	maps.Copy(input, attrs)

	resp, raw, err := testutil.OAuth2RegisterClient(c, input)
	require.NoError(c.T, err, "OAuth2 API client registration failed")
	require.NotNil(c.T, resp, "OAuth2 API client registration returned nil (status=%d body=%s)", raw.StatusCode, string(raw.Body))

	return OAuth2ClientResult{
		ClientID:     resp.ClientID,
		ClientSecret: resp.ClientSecret,
	}
}

func CreatePublicOAuth2Client(c *testutil.Client, attrs Attrs) OAuth2ClientResult {
	input := map[string]any{
		"organization_id": c.GetOrganizationID().String(),
		"client_name":     SafeName("Public OAuth2 Client"),
		"visibility":      "private",
		"redirect_uris":   []string{"http://localhost:9999/callback"},
		"grant_types": []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:device_code",
		},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
		"scopes":                     "openid email profile offline_access",
	}

	maps.Copy(input, attrs)

	resp, raw, err := testutil.OAuth2RegisterClient(c, input)
	require.NoError(c.T, err, "public OAuth2 client registration failed")
	require.NotNil(c.T, resp, "public OAuth2 client registration returned nil (status=%d body=%s)", raw.StatusCode, string(raw.Body))

	return OAuth2ClientResult{
		ClientID:     resp.ClientID,
		ClientSecret: resp.ClientSecret,
	}
}

func SafeOrigin() string {
	return fmt.Sprintf("https://%s.example.com", strings.ToLower(gofakeit.LetterN(10)))
}

func CreateCookieBanner(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateCookieBannerInput!) {
			createCookieBanner(input: $input) {
				cookieBannerEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":    c.GetOrganizationID().String(),
		"name":              a.getString("name", SafeName("CookieBanner")),
		"origin":            a.getString("origin", SafeOrigin()),
		"cookiePolicyUrl":   a.getString("cookiePolicyUrl", "https://example.com/cookies"),
		"consentExpiryDays": a.getInt("consentExpiryDays", 365),
	}
	if ppURL := a.getStringPtr("privacyPolicyUrl"); ppURL != nil {
		input["privacyPolicyUrl"] = *ppURL
	}

	var result struct {
		CreateCookieBanner struct {
			CookieBannerEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"cookieBannerEdge"`
		} `json:"createCookieBanner"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createCookieBanner mutation failed")

	return result.CreateCookieBanner.CookieBannerEdge.Node.ID
}

func CreateCompliancePortal(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateCompliancePortalInput!) {
			createCompliancePortal(input: $input) {
				compliancePortalEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"entityName":     a.getString("entityName", SafeName("Compliance Page")),
	}

	var result struct {
		CreateCompliancePortal struct {
			CompliancePortalEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"compliancePortalEdge"`
		} `json:"createCompliancePortal"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createCompliancePortal mutation failed")

	return result.CreateCompliancePortal.CompliancePortalEdge.Node.ID
}

type CookieBannerBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewCookieBanner(c *testutil.Client) *CookieBannerBuilder {
	return &CookieBannerBuilder{client: c, attrs: Attrs{}}
}

func (b *CookieBannerBuilder) WithName(name string) *CookieBannerBuilder {
	b.attrs["name"] = name
	return b
}

func (b *CookieBannerBuilder) WithOrigin(origin string) *CookieBannerBuilder {
	b.attrs["origin"] = origin
	return b
}

func (b *CookieBannerBuilder) WithCookiePolicyUrl(url string) *CookieBannerBuilder {
	b.attrs["cookiePolicyUrl"] = url
	return b
}

func (b *CookieBannerBuilder) WithPrivacyPolicyUrl(url string) *CookieBannerBuilder {
	b.attrs["privacyPolicyUrl"] = url
	return b
}

func (b *CookieBannerBuilder) WithConsentExpiryDays(days int) *CookieBannerBuilder {
	b.attrs["consentExpiryDays"] = days
	return b
}

func (b *CookieBannerBuilder) Create() string {
	return CreateCookieBanner(b.client, b.attrs)
}

func CreateCookieCategory(c *testutil.Client, bannerID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateCookieCategoryInput!) {
			createCookieCategory(input: $input) {
				cookieCategoryEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"cookieBannerId": bannerID,
		"name":           a.getString("name", SafeName("Category")),
		"slug":           a.getString("slug", strings.ToLower(gofakeit.LetterN(8))),
		"description":    a.getString("description", "Test cookie category"),
		"rank":           a.getInt("rank", 10),
	}

	var result struct {
		CreateCookieCategory struct {
			CookieCategoryEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"cookieCategoryEdge"`
		} `json:"createCookieCategory"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createCookieCategory mutation failed")

	return result.CreateCookieCategory.CookieCategoryEdge.Node.ID
}

func CreateTrackerPattern(c *testutil.Client, categoryID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateTrackerPatternInput!) {
			createTrackerPattern(input: $input) {
				trackerPatternEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"cookieCategoryId": categoryID,
		"trackerType":      a.getString("trackerType", "COOKIE"),
		"pattern":          a.getString("pattern", gofakeit.LetterN(8)+"_cookie"),
		"matchType":        a.getString("matchType", "EXACT"),
		"displayName":      a.getString("displayName", SafeName("Pattern")),
		"description":      a.getString("description", "Test tracker pattern"),
	}
	if _, ok := a["maxAgeSeconds"]; ok {
		input["maxAgeSeconds"] = a.getInt("maxAgeSeconds", 0)
	}

	var result struct {
		CreateTrackerPattern struct {
			TrackerPatternEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"trackerPatternEdge"`
		} `json:"createTrackerPattern"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createTrackerPattern mutation failed")

	return result.CreateTrackerPattern.TrackerPatternEdge.Node.ID
}

func CreateTrackerResource(c *testutil.Client, categoryID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateTrackerResourceInput!) {
			createTrackerResource(input: $input) {
				trackerResourceEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"cookieCategoryId": categoryID,
		"type":             a.getString("type", "SCRIPT"),
		"origin":           a.getString("origin", SafeOrigin()),
		"path":             a.getString("path", "/"+gofakeit.LetterN(8)+".js"),
		"displayName":      a.getString("displayName", SafeName("Resource")),
		"description":      a.getString("description", "Test tracker resource"),
	}

	var result struct {
		CreateTrackerResource struct {
			TrackerResourceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"trackerResourceEdge"`
		} `json:"createTrackerResource"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createTrackerResource mutation failed")

	return result.CreateTrackerResource.TrackerResourceEdge.Node.ID
}

func ReportDetectedResources(c *testutil.Client, bannerID string, count int) {
	c.T.Helper()

	resources := make([]map[string]string, 0, count)
	for range count {
		resources = append(resources, map[string]string{
			"url":           fmt.Sprintf("https://%s.example.com/%s.js", strings.ToLower(gofakeit.LetterN(8)), gofakeit.LetterN(6)),
			"resource_type": "script",
		})
	}

	body, err := json.Marshal(map[string]any{"resources": resources})
	require.NoError(c.T, err, "cannot marshal report body")

	url := fmt.Sprintf("%s/api/cookie-banner/v1/%s/report", c.BaseURL(), bannerID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	require.NoError(c.T, err, "report detected resources request failed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(c.T, http.StatusNoContent, resp.StatusCode, "report detected resources unexpected status")
}

func CreateRiskAnalysis(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisInput!) {
			createRiskAnalysis(input: $input) {
				riskAnalysisEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("Risk Analysis")),
		"matrixSize": map[string]any{
			"rows": a.getInt("matrixRows", 5),
			"cols": a.getInt("matrixCols", 5),
		},
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	var result struct {
		CreateRiskAnalysis struct {
			RiskAnalysisEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisEdge"`
		} `json:"createRiskAnalysis"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysis mutation failed")

	return result.CreateRiskAnalysis.RiskAnalysisEdge.Node.ID
}

func CreateTreatmentPlan(c *testutil.Client, riskID, riskAnalysisID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateTreatmentPlanInput!) {
			createTreatmentPlan(input: $input) {
				treatmentPlanEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskId":             riskID,
		"riskAnalysisId":     riskAnalysisID,
		"treatment":          a.getString("treatment", "MITIGATED"),
		"ownerId":            a.getString("ownerId", c.GetProfileID().String()),
		"inherentLikelihood": a.getInt("inherentLikelihood", 2),
		"inherentImpact":     a.getInt("inherentImpact", 3),
	}

	if residualLikelihood := a.getIntPtr("residualLikelihood"); residualLikelihood != nil {
		input["residualLikelihood"] = *residualLikelihood
	}

	if residualImpact := a.getIntPtr("residualImpact"); residualImpact != nil {
		input["residualImpact"] = *residualImpact
	}

	var result struct {
		CreateTreatmentPlan struct {
			TreatmentPlanEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"treatmentPlanEdge"`
		} `json:"createTreatmentPlan"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createTreatmentPlan mutation failed")

	return result.CreateTreatmentPlan.TreatmentPlanEdge.Node.ID
}

func LinkTreatmentPlanMeasure(c *testutil.Client, treatmentPlanID, measureID string) {
	c.T.Helper()

	const query = `
		mutation($input: CreateTreatmentPlanMeasureMappingInput!) {
			createTreatmentPlanMeasureMapping(input: $input) {
				measureEdge { node { id } }
			}
		}
	`

	var result struct {
		CreateTreatmentPlanMeasureMapping struct {
			MeasureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"measureEdge"`
		} `json:"createTreatmentPlanMeasureMapping"`
	}

	err := c.Execute(query, map[string]any{
		"input": map[string]any{
			"treatmentPlanId": treatmentPlanID,
			"measureId":       measureID,
		},
	}, &result)
	require.NoError(c.T, err, "createTreatmentPlanMeasureMapping mutation failed")
}

func UnlinkTreatmentPlanMeasure(c *testutil.Client, treatmentPlanID, measureID string) {
	c.T.Helper()

	const query = `
		mutation($input: DeleteTreatmentPlanMeasureMappingInput!) {
			deleteTreatmentPlanMeasureMapping(input: $input) {
				deletedMeasureId
			}
		}
	`

	var result struct {
		DeleteTreatmentPlanMeasureMapping struct {
			DeletedMeasureID string `json:"deletedMeasureId"`
		} `json:"deleteTreatmentPlanMeasureMapping"`
	}

	err := c.Execute(query, map[string]any{
		"input": map[string]any{
			"treatmentPlanId": treatmentPlanID,
			"measureId":       measureID,
		},
	}, &result)
	require.NoError(c.T, err, "deleteTreatmentPlanMeasureMapping mutation failed")
}

func CreateRiskAnalysisDiagram(c *testutil.Client, riskAnalysisID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisDiagramInput!) {
			createRiskAnalysisDiagram(input: $input) {
				riskAnalysisDiagramEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskAnalysisId": riskAnalysisID,
		"name":           a.getString("name", SafeName("Scope")),
	}

	var result struct {
		CreateRiskAnalysisDiagram struct {
			RiskAnalysisDiagramEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisDiagramEdge"`
		} `json:"createRiskAnalysisDiagram"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysisDiagram mutation failed")

	return result.CreateRiskAnalysisDiagram.RiskAnalysisDiagramEdge.Node.ID
}

func CreateRiskAnalysisNode(c *testutil.Client, scopeID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisNodeInput!) {
			createRiskAnalysisNode(input: $input) {
				riskAnalysisNodeEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskAnalysisDiagramId": scopeID,
		"nodeType":              a.getString("nodeType", "ASSET"),
		"name":                  a.getString("name", SafeName("Node")),
	}

	if boundaryID := a.getString("boundaryId", ""); boundaryID != "" {
		input["boundaryId"] = boundaryID
	}

	var result struct {
		CreateRiskAnalysisNode struct {
			RiskAnalysisNodeEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisNodeEdge"`
		} `json:"createRiskAnalysisNode"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysisNode mutation failed")

	return result.CreateRiskAnalysisNode.RiskAnalysisNodeEdge.Node.ID
}

func CreateRiskAnalysisBoundary(c *testutil.Client, scopeID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisBoundaryInput!) {
			createRiskAnalysisBoundary(input: $input) {
				riskAnalysisBoundaryEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskAnalysisDiagramId": scopeID,
		"name":                  a.getString("name", SafeName("Boundary")),
	}

	if parentID := a.getString("parentBoundaryId", ""); parentID != "" {
		input["parentBoundaryId"] = parentID
	}

	var result struct {
		CreateRiskAnalysisBoundary struct {
			RiskAnalysisBoundaryEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisBoundaryEdge"`
		} `json:"createRiskAnalysisBoundary"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysisBoundary mutation failed")

	return result.CreateRiskAnalysisBoundary.RiskAnalysisBoundaryEdge.Node.ID
}

func CreateRiskAnalysisProcess(c *testutil.Client, scopeID, sourceNodeID, targetNodeID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisProcessInput!) {
			createRiskAnalysisProcess(input: $input) {
				riskAnalysisProcessEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskAnalysisDiagramId": scopeID,
		"sourceNodeId":          sourceNodeID,
		"targetNodeId":          targetNodeID,
		"name":                  a.getString("name", SafeName("Process")),
	}

	var result struct {
		CreateRiskAnalysisProcess struct {
			RiskAnalysisProcessEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisProcessEdge"`
		} `json:"createRiskAnalysisProcess"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysisProcess mutation failed")

	return result.CreateRiskAnalysisProcess.RiskAnalysisProcessEdge.Node.ID
}

func CreateRiskAnalysisThreat(c *testutil.Client, scopeID, processID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisThreatInput!) {
			createRiskAnalysisThreat(input: $input) {
				riskAnalysisThreatEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskAnalysisDiagramId": scopeID,
		"processId":             processID,
		"name":                  a.getString("name", SafeName("Threat")),
		"category":              a.getString("category", "Confidentiality"),
	}

	var result struct {
		CreateRiskAnalysisThreat struct {
			RiskAnalysisThreatEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisThreatEdge"`
		} `json:"createRiskAnalysisThreat"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysisThreat mutation failed")

	return result.CreateRiskAnalysisThreat.RiskAnalysisThreatEdge.Node.ID
}

func CreateRiskAnalysisScenario(c *testutil.Client, scopeID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateRiskAnalysisScenarioInput!) {
			createRiskAnalysisScenario(input: $input) {
				riskAnalysisScenarioEdge { node { id } }
			}
		}
	`

	input := map[string]any{
		"riskAnalysisDiagramId": scopeID,
		"name":                  a.getString("name", SafeName("Scenario")),
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
	}

	var result struct {
		CreateRiskAnalysisScenario struct {
			RiskAnalysisScenarioEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"riskAnalysisScenarioEdge"`
		} `json:"createRiskAnalysisScenario"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createRiskAnalysisScenario mutation failed")

	return result.CreateRiskAnalysisScenario.RiskAnalysisScenarioEdge.Node.ID
}

func LinkRiskAnalysisScenarioThreat(c *testutil.Client, scenarioID, threatID string) {
	c.T.Helper()

	const query = `
		mutation($input: LinkRiskAnalysisScenarioThreatInput!) {
			linkRiskAnalysisScenarioThreat(input: $input) {
				riskAnalysisScenario { id }
			}
		}
	`

	_, err := c.Do(query, map[string]any{
		"input": map[string]any{
			"riskAnalysisScenarioId": scenarioID,
			"threatId":               threatID,
		},
	})
	require.NoError(c.T, err, "linkRiskAnalysisScenarioThreat mutation failed")
}

func LinkRiskToAnalysis(c *testutil.Client, riskID, riskAnalysisID string) {
	c.T.Helper()
	diagramID := CreateRiskAnalysisDiagram(c, riskAnalysisID)
	scenarioID := CreateRiskAnalysisScenario(c, diagramID)
	LinkRiskAnalysisScenarioRisk(c, scenarioID, riskID)
}

func LinkRiskAnalysisScenarioRisk(c *testutil.Client, scenarioID, riskID string) {
	c.T.Helper()

	const query = `
		mutation($input: LinkRiskAnalysisScenarioRiskInput!) {
			linkRiskAnalysisScenarioRisk(input: $input) {
				riskAnalysisScenario { id }
			}
		}
	`

	_, err := c.Do(query, map[string]any{
		"input": map[string]any{
			"riskAnalysisScenarioId": scenarioID,
			"riskId":                 riskID,
		},
	})
	require.NoError(c.T, err, "linkRiskAnalysisScenarioRisk mutation failed")
}
