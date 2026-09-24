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
	"errors"
	"fmt"
	"net/http"
	"strings"
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
		ID            string
		Identifier    string
		URL           string
		Title         string
		Description   string
		StateType     string
		Priority      int
		DueDate       string
		AssigneeEmail string
		TeamID        string
		UpdatedAt     time.Time
	}

	TeamPage struct {
		Teams       []Team
		EndCursor   string
		HasNextPage bool
	}

	IssuePage struct {
		Issues      []Issue
		EndCursor   string
		HasNextPage bool
	}

	IssueInput struct {
		TeamID      string
		Title       string
		Description string
		StateID     string
		Priority    int
		DueDate     *string
		AssigneeID  *string
	}

	IssueUpdateInput struct {
		Title       *string
		Description *string
		StateID     *string
		Priority    *int
		DueDate     *string
		DueDateSet  bool
		AssigneeID  *string
	}

	graphqlRequest struct {
		Query     string `json:"query"`
		Variables any    `json:"variables"`
	}

	graphqlError struct {
		Message string `json:"message"`
	}

	pageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	}
)

const (
	linearListPageSize = 100
	linearListMaxPages = 500
)

var ErrIssueNotFound = errors.New("linear issue was not found")

// NewClient returns a Linear GraphQL client rooted at endpoint, which callers
// take from the LINEAR_SYNC provider registration's Endpoints.APIBase rather
// than pinning, so an endpoint override moves task sync along with the OAuth
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

func (c *Client) UserIDByEmail(ctx context.Context, email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", nil
	}

	const query = `
query TaskSyncLinearUserByEmail($email: String!) {
  users(filter: { email: { eqIgnoreCase: $email } }, first: 2) {
    nodes {
      id
      email
    }
  }
}
`

	var resp struct {
		Data struct {
			Users struct {
				Nodes []struct {
					ID    string `json:"id"`
					Email string `json:"email"`
				} `json:"nodes"`
			} `json:"users"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"email": email}, &resp); err != nil {
		return "", err
	}

	if len(resp.Data.Users.Nodes) != 1 {
		return "", nil
	}

	return resp.Data.Users.Nodes[0].ID, nil
}

func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	const query = `
query TaskSyncLinearTeams($first: Int!, $after: String) {
  teams(first: $first, after: $after) {
    nodes {
      id
      name
      key
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

	var (
		teams []Team
		after *string
	)

	for range linearListMaxPages {
		var resp struct {
			Data struct {
				Teams struct {
					Nodes []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
						Key  string `json:"key"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"teams"`
			} `json:"data"`
			Errors []graphqlError `json:"errors"`
		}

		if err := c.do(
			ctx,
			query,
			map[string]any{
				"first": linearListPageSize,
				"after": after,
			},
			&resp,
		); err != nil {
			return nil, err
		}

		for _, node := range resp.Data.Teams.Nodes {
			teams = append(
				teams,
				Team{
					ID:   node.ID,
					Name: node.Name,
					Key:  node.Key,
				},
			)
		}

		if !resp.Data.Teams.PageInfo.HasNextPage || resp.Data.Teams.PageInfo.EndCursor == "" {
			return teams, nil
		}

		nextCursor := resp.Data.Teams.PageInfo.EndCursor
		after = &nextCursor
	}

	return nil, fmt.Errorf("cannot list Linear teams: pagination limit reached")
}

func (c *Client) ListWorkflowStates(ctx context.Context, teamID string) ([]WorkflowState, error) {
	const query = `
query TaskSyncLinearWorkflowStates($teamId: ID!, $first: Int!, $after: String) {
  workflowStates(filter: { team: { id: { eq: $teamId } } }, first: $first, after: $after) {
    nodes {
      id
      name
      type
      position
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

	var (
		states []WorkflowState
		after  *string
	)

	for range linearListMaxPages {
		var resp struct {
			Data struct {
				WorkflowStates struct {
					Nodes []struct {
						ID       string  `json:"id"`
						Name     string  `json:"name"`
						Type     string  `json:"type"`
						Position float64 `json:"position"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"workflowStates"`
			} `json:"data"`
			Errors []graphqlError `json:"errors"`
		}

		if err := c.do(
			ctx,
			query,
			map[string]any{
				"teamId": teamID,
				"first":  linearListPageSize,
				"after":  after,
			},
			&resp,
		); err != nil {
			return nil, err
		}

		for _, node := range resp.Data.WorkflowStates.Nodes {
			states = append(
				states,
				WorkflowState{
					ID:       node.ID,
					Name:     node.Name,
					Type:     node.Type,
					Position: node.Position,
				},
			)
		}

		if !resp.Data.WorkflowStates.PageInfo.HasNextPage || resp.Data.WorkflowStates.PageInfo.EndCursor == "" {
			return states, nil
		}

		nextCursor := resp.Data.WorkflowStates.PageInfo.EndCursor
		after = &nextCursor
	}

	return nil, fmt.Errorf("cannot list Linear workflow states: pagination limit reached")
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

	if input.AssigneeID != nil && *input.AssigneeID != "" {
		vars["input"].(map[string]any)["assigneeId"] = *input.AssigneeID
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

	if input.AssigneeID != nil && *input.AssigneeID != "" {
		update["assigneeId"] = *input.AssigneeID
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

func (c *Client) ArchiveIssue(ctx context.Context, issueID string) error {
	const query = `
mutation TaskSyncLinearIssueArchive($id: String!) {
  issueArchive(id: $id) {
    success
  }
}
`

	var resp struct {
		Data struct {
			IssueArchive struct {
				Success bool `json:"success"`
			} `json:"issueArchive"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"id": issueID}, &resp); err != nil {
		return err
	}

	if !resp.Data.IssueArchive.Success {
		return fmt.Errorf("cannot archive Linear issue: mutation unsuccessful")
	}

	return nil
}

func (c *Client) LinkAttachment(ctx context.Context, issueID, url, title string) (string, error) {
	// attachmentLinkURL fetches the page and rejects auth-gated task URLs
	// with "Unable to create issue attachment". attachmentCreate stores the link.
	const query = `
mutation TaskSyncLinearAttachmentCreate($input: AttachmentCreateInput!) {
  attachmentCreate(input: $input) {
    success
    attachment {
      id
    }
  }
}
`

	var resp struct {
		Data struct {
			AttachmentCreate struct {
				Success    bool `json:"success"`
				Attachment struct {
					ID string `json:"id"`
				} `json:"attachment"`
			} `json:"attachmentCreate"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{
		"input": map[string]any{
			"issueId": issueID,
			"url":     url,
			"title":   title,
		},
	}, &resp); err != nil {
		return "", err
	}

	if !resp.Data.AttachmentCreate.Success || resp.Data.AttachmentCreate.Attachment.ID == "" {
		return "", fmt.Errorf("cannot link Linear attachment: mutation unsuccessful")
	}

	return resp.Data.AttachmentCreate.Attachment.ID, nil
}

func (c *Client) DeleteAttachment(ctx context.Context, attachmentID string) error {
	const query = `
mutation TaskSyncLinearAttachmentDelete($id: String!) {
  attachmentDelete(id: $id) {
    success
  }
}
`

	var resp struct {
		Data struct {
			AttachmentDelete struct {
				Success bool `json:"success"`
			} `json:"attachmentDelete"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"id": attachmentID}, &resp); err != nil {
		return err
	}

	if !resp.Data.AttachmentDelete.Success {
		return fmt.Errorf("cannot delete Linear attachment: mutation unsuccessful")
	}

	return nil
}

func (c *Client) SearchTeams(ctx context.Context, query string, first int, after *string) (TeamPage, error) {
	const request = `
query TaskSyncLinearTeamSearch($first: Int!, $after: String, $filter: TeamFilter) {
  teams(first: $first, after: $after, filter: $filter) {
    nodes {
      id
      name
      key
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

	vars := map[string]any{
		"first":  first,
		"after":  after,
		"filter": teamSearchFilter(query),
	}

	var resp struct {
		Data struct {
			Teams struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Key  string `json:"key"`
				} `json:"nodes"`
				PageInfo pageInfo `json:"pageInfo"`
			} `json:"teams"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, request, vars, &resp); err != nil {
		return TeamPage{}, err
	}

	teams := make([]Team, 0, len(resp.Data.Teams.Nodes))
	for _, node := range resp.Data.Teams.Nodes {
		teams = append(teams, Team{
			ID:   node.ID,
			Name: node.Name,
			Key:  node.Key,
		})
	}

	return TeamPage{
		Teams:       teams,
		EndCursor:   resp.Data.Teams.PageInfo.EndCursor,
		HasNextPage: resp.Data.Teams.PageInfo.HasNextPage && resp.Data.Teams.PageInfo.EndCursor != "",
	}, nil
}

func (c *Client) HasTeam(ctx context.Context, teamID string) (bool, error) {
	const request = `
query TaskSyncLinearTeamByID($id: ID!) {
  teams(first: 1, filter: { id: { eq: $id } }) {
    nodes { id }
  }
}
`

	var resp struct {
		Data struct {
			Teams struct {
				Nodes []struct {
					ID string `json:"id"`
				} `json:"nodes"`
			} `json:"teams"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, request, map[string]any{"id": teamID}, &resp); err != nil {
		return false, err
	}

	return len(resp.Data.Teams.Nodes) > 0, nil
}

func teamSearchFilter(query string) any {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	return map[string]any{
		"or": []any{
			map[string]any{"name": map[string]any{"containsIgnoreCase": query}},
			map[string]any{"key": map[string]any{"containsIgnoreCase": query}},
		},
	}
}

func (c *Client) SearchIssues(ctx context.Context, teamID, query string, first int, after *string) (IssuePage, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return c.listTeamIssues(ctx, teamID, first, after)
	}

	return c.searchTeamIssues(ctx, teamID, query, first, after)
}

func (c *Client) listTeamIssues(ctx context.Context, teamID string, first int, after *string) (IssuePage, error) {
	const request = `
query TaskSyncLinearIssueList($teamId: ID!, $first: Int!, $after: String) {
  issues(
    first: $first
    after: $after
    filter: { team: { id: { eq: $teamId } } }
    orderBy: updatedAt
  ) {
    nodes {
      id
      identifier
      title
      description
      url
      updatedAt
      priority
      dueDate
      state { type }
      assignee { email }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

	var resp struct {
		Data struct {
			Issues struct {
				Nodes    []linearIssue `json:"nodes"`
				PageInfo pageInfo      `json:"pageInfo"`
			} `json:"issues"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, request, map[string]any{
		"teamId": teamID,
		"first":  first,
		"after":  after,
	}, &resp); err != nil {
		return IssuePage{}, err
	}

	return issuePage(resp.Data.Issues.Nodes, resp.Data.Issues.PageInfo)
}

func (c *Client) searchTeamIssues(ctx context.Context, teamID, query string, first int, after *string) (IssuePage, error) {
	const request = `
query TaskSyncLinearIssueSearch($term: String!, $teamId: ID!, $first: Int!, $after: String) {
  searchIssues(
    term: $term
    first: $first
    after: $after
    includeArchived: false
    filter: { team: { id: { eq: $teamId } } }
  ) {
    nodes {
      id
      identifier
      title
      description
      url
      updatedAt
      priority
      dueDate
      state { type }
      assignee { email }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

	var resp struct {
		Data struct {
			SearchIssues struct {
				Nodes    []linearIssue `json:"nodes"`
				PageInfo pageInfo      `json:"pageInfo"`
			} `json:"searchIssues"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, request, map[string]any{
		"term":   query,
		"teamId": teamID,
		"first":  first,
		"after":  after,
	}, &resp); err != nil {
		return IssuePage{}, err
	}

	return issuePage(resp.Data.SearchIssues.Nodes, resp.Data.SearchIssues.PageInfo)
}

func issuePage(nodes []linearIssue, info pageInfo) (IssuePage, error) {
	issues := make([]Issue, 0, len(nodes))
	for _, node := range nodes {
		issue, err := node.toIssue()
		if err != nil {
			return IssuePage{}, err
		}

		issues = append(issues, *issue)
	}

	return IssuePage{
		Issues:      issues,
		EndCursor:   info.EndCursor,
		HasNextPage: info.HasNextPage && info.EndCursor != "",
	}, nil
}

func (c *Client) GetIssue(ctx context.Context, issueID string) (*Issue, error) {
	const request = `
query TaskSyncLinearIssue($id: String!) {
  issue(id: $id) {
    id
    identifier
    title
    description
    url
    updatedAt
    priority
    dueDate
    state { type }
    assignee { email }
    archivedAt
    team {
      id
    }
  }
}
`

	var resp struct {
		Data struct {
			Issue *struct {
				linearIssue
				ArchivedAt *string `json:"archivedAt"`
				Team       struct {
					ID string `json:"id"`
				} `json:"team"`
			} `json:"issue"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, request, map[string]any{"id": issueID}, &resp); err != nil {
		return nil, err
	}

	if resp.Data.Issue == nil || resp.Data.Issue.ID == "" || resp.Data.Issue.ArchivedAt != nil {
		return nil, ErrIssueNotFound
	}

	issue, err := resp.Data.Issue.toIssue()
	if err != nil {
		return nil, err
	}

	issue.TeamID = resp.Data.Issue.Team.ID

	return issue, nil
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
		message := errs[0].Message
		if message == "" {
			message = "graphql error"
		}

		return fmt.Errorf("cannot call Linear graphql: %s", message)
	}

	return nil
}

type linearIssue struct {
	ID          string  `json:"id"`
	Identifier  string  `json:"identifier"`
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    int     `json:"priority"`
	DueDate     *string `json:"dueDate"`
	UpdatedAt   string  `json:"updatedAt"`
	State       *struct {
		Type string `json:"type"`
	} `json:"state"`
	Assignee *struct {
		Email string `json:"email"`
	} `json:"assignee"`
}

func (i linearIssue) toIssue() (*Issue, error) {
	updatedAt, err := time.Parse(time.RFC3339, i.UpdatedAt)
	if err != nil {
		updatedAt, err = time.Parse(time.RFC3339Nano, i.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("cannot parse Linear issue updatedAt: %w", err)
		}
	}

	issue := &Issue{
		ID:          i.ID,
		Identifier:  i.Identifier,
		URL:         i.URL,
		Title:       i.Title,
		Description: i.Description,
		Priority:    i.Priority,
		UpdatedAt:   updatedAt,
	}
	if i.DueDate != nil {
		issue.DueDate = *i.DueDate
	}

	if i.State != nil {
		issue.StateType = i.State.Type
	}

	if i.Assignee != nil {
		issue.AssigneeEmail = i.Assignee.Email
	}

	return issue, nil
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
