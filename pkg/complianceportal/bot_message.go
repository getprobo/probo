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

package complianceportal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	Renderer struct {
		baseURL string
	}

	MessageResource struct {
		ID       string `json:"ID"`
		Title    string `json:"Title"`
		Name     string `json:"Name"`
		AuditID  string `json:"AuditID"`
		Category string `json:"Category"`
		Status   string `json:"Status"`
	}

	MessageAttributes struct {
		RequesterName      string            `json:"requester_name"`
		RequesterEmail     string            `json:"requester_email"`
		CompliancePortalID string            `json:"compliance_portal_id"`
		AccessID           string            `json:"compliance_portal_access_id"`
		Documents          []MessageResource `json:"documents"`
		Reports            []MessageResource `json:"reports"`
		Files              []MessageResource `json:"files"`
	}
)

const accessRequestHeadline = "New compliance portal access request"

var _ bot.MessageRenderer = (*Renderer)(nil)

func NewRenderer(baseURL string) *Renderer {
	return &Renderer{baseURL: baseURL}
}

func (r *Renderer) RenderMessage(
	_ context.Context,
	message bot.Message,
) (bot.MessageIntent, error) {
	attributes, err := DecodeMessageAttributes(message.Attributes)
	if err != nil {
		return bot.MessageIntent{}, err
	}

	requestURL, err := url.JoinPath(
		r.baseURL,
		"organizations",
		url.PathEscape(message.OrganizationID.String()),
		"compliance-portals",
		url.PathEscape(attributes.CompliancePortalID),
		"visitors",
		url.PathEscape(attributes.AccessID),
	)
	if err != nil {
		return bot.MessageIntent{}, fmt.Errorf("cannot build access request URL: %w", err)
	}

	requester := attributes.RequesterEmail
	if attributes.RequesterName != "" {
		requester = attributes.RequesterName + " <" + attributes.RequesterEmail + ">"
	}

	documents := make([]bot.ItemIntent, 0, len(attributes.Documents))
	for _, resource := range attributes.Documents {
		resourceURL, err := url.JoinPath(
			r.baseURL,
			"organizations",
			url.PathEscape(message.OrganizationID.String()),
			"governance",
			"documents",
			url.PathEscape(resource.ID),
		)
		if err != nil {
			return bot.MessageIntent{}, fmt.Errorf("cannot build document URL: %w", err)
		}

		documents = append(documents, resourceItem(resource.Title, resourceURL, resource))
	}

	reports := make([]bot.ItemIntent, 0, len(attributes.Reports))
	for _, resource := range attributes.Reports {
		resourceURL, err := url.JoinPath(
			r.baseURL,
			"organizations",
			url.PathEscape(message.OrganizationID.String()),
			"governance",
			"audits",
			url.PathEscape(resource.AuditID),
		)
		if err != nil {
			return bot.MessageIntent{}, fmt.Errorf("cannot build audit URL: %w", err)
		}

		reports = append(reports, resourceItem(resource.Title, resourceURL, resource))
	}

	files := make([]bot.ItemIntent, 0, len(attributes.Files))
	for _, resource := range attributes.Files {
		label := resource.Name
		if resource.Category != "" {
			label += " (" + resource.Category + ")"
		}

		files = append(files, resourceItem(label, requestURL, resource))
	}

	groups := make([]bot.GroupIntent, 0, 3)

	for _, group := range []bot.GroupIntent{
		{ID: "documents", Title: groupTitle("Documents", documents), Items: documents},
		{ID: "reports", Title: groupTitle("Audit reports", reports), Items: reports},
		{ID: "files", Title: groupTitle("Files", files), Items: files},
	} {
		if len(group.Items) > 0 {
			groups = append(groups, group)
		}
	}

	return bot.MessageIntent{
		FallbackText: accessRequestHeadline,
		Headline:     "🔒 " + accessRequestHeadline,
		Context:      "Requested by " + requester,
		Actions: []bot.ActionIntent{
			{
				ID:    AccessCapability + ".approve_all",
				Label: "Grant all",
				Style: bot.ActionStylePrimary,
			},
			{
				ID:    AccessCapability + ".deny_all",
				Label: "Reject/revoke all",
				Style: bot.ActionStyleDanger,
			},
			{
				Label: "Open in Probo",
				URL:   requestURL,
			},
		},
		Groups: groups,
	}, nil
}

// resourceItem gives a status only to decided resources: on a pending row the
// review menu already carries that meaning, and repeating it on every row of a
// long request is noise.
func resourceItem(
	label string,
	labelURL string,
	resource MessageResource,
) bot.ItemIntent {
	item := bot.ItemIntent{
		ID:    resource.ID,
		Label: label,
		URL:   labelURL,
	}

	switch resource.Status {
	case "GRANTED":
		item.Status = statusLabel(resource.Status)
		item.Action = &bot.ActionIntent{
			ID:    AccessCapability + ".deny_item",
			Label: "Revoke",
			Style: bot.ActionStyleDanger,
			Value: resource.ID,
		}
	case "REJECTED", "REVOKED":
		item.Status = statusLabel(resource.Status)
		item.Action = &bot.ActionIntent{
			ID:    AccessCapability + ".approve_item",
			Label: "Grant",
			Style: bot.ActionStylePrimary,
			Value: resource.ID,
		}
	default:
		item.Action = &bot.ActionIntent{
			ID: AccessCapability + ".review_item",
			Options: []bot.ActionOptionIntent{
				{Label: "Grant", Value: "approve/" + resource.ID},
				{Label: "Reject", Value: "reject/" + resource.ID},
			},
		}
	}

	return item
}

func groupTitle(name string, items []bot.ItemIntent) string {
	return fmt.Sprintf("%s (%d)", name, len(items))
}

func statusLabel(status string) string {
	switch status {
	case "GRANTED":
		return "Granted"
	case "REJECTED":
		return "Rejected"
	case "REVOKED":
		return "Revoked"
	default:
		return "Requested"
	}
}

func RequesterEmailFromMessage(message bot.Message) (mail.Addr, bool) {
	raw, ok := message.Attributes[RequesterEmailAttribute].(string)
	if !ok || raw == "" {
		return "", false
	}

	addr, err := mail.ParseAddr(raw)
	if err != nil {
		return "", false
	}

	return addr, true
}

func DecodeMessageAttributes(attributes map[string]any) (MessageAttributes, error) {
	raw, err := json.Marshal(attributes)
	if err != nil {
		return MessageAttributes{}, fmt.Errorf("cannot encode message attributes: %w", err)
	}

	var decoded MessageAttributes
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return MessageAttributes{}, fmt.Errorf("cannot decode message attributes: %w", err)
	}

	if _, err := gid.ParseGID(decoded.CompliancePortalID); err != nil {
		return MessageAttributes{}, fmt.Errorf("invalid compliance portal ID")
	}

	accessID, err := gid.ParseGID(decoded.AccessID)
	if err != nil || accessID.EntityType() != coredata.CompliancePortalAccessEntityType {
		return MessageAttributes{}, fmt.Errorf("invalid compliance portal access ID")
	}

	if decoded.RequesterEmail != "" {
		if _, err := mail.ParseAddr(decoded.RequesterEmail); err != nil {
			return MessageAttributes{}, fmt.Errorf("invalid requester email")
		}
	}

	for _, resource := range append(
		append(decoded.Documents, decoded.Reports...),
		decoded.Files...,
	) {
		if _, err := gid.ParseGID(resource.ID); err != nil {
			return MessageAttributes{}, fmt.Errorf("invalid resource ID")
		}
	}

	return decoded, nil
}
