// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

	"github.com/jhillyerd/enmime"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
)

type (
	sendingHandler struct {
		pg          *pg.Client
		fileManager *filemanager.Service
		logger      *log.Logger
		smtp        SMTPConfig
		senderName  string
		senderEmail string
		smtpTimeout time.Duration
		staleAfter  time.Duration
	}

	SMTPConfig struct {
		Addr        string
		User        string
		Password    string
		TLSRequired bool
	}

	SendingWorkerOption func(*sendingHandler)
)

func WithSendingWorkerSMTPTimeout(d time.Duration) SendingWorkerOption {
	return func(h *sendingHandler) { h.smtpTimeout = d }
}

func WithSendingWorkerStaleAfter(d time.Duration) SendingWorkerOption {
	return func(h *sendingHandler) { h.staleAfter = d }
}

func NewSendingWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	senderName string,
	senderEmail string,
	smtpCfg SMTPConfig,
	logger *log.Logger,
	handlerOpts []SendingWorkerOption,
	workerOpts ...worker.Option,
) *worker.Worker[coredata.Email] {
	h := &sendingHandler{
		pg:          pgClient,
		fileManager: fileManager,
		logger:      logger,
		smtp:        smtpCfg,
		senderName:  senderName,
		senderEmail: senderEmail,
		smtpTimeout: 25 * time.Second,
		staleAfter:  5 * time.Minute,
	}

	for _, opt := range handlerOpts {
		opt(h)
	}

	return worker.New(
		"sending-worker",
		h,
		logger,
		workerOpts...,
	)
}

func (h *sendingHandler) Claim(ctx context.Context) (coredata.Email, error) {
	var email coredata.Email

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := email.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			email.Status = coredata.EmailStatusProcessing
			email.ProcessingStartedAt = &now
			email.AttemptCount++
			email.LastAttemptedAt = &now
			email.UpdatedAt = now

			if err := email.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update email: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoUnsentEmail) {
			return coredata.Email{}, worker.ErrNoTask
		}
		return coredata.Email{}, err
	}

	return email, nil
}

func (h *sendingHandler) Process(ctx context.Context, email coredata.Email) error {
	if sendErr := h.sendAndCommit(ctx, &email); sendErr != nil {
		if failErr := h.failEmail(ctx, &email, sendErr); failErr != nil {
			h.logger.ErrorCtx(ctx, "cannot fail email", log.Error(failErr))
		}
		return sendErr
	}
	return nil
}

func (h *sendingHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingEmails(ctx, conn, h.staleAfter)
		},
	)
}

func (h *sendingHandler) sendAndCommit(
	ctx context.Context,
	email *coredata.Email,
) error {
	var buf bytes.Buffer

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var attachments coredata.EmailAttachments
			if err := attachments.LoadByEmailID(ctx, conn, email.ID); err != nil {
				return fmt.Errorf("cannot load email attachments: %w", err)
			}

			fromName := h.senderName
			if email.SenderName != nil {
				fromName = *email.SenderName + " via " + h.senderName
			}

			mail := enmime.Builder().
				Subject(email.Subject).
				From(fromName, h.senderEmail).
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

				data, err := h.fileManager.GetFileBytes(ctx, &file)
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

	sendCtx, cancel := context.WithTimeout(ctx, h.smtpTimeout)
	defer cancel()

	if err := h.sendMail(sendCtx, []string{email.RecipientEmail}, buf.Bytes()); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("email sending timed out after %s: %w", h.smtpTimeout, err)
		}
		return fmt.Errorf("cannot send email: %w", err)
	}

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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
		h.logger.ErrorCtx(ctx,
			"email sent but failed to commit status update; will not re-queue to avoid duplicate delivery",
			log.Error(err),
			log.String("email_id", email.ID.String()),
		)
	}

	return nil
}

func (h *sendingHandler) failEmail(
	ctx context.Context,
	email *coredata.Email,
	processingError error,
) error {
	h.logger.ErrorCtx(
		ctx,
		"sending worker failure",
		log.Error(processingError),
		log.String("email_id", email.ID.String()),
	)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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

func (h *sendingHandler) sendMail(ctx context.Context, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(h.smtp.Addr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", h.smtp.Addr)
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

	if h.smtp.TLSRequired {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("TLS negotiation error: %w", err)
		}
	}

	if h.smtp.User != "" && h.smtp.Password != "" {
		auth := smtp.PlainAuth("", h.smtp.User, h.smtp.Password, host)
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication error: %w", err)
		}
	}

	if err = c.Mail(h.senderEmail); err != nil {
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
