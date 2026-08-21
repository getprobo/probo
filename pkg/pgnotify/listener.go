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

package pgnotify

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/log"
)

type Listener struct {
	config  *pgx.ConnConfig
	channel string
	logger  *log.Logger
	handler func(string)
}

const reconnectInterval = time.Second

func NewListener(
	config *pgx.ConnConfig,
	channel string,
	logger *log.Logger,
	handler func(string),
) *Listener {
	return &Listener{
		config:  config.Copy(),
		channel: channel,
		logger:  logger,
		handler: handler,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	for {
		if err := l.listen(ctx); err != nil {
			if ctx.Err() != nil {
				return nil
			}

			l.logger.WarnCtx(
				ctx,
				"PostgreSQL notification listener disconnected",
				log.Error(err),
			)
		}

		timer := time.NewTimer(reconnectInterval)
		select {
		case <-ctx.Done():
			timer.Stop()

			return nil
		case <-timer.C:
		}
	}
}

func (l *Listener) listen(ctx context.Context) error {
	connection, err := pgx.ConnectConfig(ctx, l.config)
	if err != nil {
		return fmt.Errorf("cannot connect PostgreSQL notification listener: %w", err)
	}

	defer func() { _ = connection.Close(context.Background()) }()

	channel := pgx.Identifier{l.channel}.Sanitize()
	if _, err := connection.Exec(ctx, "LISTEN "+channel); err != nil {
		return fmt.Errorf("cannot listen for PostgreSQL notifications: %w", err)
	}

	for {
		notification, err := connection.WaitForNotification(ctx)
		if err != nil {
			return fmt.Errorf("cannot wait for PostgreSQL notification: %w", err)
		}

		l.handler(notification.Payload)
	}
}
