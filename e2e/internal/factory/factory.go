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

// Package factory provides Rails-like test data factories using gofakeit.
package factory

import (
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
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

func CreateVendor(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateVendorInput!) {
			createVendor(input: $input) {
				vendorEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("Vendor")),
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

	var result struct {
		CreateVendor struct {
			VendorEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorEdge"`
		} `json:"createVendor"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createVendor mutation failed")

	return result.CreateVendor.VendorEdge.Node.ID
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
		"frameworkId":  frameworkID,
		"name":         a.getString("name", SafeName("Control")),
		"description":  a.getString("description", "Test control description"),
		"sectionTitle": a.getString("sectionTitle", fmt.Sprintf("Section %s", gofakeit.LetterN(3))),
		"bestPractice": a.getBool("bestPractice", true),
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
	}
	if measureID != nil {
		input["measureId"] = *measureID
	}
	if desc := a.getStringPtr("description"); desc != nil {
		input["description"] = *desc
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

type VendorBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewVendor(c *testutil.Client) *VendorBuilder {
	return &VendorBuilder{client: c, attrs: Attrs{}}
}

func (b *VendorBuilder) WithName(name string) *VendorBuilder {
	b.attrs["name"] = name
	return b
}

func (b *VendorBuilder) WithDescription(desc string) *VendorBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *VendorBuilder) WithWebsiteUrl(url string) *VendorBuilder {
	b.attrs["websiteUrl"] = url
	return b
}

func (b *VendorBuilder) WithCategory(category string) *VendorBuilder {
	b.attrs["category"] = category
	return b
}

func (b *VendorBuilder) Create() string {
	return CreateVendor(b.client, b.attrs)
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

func (b *TaskBuilder) WithDescription(desc string) *TaskBuilder {
	b.attrs["description"] = desc
	return b
}

func (b *TaskBuilder) Create() string {
	return CreateTask(b.client, b.measureID, b.attrs)
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

func CreateMeeting(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateMeetingInput!) {
			createMeeting(input: $input) {
				meetingEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"name":           a.getString("name", SafeName("Meeting")),
		"date":           a.getString("date", "2025-01-15T10:00:00Z"),
	}
	if minutes := a.getStringPtr("minutes"); minutes != nil {
		input["minutes"] = *minutes
	}

	var result struct {
		CreateMeeting struct {
			MeetingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"meetingEdge"`
		} `json:"createMeeting"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createMeeting mutation failed")

	return result.CreateMeeting.MeetingEdge.Node.ID
}

type MeetingBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

func NewMeeting(c *testutil.Client) *MeetingBuilder {
	return &MeetingBuilder{client: c, attrs: Attrs{}}
}

func (b *MeetingBuilder) WithName(name string) *MeetingBuilder {
	b.attrs["name"] = name
	return b
}

func (b *MeetingBuilder) WithDate(date string) *MeetingBuilder {
	b.attrs["date"] = date
	return b
}

func (b *MeetingBuilder) WithMinutes(minutes string) *MeetingBuilder {
	b.attrs["minutes"] = minutes
	return b
}

func (b *MeetingBuilder) Create() string {
	return CreateMeeting(b.client, b.attrs)
}

func CreateDocument(c *testutil.Client, approverID string, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": c.GetOrganizationID().String(),
		"approverIds":    []string{approverID},
		"title":          a.getString("title", SafeName("Document")),
		"content":        a.getString("content", "Document content"),
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
		} `json:"createDocument"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createDocument mutation failed")

	return result.CreateDocument.DocumentEdge.Node.ID
}

type DocumentBuilder struct {
	client     *testutil.Client
	approverID string
	attrs      Attrs
}

func NewDocument(c *testutil.Client, approverID string) *DocumentBuilder {
	return &DocumentBuilder{client: c, approverID: approverID, attrs: Attrs{}}
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

func (b *DocumentBuilder) Create() string {
	return CreateDocument(b.client, b.approverID, b.attrs)
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
