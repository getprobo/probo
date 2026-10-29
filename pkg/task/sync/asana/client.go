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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
)

const (
	pageSize        = 50
	maxResponseSize = 8 << 20
	maxFileSize     = 32 << 20
	requestTimeout  = 30 * time.Second
	taskOptFields   = "gid,name,html_notes,notes,completed,due_on,due_at,permalink_url,modified_at,assignee.gid,assignee.email,memberships.project.gid,memberships.section.gid,custom_fields.gid,custom_fields.enum_value.gid"
	listMaxPages    = 20
)

var numericGID = regexp.MustCompile(`^[0-9]+$`)

type (
	Client struct {
		httpClient          *http.Client
		baseURL             string
		downloadHTTP        *http.Client
		allowedDownloadHost string
	}

	APIError struct {
		Status  int
		Message string
	}

	Named struct {
		GID  string
		Name string
	}

	User struct {
		GID   string
		Email string
		Name  string
	}

	Page struct {
		Items       []Named
		EndCursor   string
		HasNextPage bool
	}

	EnumOption struct {
		GID     string
		Name    string
		Enabled bool
	}

	EnumField struct {
		GID     string
		Name    string
		Options []EnumOption
	}

	Section struct {
		GID  string
		Name string
	}

	ProjectSetup struct {
		Project      Named
		TeamGID      string
		WorkspaceGID string
		Sections     []Section
		EnumFields   []EnumField
	}

	Membership struct {
		ProjectGID string
		SectionGID string
	}

	CustomFieldValue struct {
		GID           string
		EnumOptionGID string
	}

	Task struct {
		GID           string
		Name          string
		HTMLNotes     string
		Notes         string
		Completed     bool
		DueOn         string
		DueAt         string
		AssigneeGID   string
		AssigneeEmail string
		Permalink     string
		ModifiedAt    time.Time
		Memberships   []Membership
		CustomFields  []CustomFieldValue
	}

	TaskPage struct {
		Tasks       []Task
		EndCursor   string
		HasNextPage bool
	}

	Story struct {
		GID         string
		HTMLText    string
		Text        string
		CreatedAt   time.Time
		AuthorEmail string
	}

	Attachment struct {
		GID         string
		Name        string
		DownloadURL string
		ParentGID   string
		ParentType  string
	}

	TaskCreate struct {
		Name         string
		HTMLNotes    string
		ProjectGID   string
		Completed    bool
		DueOn        *string
		AssigneeGID  *string
		CustomFields map[string]string
	}

	TaskUpdate struct {
		Name         *string
		HTMLNotes    *string
		HTMLNotesSet bool
		Completed    *bool
		DueOn        *string
		DueOnSet     bool
		AssigneeGID  *string
		AssigneeSet  bool
		CustomFields map[string]*string
	}

	dataEnvelope struct {
		Data     json.RawMessage `json:"data"`
		NextPage *struct {
			Offset string `json:"offset"`
		} `json:"next_page"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
)

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("asana api returned status %d", e.Status)
	}

	return fmt.Sprintf("asana api returned status %d: %s", e.Status, e.Message)
}

func (e *APIError) NotFound() bool {
	return e != nil && e.Status == http.StatusNotFound
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	downloadHTTP := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	downloadHTTP.Timeout = requestTimeout

	return &Client{
		httpClient:   httpClient,
		baseURL:      strings.TrimRight(baseURL, "/"),
		downloadHTTP: downloadHTTP,
	}
}

func (c *Client) ListTeams(ctx context.Context, workspaceGID, query string, limit int, offset string) (Page, error) {
	if strings.TrimSpace(query) != "" {
		return c.typeahead(ctx, workspaceGID, "team", query, limit)
	}

	return c.namedPage(
		ctx,
		[]string{"workspaces", workspaceGID, "teams"},
		url.Values{
			"opt_fields": {"gid,name"},
			"limit":      {strconv.Itoa(boundedLimit(limit))},
			"offset":     {offset},
		},
	)
}

func (c *Client) ListProjects(
	ctx context.Context,
	workspaceGID string,
	teamGID string,
	query string,
	limit int,
	offset string,
) (Page, error) {
	if strings.TrimSpace(query) != "" {
		return c.typeahead(ctx, workspaceGID, "project", query, limit)
	}

	segments := []string{"workspaces", workspaceGID, "projects"}
	if teamGID != "" {
		segments = []string{"teams", teamGID, "projects"}
	}

	values := url.Values{
		"opt_fields": {"gid,name,archived"},
		"limit":      {strconv.Itoa(boundedLimit(limit))},
		"archived":   {"false"},
		"offset":     {offset},
	}

	page, err := c.namedPage(ctx, segments, values)
	if err != nil {
		return Page{}, err
	}

	return page, nil
}

func (c *Client) ProjectSetup(ctx context.Context, projectGID string) (ProjectSetup, error) {
	var project struct {
		GID      string `json:"gid"`
		Name     string `json:"name"`
		Archived bool   `json:"archived"`
		Team     *struct {
			GID string `json:"gid"`
		} `json:"team"`
		Workspace *struct {
			GID string `json:"gid"`
		} `json:"workspace"`
	}

	if err := c.get(
		ctx,
		[]string{"projects", projectGID},
		url.Values{"opt_fields": {"gid,name,archived,team.gid,workspace.gid"}},
		&project,
	); err != nil {
		return ProjectSetup{}, err
	}

	sections, err := c.listSections(ctx, projectGID)
	if err != nil {
		return ProjectSetup{}, err
	}

	fields, err := c.listEnumFields(ctx, projectGID)
	if err != nil {
		return ProjectSetup{}, err
	}

	setup := ProjectSetup{
		Project:    Named{GID: project.GID, Name: project.Name},
		Sections:   sections,
		EnumFields: fields,
	}
	if project.Team != nil {
		setup.TeamGID = project.Team.GID
	}

	if project.Workspace != nil {
		setup.WorkspaceGID = project.Workspace.GID
	}

	return setup, nil
}

func (c *Client) ListTasks(
	ctx context.Context,
	workspaceGID string,
	projectGID string,
	query string,
	limit int,
	offset string,
) (TaskPage, error) {
	limit = boundedLimit(limit)

	if strings.TrimSpace(query) != "" {
		values := url.Values{
			"text":           {query},
			"projects.any":   {projectGID},
			"opt_fields":     {"gid,name,completed,permalink_url"},
			"limit":          {strconv.Itoa(limit)},
			"sort_by":        {"modified_at"},
			"sort_ascending": {"false"},
		}
		if offset != "" {
			values.Set("offset", offset)
		}

		return c.taskPage(ctx, []string{"workspaces", workspaceGID, "tasks", "search"}, values)
	}

	return c.taskPage(
		ctx,
		[]string{"projects", projectGID, "tasks"},
		url.Values{
			"opt_fields": {"gid,name,completed,permalink_url"},
			"limit":      {strconv.Itoa(limit)},
			"offset":     {offset},
		},
	)
}

func (c *Client) GetTask(ctx context.Context, taskGID string) (Task, error) {
	var raw taskPayload

	err := c.get(
		ctx,
		[]string{"tasks", taskGID},
		url.Values{"opt_fields": {taskOptFields}},
		&raw,
	)
	if err != nil {
		return Task{}, err
	}

	return raw.task(), nil
}

func (c *Client) CreateTask(ctx context.Context, input TaskCreate) (Task, error) {
	body := map[string]any{
		"name":      input.Name,
		"projects":  []string{input.ProjectGID},
		"completed": input.Completed,
	}
	if input.HTMLNotes != "" {
		body["html_notes"] = input.HTMLNotes
	}

	if input.DueOn != nil {
		body["due_on"] = *input.DueOn
	}

	if input.AssigneeGID != nil {
		body["assignee"] = *input.AssigneeGID
	}

	if len(input.CustomFields) > 0 {
		body["custom_fields"] = input.CustomFields
	}

	var raw taskPayload
	if err := c.send(ctx, http.MethodPost, []string{"tasks"}, nil, body, &raw); err != nil {
		return Task{}, err
	}

	return raw.task(), nil
}

func (c *Client) UpdateTask(ctx context.Context, taskGID string, input TaskUpdate) (Task, error) {
	body := map[string]any{}
	if input.Name != nil {
		body["name"] = *input.Name
	}

	if input.HTMLNotesSet {
		if input.HTMLNotes == nil || *input.HTMLNotes == "" {
			body["html_notes"] = nil
		} else {
			body["html_notes"] = *input.HTMLNotes
		}
	}

	if input.Completed != nil {
		body["completed"] = *input.Completed
	}

	if input.DueOnSet {
		if input.DueOn == nil || *input.DueOn == "" {
			body["due_on"] = nil
		} else {
			body["due_on"] = *input.DueOn
		}
	}

	if input.AssigneeSet {
		if input.AssigneeGID == nil || *input.AssigneeGID == "" {
			body["assignee"] = nil
		} else {
			body["assignee"] = *input.AssigneeGID
		}
	}

	if len(input.CustomFields) > 0 {
		fields := map[string]any{}

		for gid, option := range input.CustomFields {
			if option == nil || *option == "" {
				fields[gid] = nil
			} else {
				fields[gid] = *option
			}
		}

		body["custom_fields"] = fields
	}

	var raw taskPayload
	if err := c.send(ctx, http.MethodPut, []string{"tasks", taskGID}, nil, body, &raw); err != nil {
		return Task{}, err
	}

	return raw.task(), nil
}

func (c *Client) AddTaskToSection(ctx context.Context, sectionGID, taskGID string) error {
	body := map[string]any{"task": taskGID}

	return c.send(ctx, http.MethodPost, []string{"sections", sectionGID, "addTask"}, nil, body, nil)
}

func (c *Client) UserGIDByEmail(ctx context.Context, workspaceGID, email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", nil
	}

	values := url.Values{
		"resource_type": {"user"},
		"query":         {email},
		"count":         {"10"},
		"opt_fields":    {"gid,email"},
	}

	var users []struct {
		GID   string `json:"gid"`
		Email string `json:"email"`
	}

	if err := c.get(ctx, []string{"workspaces", workspaceGID, "typeahead"}, values, &users); err != nil {
		return "", err
	}

	for _, user := range users {
		if strings.EqualFold(strings.TrimSpace(user.Email), email) {
			return user.GID, nil
		}
	}

	return "", nil
}

func (c *Client) UserByGID(ctx context.Context, userGID string) (User, error) {
	if !numericGID.MatchString(userGID) {
		return User{}, nil
	}

	var user struct {
		GID   string `json:"gid"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	values := url.Values{"opt_fields": {"gid,email,name"}}
	if err := c.get(ctx, []string{"users", userGID}, values, &user); err != nil {
		if IsNotFound(err) {
			return User{}, nil
		}

		return User{}, err
	}

	return User{GID: user.GID, Email: user.Email, Name: user.Name}, nil
}

func (c *Client) ListStories(ctx context.Context, taskGID string) ([]Story, error) {
	raw, err := collectStories(ctx, c, taskGID)
	if err != nil {
		return nil, err
	}

	stories := make([]Story, 0, len(raw))
	for _, story := range raw {
		if story.ResourceSubtype != "comment_added" {
			continue
		}

		stories = append(stories, story.story())
	}

	return stories, nil
}

func (c *Client) CreateStory(ctx context.Context, taskGID, htmlText string) (Story, error) {
	var raw storyPayload

	err := c.send(
		ctx,
		http.MethodPost,
		[]string{"tasks", taskGID, "stories"},
		nil,
		map[string]any{"html_text": htmlText},
		&raw,
	)
	if err != nil {
		return Story{}, err
	}

	return raw.story(), nil
}

func (c *Client) UpdateStory(ctx context.Context, storyGID, htmlText string) (Story, error) {
	var raw storyPayload

	err := c.send(
		ctx,
		http.MethodPut,
		[]string{"stories", storyGID},
		nil,
		map[string]any{"html_text": htmlText},
		&raw,
	)
	if err != nil {
		return Story{}, err
	}

	return raw.story(), nil
}

func (c *Client) DeleteStory(ctx context.Context, storyGID string) error {
	return c.send(ctx, http.MethodDelete, []string{"stories", storyGID}, nil, nil, nil)
}

func (c *Client) ListAttachments(ctx context.Context, taskGID string) ([]Attachment, error) {
	return c.listAttachments(ctx, []string{"tasks", taskGID, "attachments"})
}

func (c *Client) ListStoryAttachments(ctx context.Context, storyGID string) ([]Attachment, error) {
	return c.listAttachments(ctx, []string{"stories", storyGID, "attachments"})
}

func (c *Client) listAttachments(ctx context.Context, segments []string) ([]Attachment, error) {
	raw, err := collectPages[attachmentPayload](
		ctx,
		c,
		segments,
		url.Values{"opt_fields": {"gid,name,download_url,parent.gid,parent.resource_type"}},
	)
	if err != nil {
		return nil, err
	}

	attachments := make([]Attachment, 0, len(raw))
	for _, attachment := range raw {
		attachments = append(attachments, attachment.attachment())
	}

	return attachments, nil
}

func (c *Client) UploadAttachment(ctx context.Context, parentGID, name string, body []byte) (Attachment, error) {
	var buffer bytes.Buffer

	writer := multipart.NewWriter(&buffer)
	if err := writer.WriteField("parent", parentGID); err != nil {
		return Attachment{}, fmt.Errorf("cannot write attachment parent: %w", err)
	}

	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return Attachment{}, fmt.Errorf("cannot create attachment part: %w", err)
	}

	if _, err := part.Write(body); err != nil {
		return Attachment{}, fmt.Errorf("cannot write attachment body: %w", err)
	}

	if err := writer.Close(); err != nil {
		return Attachment{}, fmt.Errorf("cannot close attachment body: %w", err)
	}

	endpoint, err := c.join("attachments")
	if err != nil {
		return Attachment{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buffer)
	if err != nil {
		return Attachment{}, fmt.Errorf("cannot build attachment request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	var raw attachmentPayload
	if _, err := c.do(req, &raw); err != nil {
		return Attachment{}, err
	}

	return raw.attachment(), nil
}

func (c *Client) DeleteAttachment(ctx context.Context, attachmentGID string) error {
	return c.send(ctx, http.MethodDelete, []string{"attachments", attachmentGID}, nil, nil, nil)
}

func (c *Client) CreateWebhook(ctx context.Context, projectGID, target string) (string, error) {
	body := map[string]any{
		"resource": projectGID,
		"target":   target,
		"filters": []map[string]string{
			{"resource_type": "task", "action": "added"},
			{"resource_type": "task", "action": "changed"},
			{"resource_type": "task", "action": "deleted"},
			{"resource_type": "task", "action": "removed"},
			{"resource_type": "story", "action": "added"},
			{"resource_type": "story", "action": "changed"},
			{"resource_type": "story", "action": "removed"},
			{"resource_type": "attachment", "action": "added"},
			{"resource_type": "attachment", "action": "deleted"},
		},
	}

	var created struct {
		GID string `json:"gid"`
	}

	if err := c.send(ctx, http.MethodPost, []string{"webhooks"}, nil, body, &created); err != nil {
		return "", err
	}

	if created.GID == "" {
		return "", fmt.Errorf("cannot create asana webhook: response has no gid")
	}

	return created.GID, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, webhookGID string) error {
	return c.send(ctx, http.MethodDelete, []string{"webhooks", webhookGID}, nil, nil, nil)
}

func (c *Client) DeleteTask(ctx context.Context, taskGID string) error {
	return c.send(ctx, http.MethodDelete, []string{"tasks", taskGID}, nil, nil, nil)
}

func (c *Client) listSections(ctx context.Context, projectGID string) ([]Section, error) {
	var raw []struct {
		GID  string `json:"gid"`
		Name string `json:"name"`
	}

	err := c.get(
		ctx,
		[]string{"projects", projectGID, "sections"},
		url.Values{"opt_fields": {"gid,name"}, "limit": {"100"}},
		&raw,
	)
	if err != nil {
		return nil, err
	}

	sections := make([]Section, 0, len(raw))
	for _, section := range raw {
		sections = append(sections, Section{GID: section.GID, Name: section.Name})
	}

	return sections, nil
}

func (c *Client) listEnumFields(ctx context.Context, projectGID string) ([]EnumField, error) {
	var raw []struct {
		CustomField struct {
			GID             string `json:"gid"`
			Name            string `json:"name"`
			ResourceSubtype string `json:"resource_subtype"`
			EnumOptions     []struct {
				GID     string `json:"gid"`
				Name    string `json:"name"`
				Enabled bool   `json:"enabled"`
			} `json:"enum_options"`
		} `json:"custom_field"`
	}

	err := c.get(
		ctx,
		[]string{"projects", projectGID, "custom_field_settings"},
		url.Values{
			"opt_fields": {"custom_field.gid,custom_field.name,custom_field.resource_subtype,custom_field.enum_options.gid,custom_field.enum_options.name,custom_field.enum_options.enabled"},
			"limit":      {"100"},
		},
		&raw,
	)
	if err != nil {
		return nil, err
	}

	fields := make([]EnumField, 0, len(raw))
	for _, row := range raw {
		if row.CustomField.ResourceSubtype != "enum" {
			continue
		}

		options := make([]EnumOption, 0, len(row.CustomField.EnumOptions))
		for _, option := range row.CustomField.EnumOptions {
			if !option.Enabled {
				continue
			}

			options = append(options, EnumOption{GID: option.GID, Name: option.Name, Enabled: option.Enabled})
		}

		fields = append(fields, EnumField{
			GID:     row.CustomField.GID,
			Name:    row.CustomField.Name,
			Options: options,
		})
	}

	return fields, nil
}

func (c *Client) typeahead(ctx context.Context, workspaceGID, resourceType, query string, limit int) (Page, error) {
	var raw []struct {
		GID  string `json:"gid"`
		Name string `json:"name"`
	}

	err := c.get(
		ctx,
		[]string{"workspaces", workspaceGID, "typeahead"},
		url.Values{
			"resource_type": {resourceType},
			"query":         {query},
			"count":         {strconv.Itoa(boundedLimit(limit))},
			"opt_fields":    {"gid,name"},
		},
		&raw,
	)
	if err != nil {
		return Page{}, err
	}

	items := make([]Named, 0, len(raw))
	for _, item := range raw {
		items = append(items, Named{GID: item.GID, Name: item.Name})
	}

	return Page{Items: items}, nil
}

func (c *Client) namedPage(ctx context.Context, segments []string, values url.Values) (Page, error) {
	var raw []struct {
		GID      string `json:"gid"`
		Name     string `json:"name"`
		Archived bool   `json:"archived"`
	}

	next, err := c.getPage(ctx, segments, values, &raw)
	if err != nil {
		return Page{}, err
	}

	items := make([]Named, 0, len(raw))
	for _, item := range raw {
		if item.Archived {
			continue
		}

		items = append(items, Named{GID: item.GID, Name: item.Name})
	}

	return Page{Items: items, EndCursor: next, HasNextPage: next != ""}, nil
}

func (c *Client) taskPage(ctx context.Context, segments []string, values url.Values) (TaskPage, error) {
	var raw []taskPayload

	next, err := c.getPage(ctx, segments, values, &raw)
	if err != nil {
		return TaskPage{}, err
	}

	tasks := make([]Task, 0, len(raw))
	for _, item := range raw {
		tasks = append(tasks, item.task())
	}

	return TaskPage{Tasks: tasks, EndCursor: next, HasNextPage: next != ""}, nil
}

func (c *Client) get(ctx context.Context, segments []string, values url.Values, out any) error {
	_, err := c.getPage(ctx, segments, values, out)

	return err
}

func (c *Client) getPage(ctx context.Context, segments []string, values url.Values, out any) (string, error) {
	if values != nil {
		if offset := values.Get("offset"); offset == "" {
			values.Del("offset")
		}
	}

	return c.sendPage(ctx, http.MethodGet, segments, values, nil, out)
}

func (c *Client) send(ctx context.Context, method string, segments []string, values url.Values, body any, out any) error {
	_, err := c.sendPage(ctx, method, segments, values, body, out)

	return err
}

func (c *Client) sendPage(
	ctx context.Context,
	method string,
	segments []string,
	values url.Values,
	body any,
	out any,
) (string, error) {
	endpoint, err := c.join(segments...)
	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse asana url: %w", err)
	}

	if len(values) > 0 {
		parsed.RawQuery = values.Encode()
	}

	var reader io.Reader

	if body != nil {
		encoded, err := json.Marshal(map[string]any{"data": body})
		if err != nil {
			return "", fmt.Errorf("cannot marshal asana request: %w", err)
		}

		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, parsed.String(), reader)
	if err != nil {
		return "", fmt.Errorf("cannot build asana request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) (string, error) {
	if c == nil || c.httpClient == nil {
		return "", fmt.Errorf("cannot call asana: client is not configured")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot call asana: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "", fmt.Errorf("cannot read asana response: %w", err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return "", nil
	}

	var envelope dataEnvelope
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &envelope); err != nil {
			if resp.StatusCode >= 300 {
				return "", &APIError{Status: resp.StatusCode}
			}

			return "", fmt.Errorf("cannot decode asana response: %w", err)
		}
	}

	if resp.StatusCode >= 300 {
		message := ""
		if len(envelope.Errors) > 0 {
			message = envelope.Errors[0].Message
		}

		return "", &APIError{Status: resp.StatusCode, Message: message}
	}

	if out != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return "", fmt.Errorf("cannot decode asana data: %w", err)
		}
	}

	if envelope.NextPage == nil {
		return "", nil
	}

	return envelope.NextPage.Offset, nil
}

func (c *Client) join(segments ...string) (string, error) {
	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			return "", fmt.Errorf("cannot build asana url: empty path segment")
		}

		escaped = append(escaped, url.PathEscape(segment))
	}

	joined, err := url.JoinPath(c.baseURL, escaped...)
	if err != nil {
		return "", fmt.Errorf("cannot build asana url: %w", err)
	}

	return joined, nil
}

func boundedLimit(limit int) int {
	if limit <= 0 || limit > pageSize {
		return pageSize
	}

	return limit
}

func IsNotFound(err error) bool {
	apiErr, ok := errors.AsType[*APIError](err)

	return ok && apiErr.NotFound()
}
