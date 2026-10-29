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

package asana

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

type (
	taskPayload struct {
		GID        string `json:"gid"`
		Name       string `json:"name"`
		HTMLNotes  string `json:"html_notes"`
		Notes      string `json:"notes"`
		Completed  bool   `json:"completed"`
		DueOn      string `json:"due_on"`
		DueAt      string `json:"due_at"`
		Permalink  string `json:"permalink_url"`
		ModifiedAt string `json:"modified_at"`
		Assignee   *struct {
			GID   string `json:"gid"`
			Email string `json:"email"`
		} `json:"assignee"`
		Memberships []struct {
			Project *struct {
				GID string `json:"gid"`
			} `json:"project"`
			Section *struct {
				GID string `json:"gid"`
			} `json:"section"`
		} `json:"memberships"`
		CustomFields []struct {
			GID       string `json:"gid"`
			EnumValue *struct {
				GID string `json:"gid"`
			} `json:"enum_value"`
		} `json:"custom_fields"`
	}

	storyPayload struct {
		GID             string `json:"gid"`
		ResourceSubtype string `json:"resource_subtype"`
		HTMLText        string `json:"html_text"`
		Text            string `json:"text"`
		CreatedAt       string `json:"created_at"`
		CreatedBy       *struct {
			Email string `json:"email"`
		} `json:"created_by"`
	}

	attachmentPayload struct {
		GID         string `json:"gid"`
		Name        string `json:"name"`
		DownloadURL string `json:"download_url"`
		Parent      *struct {
			GID          string `json:"gid"`
			ResourceType string `json:"resource_type"`
		} `json:"parent"`
	}
)

func (p taskPayload) task() Task {
	task := Task{
		GID:       p.GID,
		Name:      p.Name,
		HTMLNotes: p.HTMLNotes,
		Notes:     p.Notes,
		Completed: p.Completed,
		DueOn:     p.DueOn,
		DueAt:     p.DueAt,
		Permalink: p.Permalink,
	}

	if parsed, err := time.Parse(time.RFC3339, p.ModifiedAt); err == nil {
		task.ModifiedAt = parsed
	}

	if p.Assignee != nil {
		task.AssigneeGID = p.Assignee.GID
		task.AssigneeEmail = p.Assignee.Email
	}

	for _, membership := range p.Memberships {
		item := Membership{}
		if membership.Project != nil {
			item.ProjectGID = membership.Project.GID
		}

		if membership.Section != nil {
			item.SectionGID = membership.Section.GID
		}

		task.Memberships = append(task.Memberships, item)
	}

	for _, field := range p.CustomFields {
		value := CustomFieldValue{GID: field.GID}
		if field.EnumValue != nil {
			value.EnumOptionGID = field.EnumValue.GID
		}

		task.CustomFields = append(task.CustomFields, value)
	}

	return task
}

func (p storyPayload) story() Story {
	story := Story{
		GID:      p.GID,
		HTMLText: p.HTMLText,
		Text:     p.Text,
	}

	if p.CreatedBy != nil {
		story.AuthorEmail = p.CreatedBy.Email
	}

	if parsed, err := time.Parse(time.RFC3339, p.CreatedAt); err == nil {
		story.CreatedAt = parsed
	}

	return story
}

func (p attachmentPayload) attachment() Attachment {
	attachment := Attachment{
		GID:         p.GID,
		Name:        p.Name,
		DownloadURL: p.DownloadURL,
	}
	if p.Parent != nil {
		attachment.ParentGID = p.Parent.GID
		attachment.ParentType = p.Parent.ResourceType
	}

	return attachment
}

func collectStories(ctx context.Context, c *Client, taskGID string) ([]storyPayload, error) {
	return collectPages[storyPayload](
		ctx,
		c,
		[]string{"tasks", taskGID, "stories"},
		url.Values{"opt_fields": {"gid,resource_subtype,html_text,text,created_at,created_by.email"}},
	)
}

func collectAttachments(ctx context.Context, c *Client, taskGID string) ([]attachmentPayload, error) {
	return collectPages[attachmentPayload](
		ctx,
		c,
		[]string{"tasks", taskGID, "attachments"},
		url.Values{"opt_fields": {"gid,name,download_url,parent.gid,parent.resource_type"}},
	)
}

func collectPages[T any](
	ctx context.Context,
	c *Client,
	segments []string,
	values url.Values,
) ([]T, error) {
	var all []T

	offset := ""

	for range listMaxPages {
		pageValues := url.Values{}
		for key, items := range values {
			pageValues[key] = append([]string(nil), items...)
		}

		pageValues.Set("limit", "100")

		if offset != "" {
			pageValues.Set("offset", offset)
		}

		var page []T

		next, err := c.getPage(ctx, segments, pageValues, &page)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)
		if next == "" {
			return all, nil
		}

		offset = next
	}

	return nil, fmt.Errorf("cannot list asana resources: pagination limit reached")
}
