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
		Documents          []MessageResource `json:"documents"`
		Reports            []MessageResource `json:"reports"`
		Files              []MessageResource `json:"files"`
	}
)

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
		"access",
	)
	if err != nil {
		return bot.MessageIntent{}, fmt.Errorf("cannot build access request URL: %w", err)
	}

	requester := attributes.RequesterEmail
	if attributes.RequesterName != "" {
		requester = attributes.RequesterName + " <" + attributes.RequesterEmail + ">"
	}

	cards := []bot.CardIntent{{
		ID:       "compliance-access-request",
		Title:    "New Compliance Page Access Request",
		Subtitle: "Requested by " + requester,
		Body: fmt.Sprintf(
			"%d document(s), %d audit report(s), and %d file(s) are awaiting review.",
			len(attributes.Documents),
			len(attributes.Reports),
			len(attributes.Files),
		),
		Actions: []bot.ActionIntent{
			{
				ID:    AccessCapability + ".approve_all",
				Label: "Approve all",
				Style: bot.ActionStylePrimary,
			},
			{
				ID:    AccessCapability + ".deny_all",
				Label: "Deny all",
				Style: bot.ActionStyleDanger,
			},
			{
				Label: "Open in Probo",
				URL:   requestURL,
			},
		},
	}}
	for _, resource := range attributes.Documents {
		resourceURL, err := url.JoinPath(
			r.baseURL,
			"organizations",
			url.PathEscape(message.OrganizationID.String()),
			"documents",
			url.PathEscape(resource.ID),
		)
		if err != nil {
			return bot.MessageIntent{}, fmt.Errorf("cannot build document URL: %w", err)
		}

		cards = append(cards, resourceCard("Document", resource.Title, resourceURL, resource))
	}

	for _, resource := range attributes.Reports {
		resourceURL, err := url.JoinPath(
			r.baseURL,
			"organizations",
			url.PathEscape(message.OrganizationID.String()),
			"audits",
			url.PathEscape(resource.AuditID),
		)
		if err != nil {
			return bot.MessageIntent{}, fmt.Errorf("cannot build audit URL: %w", err)
		}

		cards = append(cards, resourceCard("Audit report", resource.Title, resourceURL, resource))
	}

	for _, resource := range attributes.Files {
		subtitle := "File"
		if resource.Category != "" {
			subtitle += " · " + resource.Category
		}

		cards = append(cards, resourceCard(subtitle, resource.Name, requestURL, resource))
	}

	return bot.MessageIntent{
		FallbackText: "New Compliance Page Access Request",
		Cards:        cards,
	}, nil
}

func resourceCard(
	subtitle string,
	title string,
	titleURL string,
	resource MessageResource,
) bot.CardIntent {
	status := statusLabel(resource.Status)

	card := bot.CardIntent{
		ID:       resource.ID,
		Title:    title,
		TitleURL: titleURL,
		Subtitle: subtitle,
		Body:     "Status: " + status,
	}
	if resource.Status == "REQUESTED" {
		card.Actions = []bot.ActionIntent{
			{
				ID:    AccessCapability + ".approve_item",
				Label: "Approve",
				Style: bot.ActionStylePrimary,
				Value: resource.ID,
			},
			{
				ID:    AccessCapability + ".deny_item",
				Label: "Deny",
				Style: bot.ActionStyleDanger,
				Value: resource.ID,
			},
		}
	}

	return card
}

func statusLabel(status string) string {
	switch status {
	case "GRANTED":
		return "Approved"
	case "REJECTED":
		return "Denied"
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
