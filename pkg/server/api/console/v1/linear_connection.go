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

package console_v1

import (
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	tasksync "go.probo.inc/probo/pkg/task/sync"
)

func linearIssueNode(issue tasksync.LinearIssue) (*types.LinearIssue, error) {
	node := &types.LinearIssue{
		ID:         issue.ID,
		Identifier: issue.Identifier,
		Title:      issue.Title,
	}

	content, err := tasksync.MarkdownToContent(issue.Description)
	if err != nil {
		return nil, err
	}

	node.Content = &content

	if issue.StateType != "" {
		state := tasksync.LinearTypeToTaskState(issue.StateType)
		node.State = &state
	}

	if priority, ok := tasksync.LinearPriorityToTask(issue.Priority); ok {
		node.Priority = &priority
	}

	return node, nil
}

func linearPageInfo(startCursor, endCursor string, hasNextPage bool) *types.LinearPageInfo {
	info := &types.LinearPageInfo{
		HasNextPage:     hasNextPage && endCursor != "",
		HasPreviousPage: false,
	}
	if startCursor != "" {
		info.StartCursor = &startCursor
	}

	if endCursor != "" {
		info.EndCursor = &endCursor
	}

	return info
}

func emptyLinearTeamConnection() *types.LinearTeamConnection {
	return &types.LinearTeamConnection{
		Edges:    []*types.LinearTeamEdge{},
		PageInfo: linearPageInfo("", "", false),
	}
}

func emptyLinearIssueConnection() *types.LinearIssueConnection {
	return &types.LinearIssueConnection{
		Edges:    []*types.LinearIssueEdge{},
		PageInfo: linearPageInfo("", "", false),
	}
}
