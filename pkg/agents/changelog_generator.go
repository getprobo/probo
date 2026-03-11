// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package agents

import (
	"context"
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

const (
	changelogGeneratorSystemPrompt = `
		# Role:You are an assistant that creates clear and concise changelogs.

		# Objective
		Given two versions of a document — the "old version" and the "new version" — identify and summarize all meaningful changes between them.
		Focus on additions, deletions, modifications, and restructuring.

		# Response Format
		Respond with ONE simple phrase that describe the changes.

		# Change types
		If possible use the following words with additional context to describe the change types:
			"Added", "Removed", "Updated", "Reworded", "Reorganized", "Fixed", etc.

		# SOP
		- Be objective and neutral in tone.
		- Do not comment on the quality of the change.
		- Use the language of the document.

		**Example output format:**
		Respond ONLY with the phrase that describes the changes. No explanation, no markdown, no preamble. Like this:
			Added clauses about sharing personal information with trusted partners
	`
)

func (a *Agent) GenerateChangelog(ctx context.Context, oldContent string, newContent string) (*string, error) {
	resp, err := a.client.ChatCompletion(ctx, &llm.ChatCompletionRequest{
		Model: a.model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Parts: []llm.Part{llm.TextPart{Text: changelogGeneratorSystemPrompt}}},
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: fmt.Sprintf(`Old content: %s`, oldContent)}}},
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: fmt.Sprintf(`New content: %s`, newContent)}}},
		},
		Temperature: &a.temp,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot generate changelog: %w", err)
	}

	text := resp.Message.Text()
	return &text, nil
}
