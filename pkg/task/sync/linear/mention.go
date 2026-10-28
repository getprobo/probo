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
	"context"
	"net/url"
	"strings"
)

// MentionUser is a Linear workspace member that can be @mentioned.
type MentionUser struct {
	ID          string
	Email       string
	Name        string
	DisplayName string
	URL         string
	Active      bool
	Mentionable bool
}

type mentionUserNode struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
	Active      bool   `json:"active"`
	Mentionable bool   `json:"isMentionable"`
}

func (n mentionUserNode) export() MentionUser {
	return MentionUser{
		ID:          n.ID,
		Email:       strings.TrimSpace(n.Email),
		Name:        strings.TrimSpace(n.Name),
		DisplayName: strings.TrimSpace(n.DisplayName),
		URL:         strings.TrimSpace(n.URL),
		Active:      n.Active,
		Mentionable: n.Mentionable,
	}
}

// MentionUserByEmail returns the only active Linear user with this email.
// Zero or several matches return a nil user.
func (c *Client) MentionUserByEmail(ctx context.Context, email string) (*MentionUser, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, nil
	}

	const query = `
query TaskSyncLinearMentionUserByEmail($email: String!) {
  users(filter: { email: { eqIgnoreCase: $email } }, first: 2) {
    nodes {
      id
      email
      name
      displayName
      url
      active
      isMentionable
    }
  }
}
`

	var resp struct {
		Data struct {
			Users struct {
				Nodes []mentionUserNode `json:"nodes"`
			} `json:"users"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"email": email}, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data.Users.Nodes) != 1 {
		return nil, nil
	}

	user := resp.Data.Users.Nodes[0].export()
	if !user.Active || user.Email == "" {
		return nil, nil
	}

	return &user, nil
}

// MentionUserForProfileURL resolves a Linear profile URL to one workspace user.
// An unknown or ambiguous URL returns a nil user.
func (c *Client) MentionUserForProfileURL(ctx context.Context, rawURL string) (*MentionUser, error) {
	slug, ok := ProfileSlug(rawURL)
	if !ok {
		return nil, nil
	}

	const query = `
query TaskSyncLinearMentionUserByProfile($displayName: String!, $name: String!) {
  byDisplayName: users(filter: { displayName: { eqIgnoreCase: $displayName } }, first: 2) {
    nodes {
      id
      email
      name
      displayName
      url
      active
      isMentionable
    }
  }
  byName: users(filter: { name: { eqIgnoreCase: $name } }, first: 2) {
    nodes {
      id
      email
      name
      displayName
      url
      active
      isMentionable
    }
  }
}
`

	var resp struct {
		Data struct {
			ByDisplayName struct {
				Nodes []mentionUserNode `json:"nodes"`
			} `json:"byDisplayName"`
			ByName struct {
				Nodes []mentionUserNode `json:"nodes"`
			} `json:"byName"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(
		ctx,
		query,
		map[string]any{
			"displayName": slug,
			"name":        slug,
		},
		&resp,
	); err != nil {
		return nil, err
	}

	user := pickMentionUser(rawURL, resp.Data.ByDisplayName.Nodes, resp.Data.ByName.Nodes)
	if user == nil {
		return nil, nil
	}

	exported := user.export()

	return &exported, nil
}

// CommentBody returns the markdown body of one Linear comment.
func (c *Client) CommentBody(ctx context.Context, commentID string) (string, error) {
	const query = `
query TaskSyncLinearCommentBody($id: String!) {
  comment(id: $id) {
    id
    body
  }
}
`

	var resp struct {
		Data struct {
			Comment *struct {
				ID   string `json:"id"`
				Body string `json:"body"`
			} `json:"comment"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	if err := c.do(ctx, query, map[string]any{"id": commentID}, &resp); err != nil {
		return "", err
	}

	if resp.Data.Comment == nil || resp.Data.Comment.ID == "" {
		return "", ErrCommentNotFound
	}

	return resp.Data.Comment.Body, nil
}

// ProfileSlug returns the user slug of a Linear profile URL.
func ProfileSlug(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)

	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil {
		return "", false
	}

	if !strings.EqualFold(parsed.Scheme, "https") && !strings.EqualFold(parsed.Scheme, "http") {
		return "", false
	}

	if !strings.EqualFold(parsed.Hostname(), "linear.app") {
		return "", false
	}

	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) != 3 || !strings.EqualFold(parts[1], "profiles") {
		return "", false
	}

	workspace, err := url.PathUnescape(parts[0])
	if err != nil || !linearWorkspaceKey(workspace) {
		return "", false
	}

	slug, err := url.PathUnescape(parts[2])
	if err != nil || strings.TrimSpace(slug) == "" || strings.Contains(slug, "/") {
		return "", false
	}

	return slug, true
}

func linearWorkspaceKey(value string) bool {
	if value == "" {
		return false
	}

	for i, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case (r == '-' || r == '_') && i > 0:
		default:
			return false
		}
	}

	return true
}

func pickMentionUser(rawURL string, byDisplayName []mentionUserNode, byName []mentionUserNode) *mentionUserNode {
	combined := append(append([]mentionUserNode{}, byDisplayName...), byName...)

	urlMatches := matchingMentionUsers(combined, func(user mentionUserNode) bool {
		return profileSlugsEqual(user.URL, rawURL)
	})
	if len(urlMatches) > 1 {
		return nil
	}

	if len(urlMatches) == 1 {
		return &urlMatches[0]
	}

	displayMatches := matchingMentionUsers(byDisplayName, func(user mentionUserNode) bool {
		slug, ok := ProfileSlug(rawURL)

		return ok && strings.EqualFold(user.DisplayName, slug)
	})
	if len(displayMatches) == 1 {
		return &displayMatches[0]
	}

	if len(displayMatches) > 1 {
		return nil
	}

	nameMatches := matchingMentionUsers(byName, func(user mentionUserNode) bool {
		slug, ok := ProfileSlug(rawURL)

		return ok && strings.EqualFold(user.Name, slug)
	})
	if len(nameMatches) == 1 {
		return &nameMatches[0]
	}

	return nil
}

func matchingMentionUsers(
	users []mentionUserNode,
	match func(mentionUserNode) bool,
) []mentionUserNode {
	var matched []mentionUserNode

	seen := map[string]struct{}{}

	for _, user := range users {
		if !user.Active || strings.TrimSpace(user.Email) == "" || user.ID == "" {
			continue
		}

		if !match(user) {
			continue
		}

		if _, ok := seen[user.ID]; ok {
			continue
		}

		seen[user.ID] = struct{}{}
		matched = append(matched, user)
	}

	return matched
}

func profileSlugsEqual(leftURL string, rightURL string) bool {
	left, leftOK := ProfileSlug(leftURL)
	right, rightOK := ProfileSlug(rightURL)

	return leftOK && rightOK && strings.EqualFold(left, right)
}
