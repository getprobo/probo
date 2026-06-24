// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"context"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

type Service struct {
	pg                 *pg.Client
	logger             *log.Logger
	slackSigningSecret string
	baseURL            string
	tokenSecret        string
}

func NewService(
	pg *pg.Client,
	slackSigningSecret string,
	baseURL string,
	tokenSecret string,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:                 pg,
		logger:             logger,
		slackSigningSecret: slackSigningSecret,
		baseURL:            baseURL,
		tokenSecret:        tokenSecret,
	}
}

func (s *Service) GetSlackClient() *Client {
	return NewClient(s.logger)
}

func (s *Service) GetSlackSigningSecret() string {
	return s.slackSigningSecret
}

func (s *Service) GetInitialSlackMessageByChannelAndTS(
	ctx context.Context,
	channelID string,
	messageTS string,
) (*coredata.SlackMessage, error) {
	var slackMessage coredata.SlackMessage

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := slackMessage.LoadInitialByChannelAndTS(ctx, conn, coredata.NewNoScope(), channelID, messageTS); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &slackMessage, nil
}
