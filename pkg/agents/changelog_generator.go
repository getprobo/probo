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

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
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
	model := openai.ChatModel(a.cfg.ModelName)
	chatCompletion, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(changelogGeneratorSystemPrompt),
			openai.UserMessage(fmt.Sprintf(`Old content: %s`, oldContent)),
			openai.UserMessage(fmt.Sprintf(`New content: %s`, newContent)),
		},
		Model:       model,
		Temperature: param.NewOpt(a.cfg.Temperature),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot parse vendor info: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned from API")
	}

	return &chatCompletion.Choices[0].Message.Content, nil
}
