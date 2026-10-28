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
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/notification"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

var linearProfileURLFinder = regexp.MustCompile(`(?i)\bhttps?://linear\.app/[^\s<>)\]]+`)

func textMayContainLinearProfile(value string) bool {
	lower := strings.ToLower(value)

	return strings.Contains(lower, "linear.app/") && strings.Contains(lower, "/profiles/")
}

func (s *Service) linearIssueDescriptionMarkdown(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	organizationID gid.GID,
	content string,
	issueID string,
) (string, error) {
	ids, err := prosemirror.MentionedProfileIDs(content)
	if err != nil {
		return "", fmt.Errorf("cannot read task mentions: %w", err)
	}

	if len(ids) == 0 {
		return ContentToMarkdown(content)
	}

	var remote *string

	if issueID != "" {
		issue, err := client.GetIssue(ctx, issueID)
		if err != nil {
			return "", fmt.Errorf("cannot load Linear issue for mentions: %w", err)
		}

		remote = &issue.Description
	}

	return s.linearOutboundMarkdown(ctx, scope, client, organizationID, content, remote)
}

func (s *Service) linearCommentMarkdown(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	organizationID gid.GID,
	content string,
	commentID string,
) (string, error) {
	ids, err := prosemirror.MentionedProfileIDs(content)
	if err != nil {
		return "", fmt.Errorf("cannot read comment mentions: %w", err)
	}

	if len(ids) == 0 {
		return ContentToMarkdown(content)
	}

	var remote *string

	if commentID != "" {
		body, err := client.CommentBody(ctx, commentID)
		if err != nil {
			return "", fmt.Errorf("cannot load Linear comment for mentions: %w", err)
		}

		remote = &body
	}

	return s.linearOutboundMarkdown(ctx, scope, client, organizationID, content, remote)
}

func (s *Service) linearOutboundMarkdown(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	organizationID gid.GID,
	content string,
	remote *string,
) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse prosemirror json: %w", err)
	}

	resolver := &outboundMentionResolver{
		service:        s,
		ctx:            ctx,
		scope:          scope,
		client:         client,
		organizationID: organizationID,
		cache:          map[string]string{},
	}

	rewritten := rewriteOutboundMentionNodes(node, func(id string, label string) (string, bool) {
		profileURL, ok := resolver.profileURL(id)
		if !ok || linearMentionKeepsPlainText(remote, label, profileURL) {
			return "", false
		}

		return profileURL, true
	})

	if resolver.err != nil {
		return "", resolver.err
	}

	encoded, err := json.Marshal(rewritten)
	if err != nil {
		return "", fmt.Errorf("cannot marshal Linear mention markdown: %w", err)
	}

	return ContentToMarkdown(string(encoded))
}

type outboundMentionResolver struct {
	service        *Service
	ctx            context.Context
	scope          coredata.Scoper
	client         *linear.Client
	organizationID gid.GID
	cache          map[string]string
	err            error
}

func (r *outboundMentionResolver) profileURL(profileID string) (string, bool) {
	if r.err != nil {
		return "", false
	}

	if profileURL, ok := r.cache[profileID]; ok {
		return profileURL, profileURL != ""
	}

	profileURL, err := r.lookup(profileID)
	if err != nil {
		r.err = err

		return "", false
	}

	r.cache[profileID] = profileURL

	return profileURL, profileURL != ""
}

func (r *outboundMentionResolver) lookup(profileID string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("cannot resolve Linear mention: client is not configured")
	}

	emails, err := r.membershipEmails(profileID)
	if err != nil {
		return "", err
	}

	for _, email := range emails {
		user, err := r.client.MentionUserByEmail(r.ctx, email)
		if err != nil {
			return "", fmt.Errorf("cannot look up Linear user: %w", err)
		}

		if user == nil || !user.Mentionable || user.URL == "" {
			continue
		}

		if _, ok := linear.ProfileSlug(user.URL); !ok {
			continue
		}

		return user.URL, nil
	}

	return "", nil
}

func (r *outboundMentionResolver) membershipEmails(profileID string) ([]string, error) {
	id, err := gid.ParseGID(profileID)
	if err != nil {
		return nil, nil
	}

	var emails []string

	err = r.service.pg.WithConn(
		r.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, conn, r.scope, id); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load mentioned profile: %w", err)
			}

			if profile.OrganizationID != r.organizationID || profile.State == coredata.ProfileStateDeactivated {
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
		return nil, err
	}

	return emails, nil
}

func (s *Service) contentFromLinearMarkdown(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	markdown string,
	metas map[string]importedFileMeta,
) (string, string, error) {
	content, err := MarkdownToContent(markdown)
	if err != nil {
		return "", "", err
	}

	content, _, err = applyImportedFiles(content, metas)
	if err != nil {
		return "", "", err
	}

	content, _, err = s.applyLinearProfileMentions(ctx, client, scope, organizationID, content)
	if err != nil {
		return "", "", err
	}

	canonical, err := ContentToMarkdown(content)
	if err != nil {
		return "", "", err
	}

	return content, canonical, nil
}

func (s *Service) attachLinearMentions(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	imported *linearImport,
) error {
	if imported == nil || !textMayContainLinearProfile(imported.markdown) {
		return nil
	}

	if client == nil {
		loaded, err := s.linearClientForConnectorID(ctx, scope, imported.connectorID)
		if err != nil {
			return fmt.Errorf("cannot resolve Linear mentions: %w", err)
		}

		client = loaded
	}

	content, canonical, err := s.contentFromLinearMarkdown(
		ctx,
		client,
		scope,
		organizationID,
		restoreLinearFiles(imported.markdown, imported.files),
		imported.metas,
	)
	if err != nil {
		return err
	}

	ids, err := prosemirror.MentionedProfileIDs(content)
	if err != nil {
		return fmt.Errorf("cannot read Linear mentions: %w", err)
	}

	if len(ids) == 0 {
		return nil
	}

	imported.mentionsApplied = true
	imported.mentionContent = content
	imported.mentionCanonical = canonical

	return nil
}

func (s *Service) applyLinearProfileMentions(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	content string,
) (string, bool, error) {
	if !textMayContainLinearProfile(content) {
		return content, false, nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", false, fmt.Errorf("cannot parse Linear content: %w", err)
	}

	resolver := &inboundMentionResolver{
		service:        s,
		ctx:            ctx,
		client:         client,
		scope:          scope,
		organizationID: organizationID,
		cache:          map[string]inboundMention{},
	}

	rewritten, changed := rewriteInboundMentionNodes(node, resolver.resolve)

	if resolver.err != nil {
		return "", false, resolver.err
	}

	if !changed {
		return content, false, nil
	}

	encoded, err := json.Marshal(rewritten)
	if err != nil {
		return "", false, fmt.Errorf("cannot marshal Linear mentions: %w", err)
	}

	encodedJSON := string(encoded)

	sanitized, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	if err != nil {
		return "", false, fmt.Errorf("cannot sanitize Linear mentions: %w", err)
	}

	return sanitized, true, nil
}

type inboundMention struct {
	attrs prosemirror.MentionAttrs
	ok    bool
}

type inboundMentionResolver struct {
	service        *Service
	ctx            context.Context
	client         *linear.Client
	scope          coredata.Scoper
	organizationID gid.GID
	cache          map[string]inboundMention
	err            error
}

func (r *inboundMentionResolver) resolve(raw string) (prosemirror.MentionAttrs, bool) {
	if r.err != nil {
		return prosemirror.MentionAttrs{}, false
	}

	key := strings.TrimSpace(raw)
	if hit, ok := r.cache[key]; ok {
		return hit.attrs, hit.ok
	}

	hit, err := r.lookup(key)
	if err != nil {
		r.err = err
		r.cache[key] = inboundMention{}

		return prosemirror.MentionAttrs{}, false
	}

	r.cache[key] = hit

	return hit.attrs, hit.ok
}

func (r *inboundMentionResolver) lookup(raw string) (inboundMention, error) {
	if _, ok := linear.ProfileSlug(raw); !ok {
		return inboundMention{}, nil
	}

	if r.client == nil {
		return inboundMention{}, fmt.Errorf("cannot resolve Linear mention: client is not configured")
	}

	user, err := r.client.MentionUserForProfileURL(r.ctx, raw)
	if err != nil {
		return inboundMention{}, fmt.Errorf("cannot look up Linear profile: %w", err)
	}

	if user == nil || user.Email == "" {
		return inboundMention{}, nil
	}

	var profile *coredata.MembershipProfile

	err = r.service.pg.WithConn(
		r.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := activeProfileByEmail(ctx, conn, r.scope, r.organizationID, user.Email)
			if err != nil {
				return err
			}

			profile = loaded

			return nil
		},
	)
	if err != nil {
		return inboundMention{}, fmt.Errorf("cannot match Linear profile: %w", err)
	}

	if profile == nil {
		return inboundMention{}, nil
	}

	label := mentionLabel(profile.FullName, user.Name, raw)
	if label == "" {
		return inboundMention{}, nil
	}

	return inboundMention{
		ok: true,
		attrs: prosemirror.MentionAttrs{
			ID:    profile.ID.String(),
			Label: label,
		},
	}, nil
}

func mentionLabel(profileName string, linearName string, rawURL string) string {
	label := strings.TrimSpace(profileName)
	if label == "" {
		label = strings.TrimSpace(linearName)
	}

	if label == "" {
		slug, ok := linear.ProfileSlug(rawURL)
		if ok {
			label = strings.TrimSpace(slug)
		}
	}

	return truncateMentionLabel(label)
}

func truncateMentionLabel(label string) string {
	if len(label) <= prosemirror.MaxMentionLabelBytes {
		return label
	}

	cut := label[:prosemirror.MaxMentionLabelBytes]
	for cut != "" && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}

	return strings.TrimSpace(cut)
}

func (s *Service) notifySyncedCommentMentions(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	taskID gid.GID,
	oldContent string,
	newContent string,
	actorProfileID *gid.GID,
) error {
	task := &coredata.Task{}
	if err := task.LoadByID(ctx, tx, scope, taskID); err != nil {
		return fmt.Errorf("cannot load task for mention notification: %w", err)
	}

	return s.notifySyncedMentions(
		ctx,
		tx,
		scope,
		task.OrganizationID,
		task.ID,
		task.Name,
		oldContent,
		newContent,
		actorProfileID,
	)
}

func (s *Service) notifySyncedMentions(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	taskID gid.GID,
	taskName string,
	oldContent string,
	newContent string,
	actorProfileID *gid.GID,
) error {
	if oldContent == newContent {
		return nil
	}

	mentionedIDs, err := prosemirror.NewlyMentionedProfileIDs(oldContent, newContent)
	if err != nil {
		return fmt.Errorf("cannot collect new mentions: %w", err)
	}

	if len(mentionedIDs) == 0 {
		return nil
	}

	if err := notification.InsertTaskMentions(
		ctx,
		tx,
		scope,
		organizationID,
		taskID,
		taskName,
		mentionedIDs,
		actorProfileID,
	); err != nil {
		return fmt.Errorf("cannot insert task mention notifications: %w", err)
	}

	return nil
}

func linearMentionKeepsPlainText(remote *string, label string, profileURL string) bool {
	if remote == nil {
		return false
	}

	if containsBoundedToken(*remote, profileURL, continuesURL) {
		return false
	}

	return containsBoundedToken(*remote, renderedMentionText(label), continuesMention)
}

func renderedMentionText(label string) string {
	attrs, err := json.Marshal(prosemirror.MentionAttrs{ID: "profile", Label: label})
	if err != nil {
		return "@" + label
	}

	document := prosemirror.Node{
		Type: prosemirror.NodeDoc,
		Content: []prosemirror.Node{
			{
				Type: prosemirror.NodeParagraph,
				Content: []prosemirror.Node{
					{Type: prosemirror.NodeMention, Attrs: attrs},
				},
			},
		},
	}

	markdown, err := prosemirror.RenderMarkdown(document)
	if err != nil {
		return "@" + label
	}

	return strings.TrimSpace(markdown)
}

func containsBoundedToken(haystack string, token string, continues func(rune) bool) bool {
	if token == "" || haystack == "" {
		return false
	}

	rest := haystack
	offset := 0

	for {
		index := strings.Index(rest, token)
		if index < 0 {
			return false
		}

		start := offset + index
		end := start + len(token)
		beforeOK := start == 0 || !continues(lastRune(haystack[:start]))
		afterOK := end >= len(haystack) || !continues(firstRune(haystack[end:]))

		if beforeOK && afterOK {
			return true
		}

		next := index + len(token)
		if next > len(rest) {
			return false
		}

		rest = rest[next:]
		offset = end
	}
}

func continuesMention(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func continuesURL(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}

	switch r {
	case '/', ':', '.', '_', '~', '%', '-', '?', '&', '=', '#':
		return true
	default:
		return false
	}
}

func firstRune(value string) rune {
	r, _ := utf8.DecodeRuneInString(value)

	return r
}

func lastRune(value string) rune {
	r, _ := utf8.DecodeLastRuneInString(value)

	return r
}

func rewriteOutboundMentionNodes(
	node prosemirror.Node,
	profileURL func(id string, label string) (string, bool),
) prosemirror.Node {
	if node.Type == prosemirror.NodeMention {
		attrs, err := node.MentionAttrs()
		if err != nil {
			return node
		}

		url, ok := profileURL(attrs.ID, attrs.Label)
		if !ok {
			return node
		}

		text := url

		return prosemirror.Node{Type: prosemirror.NodeText, Text: &text}
	}

	if node.Type == prosemirror.NodeCodeBlock {
		return node
	}

	next := node
	next.Content = make([]prosemirror.Node, len(node.Content))

	for i, child := range node.Content {
		next.Content[i] = rewriteOutboundMentionNodes(child, profileURL)
	}

	return next
}

func rewriteInboundMentionNodes(
	node prosemirror.Node,
	resolve func(string) (prosemirror.MentionAttrs, bool),
) (prosemirror.Node, bool) {
	if node.Type == prosemirror.NodeCodeBlock {
		return node, false
	}

	content, changed := rewriteInboundMentionChildren(node.Content, resolve)
	node.Content = content

	return node, changed
}

func rewriteInboundMentionChildren(
	nodes []prosemirror.Node,
	resolve func(string) (prosemirror.MentionAttrs, bool),
) ([]prosemirror.Node, bool) {
	var (
		out     []prosemirror.Node
		changed bool
	)

	for index := 0; index < len(nodes); {
		if href, ok := profileLinkHref(nodes[index]); ok {
			end := index + 1
			for end < len(nodes) && profileLinkHrefEquals(nodes[end], href) {
				end++
			}

			if attrs, ok := resolve(href); ok {
				mention, err := newMentionNode(attrs)
				if err == nil {
					out = append(out, mention)
					changed = true
					index = end

					continue
				}
			}
		}

		node := nodes[index]
		if node.Type == prosemirror.NodeText && !nodeHasMark(node, prosemirror.MarkCode) && !nodeHasMark(node, prosemirror.MarkLink) {
			split, replaced := splitProfileURLText(node, resolve)
			if replaced {
				changed = true
			}

			out = append(out, split...)
			index++

			continue
		}

		rewritten, childChanged := rewriteInboundMentionNodes(node, resolve)
		if childChanged {
			changed = true
		}

		out = append(out, rewritten)
		index++
	}

	return out, changed
}

func profileLinkHref(node prosemirror.Node) (string, bool) {
	if node.Type != prosemirror.NodeText || nodeHasMark(node, prosemirror.MarkCode) {
		return "", false
	}

	for _, mark := range node.Marks {
		if mark.Type != prosemirror.MarkLink {
			continue
		}

		attrs, err := mark.LinkAttrs()
		if err != nil {
			continue
		}

		if _, ok := linear.ProfileSlug(attrs.Href); ok {
			return attrs.Href, true
		}
	}

	return "", false
}

func profileLinkHrefEquals(node prosemirror.Node, href string) bool {
	got, ok := profileLinkHref(node)

	return ok && got == href
}

func splitProfileURLText(
	node prosemirror.Node,
	resolve func(string) (prosemirror.MentionAttrs, bool),
) ([]prosemirror.Node, bool) {
	if node.Text == nil {
		return []prosemirror.Node{node}, false
	}

	text := *node.Text
	spans := profileURLSpans(text)

	if len(spans) == 0 {
		return []prosemirror.Node{node}, false
	}

	var (
		out     []prosemirror.Node
		cursor  int
		changed bool
	)

	for _, span := range spans {
		if span.start < cursor {
			continue
		}

		if span.start > cursor {
			out = append(out, textNode(text[cursor:span.start], node.Marks))
		}

		if attrs, ok := resolve(span.raw); ok {
			mention, err := newMentionNode(attrs)
			if err == nil {
				out = append(out, mention)
				changed = true
				cursor = span.end

				continue
			}
		}

		out = append(out, textNode(text[span.start:span.end], node.Marks))
		cursor = span.end
	}

	if cursor < len(text) {
		out = append(out, textNode(text[cursor:], node.Marks))
	}

	if !changed {
		return []prosemirror.Node{node}, false
	}

	return out, true
}

type profileURLSpan struct {
	start int
	end   int
	raw   string
}

func profileURLSpans(text string) []profileURLSpan {
	indexes := linearProfileURLFinder.FindAllStringIndex(text, -1)
	spans := make([]profileURLSpan, 0, len(indexes))

	for _, pair := range indexes {
		raw := trimURLTail(text[pair[0]:pair[1]])
		if _, ok := linear.ProfileSlug(raw); !ok {
			continue
		}

		spans = append(spans, profileURLSpan{
			start: pair[0],
			end:   pair[0] + len(raw),
			raw:   raw,
		})
	}

	return spans
}

func trimURLTail(raw string) string {
	return strings.TrimRight(raw, ".,;:!?")
}

func newMentionNode(attrs prosemirror.MentionAttrs) (prosemirror.Node, error) {
	raw, err := json.Marshal(attrs)
	if err != nil {
		return prosemirror.Node{}, err
	}

	return prosemirror.Node{Type: prosemirror.NodeMention, Attrs: raw}, nil
}

func textNode(value string, marks []prosemirror.Mark) prosemirror.Node {
	text := value

	return prosemirror.Node{
		Type:  prosemirror.NodeText,
		Text:  &text,
		Marks: append([]prosemirror.Mark(nil), marks...),
	}
}

func nodeHasMark(node prosemirror.Node, kind prosemirror.MarkType) bool {
	for _, mark := range node.Marks {
		if mark.Type == kind {
			return true
		}
	}

	return false
}
