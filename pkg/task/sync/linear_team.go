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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *Service) TeamForNewTask(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	explicitSet bool,
	explicitTeamID *string,
) (string, bool, error) {
	defaultTeam, err := s.LinearDefaultTeam(ctx, scope, organizationID)
	if err != nil {
		return "", false, fmt.Errorf("cannot load default Linear team: %w", err)
	}

	defaultTeamID := ""
	if defaultTeam != nil {
		defaultTeamID = defaultTeam.ID
	}

	teamID, publish := LinearTeamForCreate(explicitSet, explicitTeamID, defaultTeamID)

	return teamID, publish, nil
}

func LinearTeamForCreate(explicitSet bool, explicit *string, defaultTeamID string) (teamID string, publish bool) {
	if !explicitSet {
		if defaultTeamID == "" {
			return "", false
		}

		return defaultTeamID, true
	}

	if explicit == nil || strings.TrimSpace(*explicit) == "" {
		return "", false
	}

	return strings.TrimSpace(*explicit), true
}
