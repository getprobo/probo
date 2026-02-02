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

package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"github.com/jhillyerd/enmime"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
)

type (
	Mailer struct {
		pg       *pg.Client
		l        *log.Logger
		cfg      Config
		interval time.Duration
	}

	Config struct {
		SenderName  string
		SenderEmail string
		Addr        string
		Timeout     time.Duration
		User        string
		Password    string
		TLSRequired bool
		Interval    time.Duration
	}
)

func NewMailer(pg *pg.Client, l *log.Logger, cfg Config) *Mailer {
	// Set a default timeout if not provided
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	// Set a default interval if not provided
	if cfg.Interval == 0 {
		cfg.Interval = 60 * time.Second
	}

	return &Mailer{pg: pg, l: l, cfg: cfg, interval: cfg.Interval}
}

func (m *Mailer) Run(ctx context.Context) error {
LOOP:
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.interval):
		ctx := context.Background()
		if err := m.batchSendEmails(ctx); err != nil {
			m.l.ErrorCtx(ctx, "cannot send email", log.Error(err))
		}

		goto LOOP
	}
}

func (m *Mailer) sendMailWithTimeout(ctx context.Context, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(m.cfg.Addr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	var d net.Dialer
	d.Timeout = 5 * time.Second

	conn, err := d.DialContext(ctx, "tcp", m.cfg.Addr)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer func() { _ = conn.Close() }()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client creation error: %w", err)
	}
	defer func() { _ = c.Quit() }()

	if m.cfg.TLSRequired {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("TLS negotiation error: %w", err)
		}
	}

	if m.cfg.User != "" && m.cfg.Password != "" {
		auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Password, host)
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication error: %w", err)
		}
	}

	if err = c.Mail(m.cfg.SenderEmail); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO error: %w", err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA command error: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("message write error: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("message close error: %w", err)
	}

	return c.Quit()
}

func (m *Mailer) batchSendEmails(ctx context.Context) error {
	for {
		err := m.pg.WithTx(
			ctx,
			func(tx pg.Conn) error {
				email := &coredata.Email{}
				err := email.LoadNextUnsentForUpdate(ctx, tx)
				if err != nil {
					return err
				}

				mail := enmime.Builder().
					Subject(email.Subject).
					From(m.cfg.SenderName, m.cfg.SenderEmail).
					To(email.RecipientName, email.RecipientEmail).
					Text([]byte(email.TextBody))

				if email.HtmlBody != nil {
					mail = mail.HTML([]byte(*email.HtmlBody))
				}

				envelope, err := mail.Build()
				if err != nil {
					return fmt.Errorf("cannot build email: %w", err)
				}

				var buf bytes.Buffer
				if err := envelope.Encode(&buf); err != nil {
					return fmt.Errorf("cannot encode email: %w", err)
				}

				sendCtx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
				defer cancel()

				if err := m.sendMailWithTimeout(sendCtx, []string{email.RecipientEmail}, buf.Bytes()); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						return fmt.Errorf("email sending timed out after %s: %w", m.cfg.Timeout, err)
					}
					return fmt.Errorf("cannot send email: %w", err)
				}

				now := time.Now()
				email.SentAt = &now
				email.UpdatedAt = now

				if err := email.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update email: %w", err)
				}

				return nil
			},
		)

		if errors.Is(err, coredata.ErrNoUnsentEmail) {
			return nil
		}

		if err != nil {
			return err
		}
	}
}
