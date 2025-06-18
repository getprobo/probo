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

package openai

import (
	"context"

	"github.com/getprobo/probo/pkg/llmgw"
	"github.com/openai/openai-go"
)

type (
	Provider struct {
		client *openai.Client
	}
)

func New(client *openai.Client) *Provider {
	return &Provider{client: client}
}

func (g *Provider) Generate(ctx context.Context, req llmgw.GenerateRequest) (*llmgw.GenerateResponse, error) {
	model := openai.ChatModel(req.Model)

	params := openai.ChatCompletionNewParams{
		Model:       model,
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Temperature: openai.Float(req.Temperature),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
	}

	response, err := g.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return &llmgw.GenerateResponse{}, nil
	}

	return &llmgw.GenerateResponse{Text: response.Choices[0].Message.Content}, nil
}

func (g *Provider) Chat(ctx context.Context, req llmgw.ChatRequest) (*llmgw.ChatResponse, error) {
	model := openai.ChatModel(req.Model)

	var messages []openai.ChatCompletionMessageParamUnion
	for _, msg := range req.Messages {
		switch msg.Role {
		case llmgw.RoleUser:
			messages = append(messages, openai.UserMessage(msg.Content))
		case llmgw.RoleSystem:
			messages = append(messages, openai.SystemMessage(msg.Content))
		case llmgw.RoleAssistant:
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:       model,
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Temperature: openai.Float(req.Temperature),
		Messages:    messages,
	}

	response, err := g.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return &llmgw.ChatResponse{}, nil
	}

	return &llmgw.ChatResponse{Text: response.Choices[0].Message.Content}, nil
}
