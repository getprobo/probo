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

package tasksync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/prosemirror"
)

var (
	asanaNumericGID  = regexp.MustCompile(`^[0-9]+$`)
	asanaUserAnchor  = regexp.MustCompile(`(?is)<a\b[^>]*\bdata-asana-gid="([0-9]+)"[^>]*>.*?</a>|<a\b[^>]*\bdata-asana-gid="([0-9]+)"[^>]*/>`)
	asanaMentionSpan = regexp.MustCompile(`<span class="mention" data-type="mention" data-id="([^"]*)" data-label="[^"]*">([^<]*)</span>`)
)

func (s *Service) asanaHTMLNotes(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	content string,
) (string, error) {
	gids, err := s.asanaMentionGIDs(ctx, scope, account, organizationID, content)
	if err != nil {
		return "", err
	}

	origin, err := asanaOrigin(s.asanaAPIBaseURL)
	if err != nil {
		return "", err
	}

	return contentToAsanaHTMLWithMentions(content, gids, origin)
}

func (s *Service) asanaCommentHTML(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	authorName string,
	content string,
) (string, error) {
	body, err := s.asanaHTMLNotes(ctx, scope, account, organizationID, content)
	if err != nil {
		return "", err
	}

	return prefixAsanaAuthor(authorName, body), nil
}

func (s *Service) contentFromAsanaHTML(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	htmlText string,
) (string, error) {
	origin, err := asanaOrigin(s.asanaAPIBaseURL)
	if err != nil {
		return "", err
	}

	content, err := asanaHTMLToContent(asanaMentionAnchorsToURLs(htmlText, origin))
	if err != nil {
		return "", err
	}

	return s.attachAsanaMentions(ctx, scope, account, organizationID, content, origin)
}

func (s *Service) asanaMentionGIDs(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	content string,
) (map[string]string, error) {
	ids, err := prosemirror.MentionedProfileIDs(content)
	if err != nil {
		return nil, fmt.Errorf("cannot collect mentions: %w", err)
	}

	if len(ids) == 0 || account.settings.WorkspaceGID == "" {
		return nil, nil
	}

	gids := map[string]string{}

	for _, id := range ids {
		emails, err := s.mentionEmails(ctx, scope, organizationID, id)
		if err != nil {
			return nil, err
		}

		for _, email := range emails {
			userGID, err := account.client.UserGIDByEmail(ctx, account.settings.WorkspaceGID, email)
			if err != nil {
				return nil, fmt.Errorf("cannot look up Asana user: %w", err)
			}

			if userGID == "" || !asanaNumericGID.MatchString(userGID) {
				continue
			}

			gids[id.String()] = userGID

			break
		}
	}

	return gids, nil
}

func (s *Service) mentionEmails(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	profileID gid.GID,
) ([]string, error) {
	var emails []string

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, conn, scope, profileID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load mentioned profile: %w", err)
			}

			if profile.OrganizationID != organizationID || profile.State == coredata.ProfileStateDeactivated {
				return nil
			}

			if profile.EmailAddress != mail.Nil {
				emails = append(emails, profile.EmailAddress.String())
			}

			for _, extra := range profile.AdditionalEmailAddresses {
				if extra != mail.Nil {
					emails = append(emails, extra.String())
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load mention emails: %w", err)
	}

	return emails, nil
}

func (s *Service) attachAsanaMentions(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	content string,
	origin string,
) (string, error) {
	if strings.TrimSpace(content) == "" {
		return content, nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse Asana content: %w", err)
	}

	var lookupErr error

	cache := map[string]struct {
		attrs prosemirror.MentionAttrs
		ok    bool
	}{}

	rewritten, changed := rewriteAsanaMentionNodes(node, origin, func(userGID string) (prosemirror.MentionAttrs, bool) {
		if lookupErr != nil {
			return prosemirror.MentionAttrs{}, false
		}

		if cached, ok := cache[userGID]; ok {
			return cached.attrs, cached.ok
		}

		attrs, ok, err := s.lookupAsanaMention(ctx, scope, account, organizationID, userGID)
		if err != nil {
			lookupErr = err

			return prosemirror.MentionAttrs{}, false
		}

		cache[userGID] = struct {
			attrs prosemirror.MentionAttrs
			ok    bool
		}{attrs: attrs, ok: ok}

		return attrs, ok
	})

	if lookupErr != nil {
		return "", lookupErr
	}

	if !changed {
		return content, nil
	}

	encoded, err := json.Marshal(rewritten)
	if err != nil {
		return "", fmt.Errorf("cannot marshal Asana mentions: %w", err)
	}

	encodedJSON := string(encoded)

	sanitized, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize Asana mentions: %w", err)
	}

	return sanitized, nil
}

func (s *Service) lookupAsanaMention(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	userGID string,
) (prosemirror.MentionAttrs, bool, error) {
	user, err := account.client.UserByGID(ctx, userGID)
	if err != nil {
		return prosemirror.MentionAttrs{}, false, fmt.Errorf("cannot load Asana user: %w", err)
	}

	addr, err := mail.ParseAddr(strings.TrimSpace(user.Email))
	if err != nil {
		return prosemirror.MentionAttrs{}, false, nil
	}

	var (
		attrs prosemirror.MentionAttrs
		ok    bool
	)

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByOrganizationIDAndEmail(ctx, conn, scope, organizationID, addr); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load mentioned profile: %w", err)
			}

			if profile.State == coredata.ProfileStateDeactivated {
				return nil
			}

			label := truncateMentionLabel(profile.FullName)
			if label == "" {
				label = truncateMentionLabel(user.Name)
			}

			if label == "" {
				label = truncateMentionLabel(addr.String())
			}

			if label == "" {
				return nil
			}

			attrs = prosemirror.MentionAttrs{ID: profile.ID.String(), Label: label}
			ok = true

			return nil
		},
	)
	if err != nil {
		return prosemirror.MentionAttrs{}, false, fmt.Errorf("cannot match Asana profile: %w", err)
	}

	return attrs, ok, nil
}

func contentToAsanaHTMLWithMentions(content string, gids map[string]string, origin string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse task content: %w", err)
	}

	node.Content = withoutAttachments(node.Content)
	if len(node.Content) == 0 {
		return "", nil
	}

	rendered, err := prosemirror.RenderHTML(node)
	if err != nil {
		return "", fmt.Errorf("cannot render task html: %w", err)
	}

	rendered = strings.TrimSpace(applyAsanaMentionAnchors(rendered, gids, origin))
	if rendered == "" {
		return "", nil
	}

	return "<body>" + rendered + "</body>", nil
}

func applyAsanaMentionAnchors(rendered string, gids map[string]string, origin string) string {
	return asanaMentionSpan.ReplaceAllStringFunc(rendered, func(match string) string {
		parts := asanaMentionSpan.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		inner := parts[2]

		userGID := ""
		if gids != nil {
			userGID = gids[parts[1]]
		}

		if !asanaNumericGID.MatchString(userGID) {
			return inner
		}

		profileURL, err := asanaProfileURL(origin, userGID)
		if err != nil {
			return inner
		}

		return `<a data-asana-gid="` + userGID + `" href="` + profileURL + `">` + inner + `</a>`
	})
}

func asanaMentionAnchorsToURLs(htmlText, origin string) string {
	return asanaUserAnchor.ReplaceAllStringFunc(htmlText, func(match string) string {
		parts := asanaUserAnchor.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		userGID := parts[1]
		if userGID == "" {
			userGID = parts[2]
		}

		profileURL, err := asanaProfileURL(origin, userGID)
		if err != nil {
			return match
		}

		return profileURL
	})
}

func asanaProfileURL(origin, userGID string) (string, error) {
	if !asanaNumericGID.MatchString(userGID) {
		return "", fmt.Errorf("cannot build Asana profile url: invalid user")
	}

	profileURL, err := url.JoinPath(origin, "0", "profile", url.PathEscape(userGID))
	if err != nil {
		return "", fmt.Errorf("cannot build Asana profile url: %w", err)
	}

	return profileURL, nil
}

func prefixAsanaAuthor(authorName, body string) string {
	authorName = strings.TrimSpace(authorName)
	if authorName == "" {
		return body
	}

	inner := strings.TrimSuffix(strings.TrimPrefix(body, "<body>"), "</body>")
	if strings.TrimSpace(inner) == "" {
		inner = ""
	} else {
		inner = "<br/>" + inner
	}

	return "<body><strong>" + htmlEscape(authorName) + "</strong>" + inner + "</body>"
}

func rewriteAsanaMentionNodes(
	node prosemirror.Node,
	origin string,
	resolve func(string) (prosemirror.MentionAttrs, bool),
) (prosemirror.Node, bool) {
	if node.Type == prosemirror.NodeCodeBlock {
		return node, false
	}

	content, changed := rewriteAsanaMentionChildren(node.Content, origin, resolve)
	node.Content = content

	return node, changed
}

func rewriteAsanaMentionChildren(
	nodes []prosemirror.Node,
	origin string,
	resolve func(string) (prosemirror.MentionAttrs, bool),
) ([]prosemirror.Node, bool) {
	var (
		out     []prosemirror.Node
		changed bool
	)

	for _, node := range nodes {
		if userGID, ok := asanaProfileGID(origin, asanaLinkHref(node)); ok {
			if attrs, ok := resolve(userGID); ok {
				mention, err := newMentionNode(attrs)
				if err == nil {
					out = append(out, mention)
					changed = true

					continue
				}
			}
		}

		if node.Type == prosemirror.NodeText && !nodeHasMark(node, prosemirror.MarkCode) && !nodeHasMark(node, prosemirror.MarkLink) {
			split, replaced := splitAsanaProfileText(node, origin, resolve)
			if replaced {
				changed = true
			}

			out = append(out, split...)

			continue
		}

		rewritten, childChanged := rewriteAsanaMentionNodes(node, origin, resolve)
		if childChanged {
			changed = true
		}

		out = append(out, rewritten)
	}

	return out, changed
}

func splitAsanaProfileText(
	node prosemirror.Node,
	origin string,
	resolve func(string) (prosemirror.MentionAttrs, bool),
) ([]prosemirror.Node, bool) {
	if node.Text == nil {
		return []prosemirror.Node{node}, false
	}

	text := *node.Text
	prefix := strings.TrimRight(origin, "/") + "/0/profile/"

	var (
		out     []prosemirror.Node
		cursor  int
		changed bool
	)

	for cursor < len(text) {
		start := strings.Index(text[cursor:], prefix)
		if start < 0 {
			break
		}

		start += cursor

		gidStart := start + len(prefix)
		gidEnd := gidStart

		for gidEnd < len(text) && text[gidEnd] >= '0' && text[gidEnd] <= '9' {
			gidEnd++
		}

		if gidEnd == gidStart {
			cursor = gidStart

			continue
		}

		if start > cursor {
			out = append(out, textNode(text[cursor:start], node.Marks))
		}

		userGID := text[gidStart:gidEnd]
		if attrs, ok := resolve(userGID); ok {
			mention, err := newMentionNode(attrs)
			if err == nil {
				out = append(out, mention)
				changed = true
				cursor = gidEnd

				continue
			}
		}

		out = append(out, textNode(text[start:gidEnd], node.Marks))
		cursor = gidEnd
	}

	if !changed {
		return []prosemirror.Node{node}, false
	}

	if cursor < len(text) {
		out = append(out, textNode(text[cursor:], node.Marks))
	}

	return out, true
}

func asanaLinkHref(node prosemirror.Node) string {
	if node.Type != prosemirror.NodeText || nodeHasMark(node, prosemirror.MarkCode) {
		return ""
	}

	for _, mark := range node.Marks {
		if mark.Type != prosemirror.MarkLink {
			continue
		}

		attrs, err := mark.LinkAttrs()
		if err != nil {
			return ""
		}

		return attrs.Href
	}

	return ""
}

func asanaProfileGID(origin, raw string) (string, bool) {
	prefix := strings.TrimRight(origin, "/") + "/0/profile/"
	if !strings.HasPrefix(raw, prefix) {
		return "", false
	}

	userGID := strings.TrimPrefix(raw, prefix)
	if !asanaNumericGID.MatchString(userGID) {
		return "", false
	}

	return userGID, true
}
