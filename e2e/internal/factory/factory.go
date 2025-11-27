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

// SafeName generates a name without special characters that might cause validation issues.
func SafeName(prefix string) string {
	return fmt.Sprintf("%s %s", prefix, gofakeit.LetterN(8))
}

// SafeEmail generates an email using example.com domain to avoid validation issues.
func SafeEmail() string {
	return fmt.Sprintf("%s@example.com", strings.ToLower(gofakeit.LetterN(12)))
}

// Attrs is a map of attribute overrides for factory functions.
type Attrs map[string]any

// get returns the value for key if present, otherwise returns defaultVal.
func (a Attrs) get(key string, defaultVal any) any {
	if a == nil {
		return defaultVal
	}
	if v, ok := a[key]; ok {
		return v
	}
	return defaultVal
}

// getString returns the string value for key if present, otherwise returns defaultVal.
func (a Attrs) getString(key string, defaultVal string) string {
	if v, ok := a.get(key, defaultVal).(string); ok {
		return v
	}
	return defaultVal
}

// getStringPtr returns a pointer to the string value for key if present.
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

// getInt returns the int value for key if present, otherwise returns defaultVal.
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

// CreateVendor creates a vendor with optional attribute overrides.
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

// CreateFramework creates a framework with optional attribute overrides.
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

// CreateControl creates a control with optional attribute overrides.
// The frameworkId is required and must be passed in attrs.
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
		"status":       a.getString("status", "INCLUDED"),
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

// CreateMeasure creates a measure with optional attribute overrides.
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

// CreateTask creates a task with optional attribute overrides.
// The measureId is required.
func CreateTask(c *testutil.Client, measureID string, attrs ...Attrs) string {
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
		"measureId":      measureID,
		"name":           a.getString("name", SafeName("Task")),
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

// CreateRisk creates a risk with optional attribute overrides.
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

// CreatePeople creates a people record with optional attribute overrides.
func CreatePeople(c *testutil.Client, attrs ...Attrs) string {
	c.T.Helper()

	var a Attrs
	if len(attrs) > 0 {
		a = attrs[0]
	}

	const query = `
		mutation($input: CreatePeopleInput!) {
			createPeople(input: $input) {
				peopleEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":      c.GetOrganizationID().String(),
		"fullName":            a.getString("fullName", SafeName("Person")),
		"primaryEmailAddress": a.getString("primaryEmailAddress", SafeEmail()),
		"kind":                a.getString("kind", "EMPLOYEE"),
	}

	var result struct {
		CreatePeople struct {
			PeopleEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"peopleEdge"`
		} `json:"createPeople"`
	}

	err := c.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(c.T, err, "createPeople mutation failed")

	return result.CreatePeople.PeopleEdge.Node.ID
}

// =============================================================================
// Builder-style wrappers for backward compatibility
// These provide the fluent builder API that wraps the simpler factory functions
// =============================================================================

// VendorBuilder provides a fluent API for creating vendors.
type VendorBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewVendor creates a new VendorBuilder.
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

// FrameworkBuilder provides a fluent API for creating frameworks.
type FrameworkBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewFramework creates a new FrameworkBuilder.
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

// ControlBuilder provides a fluent API for creating controls.
type ControlBuilder struct {
	client      *testutil.Client
	frameworkID string
	attrs       Attrs
}

// NewControl creates a new ControlBuilder.
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

func (b *ControlBuilder) Create() string {
	return CreateControl(b.client, b.frameworkID, b.attrs)
}

// MeasureBuilder provides a fluent API for creating measures.
type MeasureBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewMeasure creates a new MeasureBuilder.
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

// TaskBuilder provides a fluent API for creating tasks.
type TaskBuilder struct {
	client    *testutil.Client
	measureID string
	attrs     Attrs
}

// NewTask creates a new TaskBuilder.
func NewTask(c *testutil.Client, measureID string) *TaskBuilder {
	return &TaskBuilder{client: c, measureID: measureID, attrs: Attrs{}}
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

// RiskBuilder provides a fluent API for creating risks.
type RiskBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewRisk creates a new RiskBuilder.
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

// PeopleBuilder provides a fluent API for creating people.
type PeopleBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewPeople creates a new PeopleBuilder.
func NewPeople(c *testutil.Client) *PeopleBuilder {
	return &PeopleBuilder{client: c, attrs: Attrs{}}
}

func (b *PeopleBuilder) WithFullName(name string) *PeopleBuilder {
	b.attrs["fullName"] = name
	return b
}

func (b *PeopleBuilder) WithKind(kind string) *PeopleBuilder {
	b.attrs["kind"] = kind
	return b
}

func (b *PeopleBuilder) Create() string {
	return CreatePeople(b.client, b.attrs)
}

// CreateAudit creates an audit with optional attribute overrides.
// The frameworkId is required.
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

// AuditBuilder provides a fluent API for creating audits.
type AuditBuilder struct {
	client      *testutil.Client
	frameworkID string
	attrs       Attrs
}

// NewAudit creates a new AuditBuilder.
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

// CreateDatum creates a datum with optional attribute overrides.
// The ownerId (peopleId) is required.
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

// DatumBuilder provides a fluent API for creating data.
type DatumBuilder struct {
	client  *testutil.Client
	ownerID string
	attrs   Attrs
}

// NewDatum creates a new DatumBuilder.
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

// CreateMeeting creates a meeting with optional attribute overrides.
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

// MeetingBuilder provides a fluent API for creating meetings.
type MeetingBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewMeeting creates a new MeetingBuilder.
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

// CreateDocument creates a document with optional attribute overrides.
// The ownerId (peopleId) is required.
func CreateDocument(c *testutil.Client, ownerID string, attrs ...Attrs) string {
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
		"ownerId":        ownerID,
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

// DocumentBuilder provides a fluent API for creating documents.
type DocumentBuilder struct {
	client  *testutil.Client
	ownerID string
	attrs   Attrs
}

// NewDocument creates a new DocumentBuilder.
func NewDocument(c *testutil.Client, ownerID string) *DocumentBuilder {
	return &DocumentBuilder{client: c, ownerID: ownerID, attrs: Attrs{}}
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
	return CreateDocument(b.client, b.ownerID, b.attrs)
}

// CreateProcessingActivity creates a processing activity with optional attribute overrides.
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
		"organizationId":                 c.GetOrganizationID().String(),
		"name":                           a.getString("name", SafeName("ProcessingActivity")),
		"specialOrCriminalData":          a.getString("specialOrCriminalData", "NO"),
		"lawfulBasis":                    a.getString("lawfulBasis", "CONSENT"),
		"internationalTransfers":         a.getBool("internationalTransfers", false),
		"dataProtectionImpactAssessment": a.getString("dataProtectionImpactAssessment", "NOT_NEEDED"),
		"transferImpactAssessment":       a.getString("transferImpactAssessment", "NOT_NEEDED"),
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

// ProcessingActivityBuilder provides a fluent API for creating processing activities.
type ProcessingActivityBuilder struct {
	client *testutil.Client
	attrs  Attrs
}

// NewProcessingActivity creates a new ProcessingActivityBuilder.
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
