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

package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type (
	Client struct {
		httpClient *http.Client
		endpoint   string
	}

	Team struct {
		ID   string
		Name string
		Key  string
	}

	WorkflowState struct {
		ID       string
		Name     string
		Type     string
		Position float64
	}

	Issue struct {
		ID         string
		Identifier string
		URL        string
		Title      string
		UpdatedAt  time.Time
	}

	IssueInput struct {
		TeamID      string
		Title       string
		Description string
		StateID     string
		Priority    int
		DueDate     *string
	}

	IssueUpdateInput struct {
		Title       *string
		Description *string
		StateID     *string
		Priority    *int
		DueDate     *string
		DueDateSet  bool
	}

	graphqlRequest struct {
		Query     string `json:"query"`
		Variables any    `json:"variables"`
	}

	graphqlError struct {
		Message string `json:"message"`
	}
)

// NewClient returns a Linear GraphQL client rooted at endpoint, which callers
// take from the LINEAR provider registration's Endpoints.APIBase rather than
// pinning, so an endpoint override moves task sync along with the OAuth
// handshake that mints the token they carry.
func NewClient(httpClient *http.Client, endpoint string) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

func (c *Client) ViewerID(ctx context.Context) (string, error) {
	const query = `
query TaskSyncLinearViewer {
  viewer {
    id
  }
}
`

	var resp struct {
		Data struct {
			Viewer struct {
				ID string `json:"id"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{}, &resp); err != nil {
		return "", err
	}

	if resp.Data.Viewer.ID == "" {
		return "", fmt.Errorf("cannot load Linear viewer: empty id")
	}

	return resp.Data.Viewer.ID, nil
}

func (c *Client) OrganizationID(ctx context.Context) (string, error) {
	const query = `
query TaskSyncLinearOrganization {
  organization {
    id
  }
}
`

	var resp struct {
		Data struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{}, &resp); err != nil {
		return "", err
	}

	if resp.Data.Organization.ID == "" {
		return "", fmt.Errorf("cannot load Linear organization: empty id")
	}

	return resp.Data.Organization.ID, nil
}

func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	const query = `
query TaskSyncLinearTeams {
  teams {
    nodes {
      id
      name
      key
    }
  }
}
`

	var resp struct {
		Data struct {
			Teams struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Key  string `json:"key"`
				} `json:"nodes"`
			} `json:"teams"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{}, &resp); err != nil {
		return nil, err
	}

	teams := make([]Team, 0, len(resp.Data.Teams.Nodes))
	for _, node := range resp.Data.Teams.Nodes {
		teams = append(teams, Team{
			ID:   node.ID,
			Name: node.Name,
			Key:  node.Key,
		})
	}

	return teams, nil
}

func (c *Client) ListWorkflowStates(ctx context.Context, teamID string) ([]WorkflowState, error) {
	const query = `
query TaskSyncLinearWorkflowStates($teamId: ID!) {
  workflowStates(filter: { team: { id: { eq: $teamId } } }) {
    nodes {
      id
      name
      type
      position
    }
  }
}
`

	var resp struct {
		Data struct {
			WorkflowStates struct {
				Nodes []struct {
					ID       string  `json:"id"`
					Name     string  `json:"name"`
					Type     string  `json:"type"`
					Position float64 `json:"position"`
				} `json:"nodes"`
			} `json:"workflowStates"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"teamId": teamID}, &resp); err != nil {
		return nil, err
	}

	states := make([]WorkflowState, 0, len(resp.Data.WorkflowStates.Nodes))
	for _, node := range resp.Data.WorkflowStates.Nodes {
		states = append(states, WorkflowState{
			ID:       node.ID,
			Name:     node.Name,
			Type:     node.Type,
			Position: node.Position,
		})
	}

	return states, nil
}

func (c *Client) CreateIssue(ctx context.Context, input IssueInput) (*Issue, error) {
	const query = `
mutation TaskSyncLinearIssueCreate($input: IssueCreateInput!) {
  issueCreate(input: $input) {
    success
    issue {
      id
      identifier
      url
      title
      updatedAt
    }
  }
}
`

	vars := map[string]any{
		"input": map[string]any{
			"teamId":      input.TeamID,
			"title":       input.Title,
			"description": input.Description,
			"stateId":     input.StateID,
			"priority":    input.Priority,
		},
	}

	if input.DueDate != nil {
		vars["input"].(map[string]any)["dueDate"] = *input.DueDate
	}

	var resp struct {
		Data struct {
			IssueCreate struct {
				Success bool        `json:"success"`
				Issue   linearIssue `json:"issue"`
			} `json:"issueCreate"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, vars, &resp); err != nil {
		return nil, err
	}

	if !resp.Data.IssueCreate.Success {
		return nil, fmt.Errorf("cannot create Linear issue: mutation unsuccessful")
	}

	return resp.Data.IssueCreate.Issue.toIssue()
}

func (c *Client) UpdateIssue(ctx context.Context, issueID string, input IssueUpdateInput) (*Issue, error) {
	const query = `
mutation TaskSyncLinearIssueUpdate($id: String!, $input: IssueUpdateInput!) {
  issueUpdate(id: $id, input: $input) {
    success
    issue {
      id
      identifier
      url
      title
      updatedAt
    }
  }
}
`

	update := map[string]any{}
	if input.Title != nil {
		update["title"] = *input.Title
	}

	if input.Description != nil {
		update["description"] = *input.Description
	}

	if input.StateID != nil {
		update["stateId"] = *input.StateID
	}

	if input.Priority != nil {
		update["priority"] = *input.Priority
	}

	if input.DueDateSet {
		if input.DueDate != nil {
			update["dueDate"] = *input.DueDate
		} else {
			update["dueDate"] = nil
		}
	}

	var resp struct {
		Data struct {
			IssueUpdate struct {
				Success bool        `json:"success"`
				Issue   linearIssue `json:"issue"`
			} `json:"issueUpdate"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"id": issueID, "input": update}, &resp); err != nil {
		return nil, err
	}

	if !resp.Data.IssueUpdate.Success {
		return nil, fmt.Errorf("cannot update Linear issue: mutation unsuccessful")
	}

	return resp.Data.IssueUpdate.Issue.toIssue()
}

func (c *Client) LinkAttachment(ctx context.Context, issueID, url, title string) (string, error) {
	const query = `
mutation TaskSyncLinearAttachmentLink($issueId: String!, $url: String!, $title: String) {
  attachmentLinkURL(issueId: $issueId, url: $url, title: $title) {
    success
    attachment {
      id
    }
  }
}
`

	var resp struct {
		Data struct {
			AttachmentLinkURL struct {
				Success    bool `json:"success"`
				Attachment struct {
					ID string `json:"id"`
				} `json:"attachment"`
			} `json:"attachmentLinkURL"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{
		"issueId": issueID,
		"url":     url,
		"title":   title,
	}, &resp); err != nil {
		return "", err
	}

	if !resp.Data.AttachmentLinkURL.Success {
		return "", fmt.Errorf("cannot link Linear attachment: mutation unsuccessful")
	}

	return resp.Data.AttachmentLinkURL.Attachment.ID, nil
}

func (c *Client) do(ctx context.Context, query string, variables any, dest any) error {
	payload, err := json.Marshal(
		graphqlRequest{
			Query:     query,
			Variables: variables,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal Linear graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("cannot create Linear graphql request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute Linear graphql request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("cannot call Linear graphql: unexpected status %d", httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(dest); err != nil {
		return fmt.Errorf("cannot decode Linear graphql response: %w", err)
	}

	if errs := extractGraphQLErrors(dest); len(errs) > 0 {
		return fmt.Errorf("cannot call Linear graphql: graphql error")
	}

	return nil
}

type linearIssue struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	UpdatedAt  string `json:"updatedAt"`
}

func (i linearIssue) toIssue() (*Issue, error) {
	updatedAt, err := time.Parse(time.RFC3339, i.UpdatedAt)
	if err != nil {
		updatedAt, err = time.Parse(time.RFC3339Nano, i.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("cannot parse Linear issue updatedAt: %w", err)
		}
	}

	return &Issue{
		ID:         i.ID,
		Identifier: i.Identifier,
		URL:        i.URL,
		Title:      i.Title,
		UpdatedAt:  updatedAt,
	}, nil
}

func extractGraphQLErrors(dest any) []graphqlError {
	raw, err := json.Marshal(dest)
	if err != nil {
		return nil
	}

	var envelope struct {
		Errors []graphqlError `json:"errors"`
	}

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil
	}

	return envelope.Errors
}
