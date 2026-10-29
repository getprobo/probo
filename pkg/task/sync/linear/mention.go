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
	"fmt"
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
		if linearEntityMissing(err) {
			return "", ErrCommentNotFound
		}

		return "", fmt.Errorf("cannot load Linear comment: %w", err)
	}

	if resp.Data.Comment == nil || resp.Data.Comment.ID == "" {
		return "", ErrCommentNotFound
	}

	return resp.Data.Comment.Body, nil
}

type profileRef struct {
	workspace string
	slug      string
}

// ProfileSlug returns the user slug of a Linear profile URL.
func ProfileSlug(raw string) (string, bool) {
	ref, ok := profileRefFromURL(raw)
	if !ok {
		return "", false
	}

	return ref.slug, true
}

func profileRefFromURL(raw string) (profileRef, bool) {
	raw = strings.TrimSpace(raw)

	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil {
		return profileRef{}, false
	}

	if !strings.EqualFold(parsed.Scheme, "https") && !strings.EqualFold(parsed.Scheme, "http") {
		return profileRef{}, false
	}

	if !strings.EqualFold(parsed.Hostname(), "linear.app") {
		return profileRef{}, false
	}

	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) != 3 || !strings.EqualFold(parts[1], "profiles") {
		return profileRef{}, false
	}

	workspace, err := url.PathUnescape(parts[0])
	if err != nil || !linearWorkspaceKey(workspace) {
		return profileRef{}, false
	}

	slug, err := url.PathUnescape(parts[2])
	if err != nil || strings.TrimSpace(slug) == "" || strings.Contains(slug, "/") {
		return profileRef{}, false
	}

	return profileRef{workspace: workspace, slug: slug}, true
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

	user, ambiguous := matchUniqueMentionUser(combined, func(user mentionUserNode) bool {
		return profileRefsEqual(user.URL, rawURL)
	})
	if ambiguous {
		return nil
	}

	if user != nil {
		return user
	}

	slug, ok := ProfileSlug(rawURL)
	if !ok {
		return nil
	}

	user, ambiguous = matchUniqueMentionUser(byDisplayName, func(user mentionUserNode) bool {
		return profileInWorkspace(user.URL, rawURL) && strings.EqualFold(user.DisplayName, slug)
	})
	if ambiguous {
		return nil
	}

	if user != nil {
		return user
	}

	user, _ = matchUniqueMentionUser(byName, func(user mentionUserNode) bool {
		return profileInWorkspace(user.URL, rawURL) && strings.EqualFold(user.Name, slug)
	})

	return user
}

// matchUniqueMentionUser finds the single user in users matching match,
// identified before the active/email check so a deactivated or email-less
// namesake still makes the match ambiguous rather than letting a different
// active user stand in for them. The second result is true only when two or
// more distinct users match; the caller must then treat the lookup as
// unresolved rather than fall back to a looser query.
func matchUniqueMentionUser(
	users []mentionUserNode,
	match func(mentionUserNode) bool,
) (*mentionUserNode, bool) {
	seen := map[string]struct{}{}

	var unique *mentionUserNode

	for i := range users {
		user := users[i]
		if user.ID == "" || !match(user) {
			continue
		}

		if _, ok := seen[user.ID]; ok {
			continue
		}

		seen[user.ID] = struct{}{}

		if unique != nil {
			return nil, true
		}

		unique = &user
	}

	if unique == nil || !unique.Active || strings.TrimSpace(unique.Email) == "" {
		return nil, false
	}

	return unique, false
}

func profileRefsEqual(leftURL string, rightURL string) bool {
	left, leftOK := profileRefFromURL(leftURL)
	right, rightOK := profileRefFromURL(rightURL)

	return leftOK && rightOK &&
		strings.EqualFold(left.workspace, right.workspace) &&
		strings.EqualFold(left.slug, right.slug)
}

func profileInWorkspace(userURL string, rawURL string) bool {
	user, userOK := profileRefFromURL(userURL)
	raw, rawOK := profileRefFromURL(rawURL)

	return userOK && rawOK && strings.EqualFold(user.workspace, raw.workspace)
}
