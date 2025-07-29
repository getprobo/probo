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

package probod

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/agents"
	"github.com/getprobo/probo/pkg/awsconfig"
	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/crypto/passwdhash"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/mailer"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	"github.com/getprobo/probo/pkg/server"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	"github.com/getprobo/probo/pkg/trust"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/migrator"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/unit"
	"go.opentelemetry.io/otel/trace"
)

type (
	Implm struct {
		cfg config
	}

	config struct {
		Hostname      string               `json:"hostname"`
		HTTPS         bool                 `json:"https"`
		EncryptionKey cipher.EncryptionKey `json:"encryption-key"`
		Pg            pgConfig             `json:"pg"`
		Api           apiConfig            `json:"api"`
		Auth          authConfig           `json:"auth"`
		AWS           awsConfig            `json:"aws"`
		Mailer        mailerConfig         `json:"mailer"`
		Connectors    []connectorConfig    `json:"connectors"`
		OpenAI        openaiConfig         `json:"openai"`
		ChromeDPAddr  string               `json:"chrome-dp-addr"`
	}
)

var (
	_ unit.Configurable = (*Implm)(nil)
	_ unit.Runnable     = (*Implm)(nil)
)

func New() *Implm {
	return &Implm{
		cfg: config{
			Hostname: "localhost:8080",
			HTTPS:    true,
			Api: apiConfig{
				Addr: "localhost:8080",
			},
			Pg: pgConfig{
				Addr:     "localhost:5432",
				Username: "probod",
				Password: "probod",
				Database: "probod",
				PoolSize: 100,
			},
			ChromeDPAddr: "localhost:9222",
			Auth: authConfig{
				Password: passwordConfig{
					Pepper:     "this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes",
					Iterations: 1000000,
				},
				Cookie: cookieConfig{
					Name:     "SSID",
					Secret:   "this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes",
					Duration: 24,
					Domain:   "localhost",
				},
				DisableSignup: false,
			},
			AWS: awsConfig{
				Region: "us-east-1",
				Bucket: "probod",
			},
			Mailer: mailerConfig{
				SenderEmail: "no-reply@notification.getprobo.com",
				SenderName:  "Probo",
				SMTP: smtpConfig{
					Addr: "localhost:1025",
				},
			},
		},
	}
}

func (impl *Implm) GetConfiguration() any {
	return &impl.cfg
}

func (impl *Implm) Run(
	parentCtx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
) error {
	tracer := tp.Tracer("probod")
	ctx, rootSpan := tracer.Start(parentCtx, "probod.Run")
	defer rootSpan.End()

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	pgClient, err := pg.NewClient(
		impl.cfg.Pg.Options(
			pg.WithLogger(l),
			pg.WithRegisterer(r),
			pg.WithTracerProvider(tp),
		)...,
	)
	if err != nil {
		rootSpan.RecordError(err)
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	pepper, err := impl.cfg.Auth.GetPepperBytes()
	if err != nil {
		rootSpan.RecordError(err)
		return fmt.Errorf("cannot get pepper bytes: %w", err)
	}

	_, err = impl.cfg.Auth.GetCookieSecretBytes()
	if err != nil {
		rootSpan.RecordError(err)
		return fmt.Errorf("cannot get cookie secret bytes: %w", err)
	}

	awsConfig := awsconfig.NewConfig(
		l,
		httpclient.DefaultPooledClient(
			httpclient.WithLogger(l),
			httpclient.WithTracerProvider(tp),
			httpclient.WithRegisterer(r),
		),
		awsconfig.Options{
			Region:          impl.cfg.AWS.Region,
			AccessKeyID:     impl.cfg.AWS.AccessKeyID,
			SecretAccessKey: impl.cfg.AWS.SecretAccessKey,
			Endpoint:        impl.cfg.AWS.Endpoint,
		},
	)

	html2pdfConverter := html2pdf.NewConverter(
		impl.cfg.ChromeDPAddr,
		html2pdf.WithLogger(l),
		html2pdf.WithTracerProvider(tp),
	)

	s3Client := s3.NewFromConfig(awsConfig)

	err = migrator.NewMigrator(pgClient, coredata.Migrations, l.Named("migrations")).Run(ctx, "migrations")
	if err != nil {
		return fmt.Errorf("cannot migrate database schema: %w", err)
	}

	hp, err := passwdhash.NewProfile(pepper, uint32(impl.cfg.Auth.Password.Iterations))
	if err != nil {
		return fmt.Errorf("cannot create hashing profile: %w", err)
	}

	defaultConnectorRegistry := connector.NewConnectorRegistry()
	for _, connector := range impl.cfg.Connectors {
		if err := defaultConnectorRegistry.Register(connector.Name, connector.Config); err != nil {
			return fmt.Errorf("cannot register connector: %w", err)
		}
	}

	agentConfig := agents.Config{
		OpenAIAPIKey: impl.cfg.OpenAI.APIKey,
		Temperature:  impl.cfg.OpenAI.Temperature,
		ModelName:    impl.cfg.OpenAI.ModelName,
	}

	agent := agents.NewAgent(l.Named("agent"), agentConfig)

	usrmgrService, err := usrmgr.NewService(
		ctx,
		pgClient,
		hp,
		impl.cfg.Auth.Cookie.Secret,
		impl.cfg.Hostname,
		impl.cfg.Auth.DisableSignup,
	)
	if err != nil {
		return fmt.Errorf("cannot create usrmgr service: %w", err)
	}

	proboService, err := probo.NewService(
		ctx,
		impl.cfg.EncryptionKey,
		pgClient,
		s3Client,
		impl.cfg.AWS.Bucket,
		impl.cfg.Hostname,
		impl.cfg.HTTPS,
		impl.cfg.Auth.Cookie.Secret,
		agentConfig,
		html2pdfConverter,
		usrmgrService,
	)
	if err != nil {
		return fmt.Errorf("cannot create probo service: %w", err)
	}

	trustService := trust.NewService(
		pgClient,
		impl.cfg.AWS.Bucket,
	)

	serverHandler, err := server.NewServer(
		server.Config{
			AllowedOrigins:    impl.cfg.Api.Cors.AllowedOrigins,
			ExtraHeaderFields: impl.cfg.Api.ExtraHeaderFields,
			Probo:             proboService,
			Usrmgr:            usrmgrService,
			Trust:             trustService,
			ConnectorRegistry: defaultConnectorRegistry,
			Agent:             agent,
			SafeRedirect:      &saferedirect.SafeRedirect{AllowedHost: impl.cfg.Hostname},
			Logger:            l.Named("http.server"),
			Auth: console_v1.AuthConfig{
				CookieName:      impl.cfg.Auth.Cookie.Name,
				CookieDomain:    impl.cfg.Auth.Cookie.Domain,
				SessionDuration: time.Duration(impl.cfg.Auth.Cookie.Duration) * time.Hour,
				CookieSecret:    impl.cfg.Auth.Cookie.Secret,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	apiServerCtx, stopApiServer := context.WithCancel(context.Background())
	defer stopApiServer()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := impl.runApiServer(apiServerCtx, l, r, tp, serverHandler); err != nil {
			cancel(fmt.Errorf("api server crashed: %w", err))
		}
	}()

	mailerCtx, stopMailer := context.WithCancel(context.Background())
	mailer := mailer.NewMailer(pgClient, l, mailer.Config{
		SenderEmail: impl.cfg.Mailer.SenderEmail,
		SenderName:  impl.cfg.Mailer.SenderName,
		Addr:        impl.cfg.Mailer.SMTP.Addr,
		User:        impl.cfg.Mailer.SMTP.User,
		Password:    impl.cfg.Mailer.SMTP.Password,
		TLSRequired: impl.cfg.Mailer.SMTP.TLSRequired,
		Timeout:     time.Second * 10,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mailer.Run(mailerCtx); err != nil {
			cancel(fmt.Errorf("mailer crashed: %w", err))
		}
	}()

	<-ctx.Done()

	stopMailer()
	stopApiServer()

	wg.Wait()

	pgClient.Close()

	return context.Cause(ctx)
}

func (impl *Implm) runApiServer(
	ctx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
	handler http.Handler,
) error {
	tracer := tp.Tracer("github.com/getprobo/probo/pkg/probod")
	ctx, span := tracer.Start(ctx, "probod.runApiServer")
	defer span.End()

	apiServer := httpserver.NewServer(
		impl.cfg.Api.Addr,
		handler,
		httpserver.WithLogger(l),
		httpserver.WithRegisterer(r),
		httpserver.WithTracerProvider(tp),
	)

	l.Info("starting api server", log.String("addr", apiServer.Addr))
	span.AddEvent("API server starting")

	listener, err := net.Listen("tcp", apiServer.Addr)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("cannot listen on %q: %w", apiServer.Addr, err)
	}
	defer listener.Close()

	serverErrCh := make(chan error, 1)
	go func() {
		err := apiServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- fmt.Errorf("cannot server http request: %w", err)
		}
		close(serverErrCh)
	}()

	l.Info("api server started")
	span.AddEvent("API server started")

	select {
	case err := <-serverErrCh:
		if err != nil {
			span.RecordError(err)
		}
		return err
	case <-ctx.Done():
	}

	l.InfoCtx(ctx, "shutting down api server")
	span.AddEvent("API server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("cannot shutdown api server: %w", err)
	}

	span.AddEvent("API server shutdown complete")
	return ctx.Err()
}
