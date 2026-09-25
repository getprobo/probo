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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

const (
	linearPickerPageSizeDefault = 20
	linearPickerPageSizeMax     = 50
	linearListMaxPages          = 500
)

type (
	LinearTeamPage struct {
		Teams       []LinearTeam
		EndCursor   string
		HasNextPage bool
	}

	LinearIssue struct {
		ID          string
		Identifier  string
		Title       string
		Description string
		StateType   string
		Priority    int
	}

	LinearIssuePage struct {
		Issues      []LinearIssue
		EndCursor   string
		HasNextPage bool
	}

	teamPageCursor struct {
		Accounts []teamAccountCursor `json:"accounts"`
	}

	teamAccountCursor struct {
		ConnectorID string  `json:"c"`
		After       *string `json:"a,omitempty"`
		Done        bool    `json:"d,omitempty"`
	}
)

func (s *Service) SearchLinearTeams(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	query string,
	first int,
	after *string,
) (LinearTeamPage, error) {
	accounts, err := s.linearAccountsForOrganization(ctx, scope, organizationID)
	if err != nil {
		return LinearTeamPage{}, err
	}

	first = clampLinearPageSize(first)

	positions, err := decodeTeamPageCursor(after, accounts)
	if err != nil {
		return LinearTeamPage{}, err
	}

	teams := make([]LinearTeam, 0, first)
	seen := make(map[string]struct{}, first)
	remaining := first

	var listErr error

	for i := range accounts {
		for remaining > 0 && !positions[i].Done {
			page, err := accounts[i].client.SearchTeams(ctx, query, remaining, positions[i].After)
			if err != nil {
				listErr = err

				positions[i].Done = true
				if s.logger != nil {
					s.logger.WarnCtx(
						ctx,
						"cannot search Linear teams for connector",
						log.String("connector_id", accounts[i].connector.ID.String()),
						log.Error(err),
					)
				}

				break
			}

			if len(page.Teams) == 0 {
				positions[i].Done = true
				positions[i].After = nil

				break
			}

			for _, team := range page.Teams {
				if _, ok := seen[team.ID]; ok {
					continue
				}

				seen[team.ID] = struct{}{}
				teams = append(teams, LinearTeam{
					ID:   team.ID,
					Name: team.Name,
					Key:  team.Key,
				})

				remaining--
				if remaining == 0 {
					break
				}
			}

			if page.HasNextPage {
				next := page.EndCursor
				positions[i].After = &next

				break
			}

			positions[i].Done = true
			positions[i].After = nil
		}

		if remaining == 0 {
			break
		}
	}

	if len(teams) == 0 && listErr != nil {
		return LinearTeamPage{}, fmt.Errorf("cannot search Linear teams: %w", listErr)
	}

	hasNext := false

	for _, position := range positions {
		if !position.Done {
			hasNext = true
			break
		}
	}

	endCursor := ""

	if hasNext {
		encoded, err := encodeTeamPageCursor(positions)
		if err != nil {
			return LinearTeamPage{}, err
		}

		endCursor = encoded
	}

	return LinearTeamPage{
		Teams:       teams,
		EndCursor:   endCursor,
		HasNextPage: hasNext,
	}, nil
}

func (s *Service) SearchLinearIssues(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	teamID string,
	query string,
	first int,
	after *string,
) (LinearIssuePage, error) {
	if strings.TrimSpace(teamID) == "" {
		return LinearIssuePage{}, ErrLinearTeamIDRequired
	}

	accounts, err := s.linearAccountsForOrganization(ctx, scope, organizationID)
	if err != nil {
		return LinearIssuePage{}, err
	}

	account, err := s.linearAccountForTeam(ctx, accounts, teamID)
	if err != nil {
		return LinearIssuePage{}, err
	}

	page, err := account.client.SearchIssues(ctx, teamID, query, clampLinearPageSize(first), after)
	if err != nil {
		return LinearIssuePage{}, fmt.Errorf("cannot search Linear issues: %w", err)
	}

	issues := make([]LinearIssue, 0, len(page.Issues))
	for _, issue := range page.Issues {
		issues = append(issues, LinearIssue{
			ID:          issue.ID,
			Identifier:  issue.Identifier,
			Title:       issue.Title,
			Description: issue.Description,
			StateType:   issue.StateType,
			Priority:    issue.Priority,
		})
	}

	return LinearIssuePage{
		Issues:      issues,
		EndCursor:   page.EndCursor,
		HasNextPage: page.HasNextPage,
	}, nil
}

func clampLinearPageSize(first int) int {
	if first <= 0 {
		return linearPickerPageSizeDefault
	}

	if first > linearPickerPageSizeMax {
		return linearPickerPageSizeMax
	}

	return first
}

func decodeTeamPageCursor(after *string, accounts []linearAccount) ([]teamAccountCursor, error) {
	positions := make([]teamAccountCursor, len(accounts))
	for i := range accounts {
		positions[i] = teamAccountCursor{ConnectorID: accounts[i].connector.ID.String()}
	}

	if after == nil || strings.TrimSpace(*after) == "" {
		return positions, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(*after)
	if err != nil {
		return nil, fmt.Errorf("cannot decode Linear team page cursor: %w", err)
	}

	var decoded teamPageCursor
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("cannot decode Linear team page cursor: %w", err)
	}

	byConnector := make(map[string]teamAccountCursor, len(decoded.Accounts))
	for _, account := range decoded.Accounts {
		byConnector[account.ConnectorID] = account
	}

	for i := range positions {
		if saved, ok := byConnector[positions[i].ConnectorID]; ok {
			positions[i] = saved
		}
	}

	return positions, nil
}

func encodeTeamPageCursor(positions []teamAccountCursor) (string, error) {
	raw, err := json.Marshal(teamPageCursor{Accounts: positions})
	if err != nil {
		return "", fmt.Errorf("cannot encode Linear team page cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func linearIssueNotFound(err error) bool {
	return errors.Is(err, linear.ErrIssueNotFound)
}
