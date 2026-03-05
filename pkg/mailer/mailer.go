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
	"sync"
	"time"

	"github.com/jhillyerd/enmime"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
)

type (
	SendingWorker struct {
		pg             *pg.Client
		fileManager    *filemanager.Service
		logger         *log.Logger
		smtp           SMTPConfig
		senderName     string
		senderEmail    string
		interval       time.Duration
		smtpTimeout    time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	SMTPConfig struct {
		Addr        string
		User        string
		Password    string
		TLSRequired bool
	}

	SendingWorkerOption func(*SendingWorker)
)

func WithSendingWorkerInterval(d time.Duration) SendingWorkerOption {
	return func(w *SendingWorker) { w.interval = d }
}

func WithSendingWorkerSMTPTimeout(d time.Duration) SendingWorkerOption {
	return func(w *SendingWorker) { w.smtpTimeout = d }
}

func WithSendingWorkerStaleAfter(d time.Duration) SendingWorkerOption {
	return func(w *SendingWorker) { w.staleAfter = d }
}

func WithSendingWorkerMaxConcurrency(n int) SendingWorkerOption {
	return func(w *SendingWorker) {
		if n > 0 {
			w.maxConcurrency = n
		}
	}
}

func NewSendingWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	senderName string,
	senderEmail string,
	smtpCfg SMTPConfig,
	logger *log.Logger,
	opts ...SendingWorkerOption,
) *SendingWorker {
	w := &SendingWorker{
		pg:             pgClient,
		fileManager:    fileManager,
		logger:         logger,
		smtp:           smtpCfg,
		senderName:     senderName,
		senderEmail:    senderEmail,
		interval:       30 * time.Second,
		smtpTimeout:    25 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 20,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *SendingWorker) Run(ctx context.Context) error {
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, w.maxConcurrency)
	)

	defer wg.Wait()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			nonCancelableCtx := context.WithoutCancel(ctx)
			w.recoverStaleRows(nonCancelableCtx)

			for {
				if err := w.processNext(ctx, sem, &wg); err != nil {
					if !errors.Is(err, coredata.ErrNoUnsentEmail) {
						w.logger.ErrorCtx(nonCancelableCtx, "cannot process email", log.Error(err))
					}
					break
				}
			}
		}
	}
}

func (w *SendingWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		email            = coredata.Email{}
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := email.LoadNextPendingForUpdateSkipLocked(nonCancelableCtx, tx); err != nil {
				return err
			}

			now := time.Now()
			email.Status = coredata.EmailStatusProcessing
			email.ProcessingStartedAt = &now
			email.AttemptCount++
			email.LastAttemptedAt = &now
			email.UpdatedAt = now

			if err := email.Update(nonCancelableCtx, tx); err != nil {
				return fmt.Errorf("cannot update email: %w", err)
			}

			return nil
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(email coredata.Email) {
		defer wg.Done()
		defer func() { <-sem }()

		if sendErr := w.sendAndCommit(nonCancelableCtx, &email); sendErr != nil {
			if failErr := w.failEmail(nonCancelableCtx, &email, sendErr); failErr != nil {
				w.logger.ErrorCtx(nonCancelableCtx, "cannot fail email", log.Error(failErr))
			}
		}
	}(email)

	return nil
}

func (w *SendingWorker) sendAndCommit(
	ctx context.Context,
	email *coredata.Email,
) error {
	var buf bytes.Buffer

	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var attachments coredata.EmailAttachments
			if err := attachments.LoadByEmailID(ctx, conn, email.ID); err != nil {
				return fmt.Errorf("cannot load email attachments: %w", err)
			}

			mail := enmime.Builder().
				Subject(email.Subject).
				From(w.senderName, w.senderEmail).
				To(email.RecipientName, email.RecipientEmail).
				Text([]byte(email.TextBody))

			if email.ReplyTo != nil {
				mail = mail.ReplyTo("", email.ReplyTo.String())
			}

			if email.HtmlBody != nil {
				mail = mail.HTML([]byte(*email.HtmlBody))
			}

			if email.UnsubscribeURL != nil {
				mail = mail.
					Header("List-Unsubscribe", "<"+*email.UnsubscribeURL+">").
					Header("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")
			}

			for _, att := range attachments {
				var file coredata.File
				if err := file.LoadByID(ctx, conn, coredata.NewNoScope(), att.FileID); err != nil {
					return fmt.Errorf("cannot load file record for attachment %s: %w", att.Filename, err)
				}

				data, err := w.fileManager.GetFileBytes(ctx, &file)
				if err != nil {
					return fmt.Errorf("cannot download attachment %s: %w", att.Filename, err)
				}

				mail = mail.AddAttachment(data, file.MimeType, att.Filename)
			}

			envelope, err := mail.Build()
			if err != nil {
				return fmt.Errorf("cannot build email: %w", err)
			}

			if err := envelope.Encode(&buf); err != nil {
				return fmt.Errorf("cannot encode email: %w", err)
			}

			return nil
		},
	); err != nil {
		return err
	}

	sendCtx, cancel := context.WithTimeout(ctx, w.smtpTimeout)
	defer cancel()

	if err := w.sendMail(sendCtx, []string{email.RecipientEmail}, buf.Bytes()); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("email sending timed out after %s: %w", w.smtpTimeout, err)
		}
		return fmt.Errorf("cannot send email: %w", err)
	}

	if err := w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			now := time.Now()
			email.Status = coredata.EmailStatusSent
			email.SentAt = &now
			email.ProcessingStartedAt = nil
			email.LastError = nil
			email.UpdatedAt = now

			if err := email.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update email: %w", err)
			}

			return nil
		},
	); err != nil {
		w.logger.ErrorCtx(ctx,
			"email sent but failed to commit status update; will not re-queue to avoid duplicate delivery",
			log.Error(err),
			log.String("email_id", email.ID.String()),
		)
	}

	return nil
}

func (w *SendingWorker) failEmail(
	ctx context.Context,
	email *coredata.Email,
	processingError error,
) error {
	w.logger.ErrorCtx(ctx, "sending worker failure",
		log.Error(processingError),
		log.String("email_id", email.ID.String()),
	)

	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			errStr := processingError.Error()
			email.LastError = &errStr
			email.ProcessingStartedAt = nil
			email.UpdatedAt = time.Now()

			if email.AttemptCount >= email.MaxAttempts {
				email.Status = coredata.EmailStatusFailed
			} else {
				email.Status = coredata.EmailStatusPending
			}

			if err := email.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update email: %w", err)
			}

			return nil
		},
	)
}

func (w *SendingWorker) recoverStaleRows(ctx context.Context) {
	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return coredata.ResetStaleProcessingEmails(ctx, conn, w.staleAfter)
		},
	); err != nil {
		w.logger.ErrorCtx(ctx, "cannot recover stale emails", log.Error(err))
	}
}

func (w *SendingWorker) sendMail(ctx context.Context, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(w.smtp.Addr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", w.smtp.Addr)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("cannot set connection deadline: %w", err)
		}
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client creation error: %w", err)
	}
	defer func() { _ = c.Quit() }()

	if w.smtp.TLSRequired {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("TLS negotiation error: %w", err)
		}
	}

	if w.smtp.User != "" && w.smtp.Password != "" {
		auth := smtp.PlainAuth("", w.smtp.User, w.smtp.Password, host)
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication error: %w", err)
		}
	}

	if err = c.Mail(w.senderEmail); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO error: %w", err)
		}
	}

	wr, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA command error: %w", err)
	}

	_, err = wr.Write(msg)
	if err != nil {
		return fmt.Errorf("message write error: %w", err)
	}

	if err = wr.Close(); err != nil {
		return fmt.Errorf("message close error: %w", err)
	}

	return nil
}
