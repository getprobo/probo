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

package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

type (
	InteractiveReply struct {
		ResponseType    string `json:"response_type"`
		Text            string `json:"text"`
		ReplaceOriginal bool   `json:"replace_original"`
	}

	interactiveReplyPoster interface {
		PostEphemeralReply(ctx context.Context, responseURL, text string) error
	}

	responseURLPoster struct {
		httpClient *http.Client
	}
)

func UnboundInteractiveResponse() InteractiveReply {
	return newInteractiveReply(bindRequiredText)
}

func ForbiddenInteractiveResponse() InteractiveReply {
	return newInteractiveReply(interactiveForbiddenText)
}

func newInteractiveReply(text string) InteractiveReply {
	return InteractiveReply{
		ResponseType:    SlashResponseTypeEphemeral,
		Text:            text,
		ReplaceOriginal: false,
	}
}

func newResponseURLPoster(logger *log.Logger) *responseURLPoster {
	return &responseURLPoster{
		httpClient: httpclient.DefaultPooledClient(
			httpclient.WithLogger(logger),
			httpclient.WithSSRFProtection(),
		),
	}
}

func (s *Service) ReplyInteractiveEphemeral(
	ctx context.Context,
	responseURL string,
	text string,
) error {
	if s == nil {
		return fmt.Errorf("cannot post Slack interactive reply: service unavailable")
	}

	return postInteractiveEphemeral(ctx, s.httpClient, responseURL, text)
}

func (p *responseURLPoster) PostEphemeralReply(
	ctx context.Context,
	responseURL string,
	text string,
) error {
	if p == nil {
		return fmt.Errorf("cannot post Slack interactive reply: poster unavailable")
	}

	return postInteractiveEphemeral(ctx, p.httpClient, responseURL, text)
}

func postInteractiveEphemeral(
	ctx context.Context,
	httpClient *http.Client,
	responseURL string,
	text string,
) error {
	if responseURL == "" {
		return nil
	}

	if httpClient == nil {
		return fmt.Errorf("cannot post Slack interactive reply: client unavailable")
	}

	normalized, err := normalizeSlackResponseURL(responseURL)
	if err != nil {
		return fmt.Errorf("cannot normalize Slack interactive reply URL: %w", err)
	}

	body, err := json.Marshal(newInteractiveReply(text))
	if err != nil {
		return fmt.Errorf("cannot marshal Slack interactive reply: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		normalized,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("cannot create Slack interactive reply request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot post Slack interactive reply: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("cannot post Slack interactive reply: status %d", resp.StatusCode)
	}

	return nil
}
