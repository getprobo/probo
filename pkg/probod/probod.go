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
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/migrator"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/unit"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/agents"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/awsconfig"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/keys"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/crypto/pem"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/mailer"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/server"
	"go.probo.inc/probo/pkg/server/api"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/trust"
	"golang.org/x/sync/errgroup"
)

type (
	Implm struct {
		cfg config
	}

	config struct {
		BaseURL       *baseurl.BaseURL     `json:"base-url"`
		EncryptionKey cipher.EncryptionKey `json:"encryption-key"`
		Pg            pgConfig             `json:"pg"`
		Api           apiConfig            `json:"api"`
		Auth          authConfig           `json:"auth"`
		TrustAuth     trustAuthConfig      `json:"trust-auth"`
		TrustCenter   trustCenterConfig    `json:"trust-center"`
		AWS           awsConfig            `json:"aws"`
		Notifications notificationsConfig  `json:"notifications"`
		Connectors    []connectorConfig    `json:"connectors"`
		OpenAI        openaiConfig         `json:"openai"`
		ChromeDPAddr  string               `json:"chrome-dp-addr"`
		CustomDomains customDomainsConfig  `json:"custom-domains"`
	}

	trustCenterConfig struct {
		HTTPAddr  string `json:"http-addr"`
		HTTPSAddr string `json:"https-addr"`
	}
)

var (
	_ unit.Configurable = (*Implm)(nil)
	_ unit.Runnable     = (*Implm)(nil)
)

func New() *Implm {
	return &Implm{
		cfg: config{
			BaseURL: baseurl.MustParse("http://localhost:8080"),
			Api: apiConfig{
				Addr: "localhost:8080",
			},
			Pg: pgConfig{
				Addr:     "localhost:5432",
				Username: "postgres",
				Password: "postgres",
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
					Secure:   true,
				},
				DisableSignup:                       false,
				InvitationConfirmationTokenValidity: 3600,
				SAML: samlConfig{
					SessionDuration:        604800,
					CleanupIntervalSeconds: 86400,
				},
			},
			TrustAuth: trustAuthConfig{
				CookieName:        "TCT",
				CookieDomain:      "localhost",
				CookieDuration:    24,
				TokenDuration:     720,
				ReportURLDuration: 15,
				TokenSecret:       "this-is-a-secure-secret-for-trust-token-signing-at-least-32-bytes",
				Scope:             "trust_center_readonly",
				TokenType:         "trust_center_access",
			},
			TrustCenter: trustCenterConfig{
				HTTPAddr:  ":80",
				HTTPSAddr: ":443",
			},
			AWS: awsConfig{
				Region: "us-east-1",
				Bucket: "probod",
			},
			Notifications: notificationsConfig{
				Mailer: mailerConfig{
					MailerInterval: 60,
					SenderEmail:    "no-reply@notification.getprobo.com",
					SenderName:     "Probo",
					SMTP: smtpConfig{
						Addr: "localhost:1025",
					},
				},
				Slack: slackConfig{
					SenderInterval: 60,
				},
			},
			CustomDomains: customDomainsConfig{
				RenewalInterval:   3600,
				ProvisionInterval: 30,
				ACME: acmeConfig{
					Directory: "https://acme-v02.api.letsencrypt.org/directory",
					Email:     "admin@getprobo.com",
					KeyType:   "EC256",
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

	_, err = impl.cfg.TrustAuth.GetTokenSecretBytes()
	if err != nil {
		rootSpan.RecordError(err)
		return fmt.Errorf("cannot get trust auth token secret bytes: %w", err)
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
		if err := defaultConnectorRegistry.Register(connector.Provider, connector.Config); err != nil {
			return fmt.Errorf("cannot register connector: %w", err)
		}
	}

	agentConfig := agents.Config{
		OpenAIAPIKey: impl.cfg.OpenAI.APIKey,
		Temperature:  impl.cfg.OpenAI.Temperature,
		ModelName:    impl.cfg.OpenAI.ModelName,
	}

	trustConfig := probo.TrustConfig{
		TokenSecret:   impl.cfg.TrustAuth.TokenSecret,
		TokenDuration: time.Duration(impl.cfg.TrustAuth.TokenDuration) * time.Hour,
		TokenType:     impl.cfg.TrustAuth.TokenType,
	}

	agent := agents.NewAgent(l.Named("agent"), agentConfig)

	authService, err := auth.NewService(
		ctx,
		pgClient,
		impl.cfg.EncryptionKey,
		hp,
		impl.cfg.Auth.Cookie.Secret,
		impl.cfg.BaseURL.String(),
		impl.cfg.Auth.DisableSignup,
		time.Duration(impl.cfg.Auth.InvitationConfirmationTokenValidity)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("cannot create auth service: %w", err)
	}

	authzService, err := authz.NewService(
		ctx,
		pgClient,
		impl.cfg.BaseURL.String(),
		impl.cfg.Auth.Cookie.Secret,
		time.Duration(impl.cfg.Auth.InvitationConfirmationTokenValidity)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("cannot create authz service: %w", err)
	}

	fileManagerService := filemanager.NewService(s3Client)

	samlService, err := auth.NewSAMLService(
		pgClient,
		impl.cfg.EncryptionKey,
		impl.cfg.BaseURL.String(),
		impl.cfg.Auth.SAML.SessionDurationTime(),
		impl.cfg.Auth.Cookie.Name,
		impl.cfg.Auth.Cookie.Secret,
		impl.cfg.Auth.SAML.Certificate,
		impl.cfg.Auth.SAML.PrivateKey,
		l.Named("saml"),
	)
	if err != nil {
		return fmt.Errorf("cannot create SAML service: %w", err)
	}

	var accountKey crypto.Signer
	if impl.cfg.CustomDomains.ACME.AccountKey != "" {
		accountKey, err = pem.DecodePrivateKey([]byte(impl.cfg.CustomDomains.ACME.AccountKey))
		if err != nil {
			return fmt.Errorf("cannot decode ACME account key: %w", err)
		}
		l.Info("using configured ACME account key")
	}

	var rootCAs *x509.CertPool
	if impl.cfg.CustomDomains.ACME.RootCA != "" {
		rootCAs = x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM([]byte(impl.cfg.CustomDomains.ACME.RootCA)) {
			return fmt.Errorf("cannot parse ACME root CA certificate")
		}
	}

	acmeService, err := certmanager.NewACMEService(
		impl.cfg.CustomDomains.ACME.Email,
		keys.Type(impl.cfg.CustomDomains.ACME.KeyType),
		impl.cfg.CustomDomains.ACME.Directory,
		accountKey,
		rootCAs,
		l,
	)
	if err != nil {
		return fmt.Errorf("cannot initialize ACME service: %w", err)
	}

	proboService, err := probo.NewService(
		ctx,
		impl.cfg.EncryptionKey,
		pgClient,
		s3Client,
		impl.cfg.AWS.Bucket,
		impl.cfg.BaseURL.String(),
		impl.cfg.Auth.Cookie.Secret,
		trustConfig,
		agentConfig,
		html2pdfConverter,
		acmeService,
		fileManagerService,
		authService,
		authzService,
		l.Named("probo"),
	)
	if err != nil {
		return fmt.Errorf("cannot create probo service: %w", err)
	}

	trustService := trust.NewService(
		pgClient,
		s3Client,
		impl.cfg.AWS.Bucket,
		impl.cfg.BaseURL.String(),
		impl.cfg.EncryptionKey,
		impl.cfg.TrustAuth.TokenSecret,
		impl.cfg.GetSlackSigningSecret(),
		authService,
		html2pdfConverter,
		fileManagerService,
		l,
		trust.TrustConfig{
			TokenSecret:   impl.cfg.TrustAuth.TokenSecret,
			TokenDuration: time.Duration(impl.cfg.TrustAuth.TokenDuration) * time.Hour,
			TokenType:     impl.cfg.TrustAuth.TokenType,
		},
	)

	serverHandler, err := server.NewServer(
		server.Config{
			AllowedOrigins:    impl.cfg.Api.Cors.AllowedOrigins,
			ExtraHeaderFields: impl.cfg.Api.ExtraHeaderFields,
			Probo:             proboService,
			Auth:              authService,
			Authz:             authzService,
			Trust:             trustService,
			SAML:              samlService,
			ConnectorRegistry: defaultConnectorRegistry,
			Agent:             agent,
			SafeRedirect:      &saferedirect.SafeRedirect{AllowedHost: impl.cfg.BaseURL.Host()},
			CustomDomainCname: impl.cfg.CustomDomains.CnameTarget,
			FileManager:       fileManagerService,
			PGClient:          pgClient,
			Logger:            l.Named("http.server"),
			ConsoleAuth: api.ConsoleAuthConfig{
				CookieName:      impl.cfg.Auth.Cookie.Name,
				CookieDomain:    impl.cfg.Auth.Cookie.Domain,
				SessionDuration: time.Duration(impl.cfg.Auth.Cookie.Duration) * time.Hour,
				CookieSecret:    impl.cfg.Auth.Cookie.Secret,
				CookieSecure:    impl.cfg.Auth.Cookie.Secure,
			},
			TrustAuth: api.TrustAuthConfig{
				CookieName:        impl.cfg.TrustAuth.CookieName,
				CookieDomain:      impl.cfg.TrustAuth.CookieDomain,
				CookieDuration:    time.Duration(impl.cfg.TrustAuth.CookieDuration) * time.Hour,
				TokenDuration:     time.Duration(impl.cfg.TrustAuth.TokenDuration) * time.Hour,
				ReportURLDuration: time.Duration(impl.cfg.TrustAuth.ReportURLDuration) * time.Minute,
				TokenSecret:       impl.cfg.TrustAuth.TokenSecret,
				Scope:             impl.cfg.TrustAuth.Scope,
				TokenType:         impl.cfg.TrustAuth.TokenType,
				CookieSecure:      impl.cfg.Auth.Cookie.Secure,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	apiServerCtx, stopApiServer := context.WithCancel(context.Background())
	defer stopApiServer()
	wg.Go(
		func() {
			if err := impl.runApiServer(apiServerCtx, l, r, tp, serverHandler); err != nil {
				cancel(fmt.Errorf("api server crashed: %w", err))
			}
		},
	)

	mailerCtx, stopMailer := context.WithCancel(context.Background())
	mailer := mailer.NewMailer(pgClient, l, mailer.Config{
		SenderEmail: impl.cfg.Notifications.Mailer.SenderEmail,
		SenderName:  impl.cfg.Notifications.Mailer.SenderName,
		Addr:        impl.cfg.Notifications.Mailer.SMTP.Addr,
		User:        impl.cfg.Notifications.Mailer.SMTP.User,
		Password:    impl.cfg.Notifications.Mailer.SMTP.Password,
		TLSRequired: impl.cfg.Notifications.Mailer.SMTP.TLSRequired,
		Timeout:     time.Second * 10,
		Interval:    time.Duration(impl.cfg.Notifications.Mailer.MailerInterval) * time.Second,
	})
	wg.Go(
		func() {
			if err := mailer.Run(mailerCtx); err != nil {
				cancel(fmt.Errorf("mailer crashed: %w", err))
			}
		},
	)

	slackSenderCtx, stopSlackSender := context.WithCancel(context.Background())
	slackSender := slack.NewSender(pgClient, l.Named("slack-sender"), impl.cfg.EncryptionKey, slack.Config{
		Interval: time.Duration(impl.cfg.Notifications.Slack.SenderInterval) * time.Second,
	})
	wg.Go(
		func() {
			if err := slackSender.Run(slackSenderCtx); err != nil {
				cancel(fmt.Errorf("slack sender crashed: %w", err))
			}
		},
	)

	exportJobExporterCtx, stopExportJobExporter := context.WithCancel(context.Background())
	wg.Go(
		func() {
			if err := impl.runExportJob(exportJobExporterCtx, proboService, l.Named("export-job-exporter")); err != nil {
				cancel(fmt.Errorf("export job exporter crashed: %w", err))
			}
		},
	)

	samlCleanerCtx, stopSAMLCleaner := context.WithCancel(context.Background())
	samlCleaner := auth.NewCleaner(
		pgClient,
		impl.cfg.Auth.SAML.CleanupInterval(),
		l.Named("saml-cleaner"),
	)
	wg.Go(
		func() {
			if err := samlCleaner.Run(samlCleanerCtx); err != nil {
				cancel(fmt.Errorf("saml cleaner crashed: %w", err))
			}
		},
	)

	trustCenterServerCtx, stopTrustCenterServer := context.WithCancel(context.Background())
	defer stopTrustCenterServer()
	wg.Go(
		func() {
			if err := impl.runTrustCenterServer(trustCenterServerCtx, l, r, tp, pgClient, serverHandler.TrustCenterHandler(), acmeService); err != nil {
				cancel(fmt.Errorf("trust center server crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopMailer()
	stopSlackSender()
	stopExportJobExporter()
	stopSAMLCleaner()
	stopApiServer()
	stopTrustCenterServer()

	wg.Wait()

	pgClient.Close()

	return context.Cause(ctx)
}

func (impl *Implm) runExportJob(
	ctx context.Context,
	proboService *probo.Service,
	l *log.Logger,
) error {
LOOP:
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		if err := proboService.ExportJob(ctx); err != nil {
			if !errors.Is(err, coredata.ErrNoExportJobAvailable) {
				l.ErrorCtx(ctx, "cannot process export job", log.Error(err))
			}
		}

		goto LOOP
	}
}

func (impl *Implm) runApiServer(
	ctx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
	handler http.Handler,
) error {
	tracer := tp.Tracer("go.probo.inc/probo/pkg/probod")
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

func (impl *Implm) runTrustCenterServer(
	ctx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
	pgClient *pg.Client,
	trustRouter http.Handler,
	acmeService *certmanager.ACMEService,
) error {
	tracer := tp.Tracer("go.probo.inc/probo/pkg/probod")
	ctx, span := tracer.Start(ctx, "probod.runTrustCenterServer")
	defer span.End()

	certSelector := certmanager.NewSelector(pgClient, impl.cfg.EncryptionKey)

	warmer := certmanager.NewCacheStore(pgClient, impl.cfg.EncryptionKey, l)
	if err := warmer.WarmCache(ctx); err != nil {
		span.RecordError(err)
		l.ErrorCtx(ctx, "cannot warm certificate cache", log.Error(err))
	}

	renewalInterval := time.Duration(impl.cfg.CustomDomains.RenewalInterval) * time.Second
	if renewalInterval == 0 {
		renewalInterval = time.Hour
	}

	renewer := certmanager.NewRenewer(pgClient, acmeService, impl.cfg.EncryptionKey, renewalInterval, l)

	certProvisioningInterval := time.Duration(impl.cfg.CustomDomains.ProvisionInterval) * time.Second
	if certProvisioningInterval == 0 {
		certProvisioningInterval = 30 * time.Second
	}
	certProvisioner := certmanager.NewProvisioner(pgClient, acmeService, impl.cfg.EncryptionKey, certProvisioningInterval, l)

	g, ctx := errgroup.WithContext(ctx)

	l.Info("starting trust center services")
	span.AddEvent("Trust center services starting")

	g.Go(
		func() error {
			l.Info("starting certificate renewer")
			return renewer.Run(ctx)
		},
	)

	g.Go(
		func() error {
			l.Info("starting certificate provisioner")
			return certProvisioner.Run(ctx)
		},
	)

	httpACMEHandler := certmanager.NewACMEChallengeHandler(
		pgClient,
		impl.cfg.EncryptionKey,
		l.Named("http_acme_handler"),
	)

	httpServer := httpserver.NewServer(
		impl.cfg.TrustCenter.HTTPAddr,
		httpACMEHandler.Handle(http.NotFoundHandler()),
		httpserver.WithLogger(l),
		httpserver.WithRegisterer(r),
		httpserver.WithTracerProvider(tp),
	)

	g.Go(
		func() error {
			l.InfoCtx(ctx, "starting HTTP server for ACME challenges", log.String("addr", httpServer.Addr))
			span.AddEvent("HTTP server starting")

			listener, err := net.Listen("tcp", httpServer.Addr)
			if err != nil {
				return fmt.Errorf("cannot listen on %q: %w", httpServer.Addr, err)
			}
			defer listener.Close()

			if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("cannot serve http requests: %w", err)
			}
			return nil
		},
	)

	acmeHandler := certmanager.NewACMEChallengeHandler(
		pgClient,
		impl.cfg.EncryptionKey,
		l.Named("acme_handler"),
	)

	handler := acmeHandler.Handle(trustRouter)

	httpsServer := httpserver.NewServer(
		impl.cfg.TrustCenter.HTTPSAddr,
		handler,
		httpserver.WithLogger(l),
		httpserver.WithRegisterer(r),
		httpserver.WithTracerProvider(tp),
	)

	httpsServer.TLSConfig = &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := certSelector.GetCertificate(hello)
			// Silently reject connections without SNI (load balancers, health checks, scanners)
			if err != nil {
				var noSNIErr *certmanager.NoSNIError
				if errors.As(err, &noSNIErr) {
					return nil, nil
				}
			}
			return cert, err
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
	httpsServer.ReadTimeout = 30 * time.Second
	httpsServer.WriteTimeout = 30 * time.Second

	g.Go(
		func() error {
			l.InfoCtx(ctx, "starting trust center https server", log.String("addr", httpsServer.Addr))
			span.AddEvent("HTTPS server starting")

			listener, err := net.Listen("tcp", httpsServer.Addr)
			if err != nil {
				return fmt.Errorf("cannot listen on %q: %w", httpsServer.Addr, err)
			}
			defer listener.Close()

			if err := httpsServer.ServeTLS(listener, "", ""); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("cannot serve https requests: %w", err)
			}

			return nil
		},
	)

	l.Info("trust center servers started")
	span.AddEvent("Trust center servers started")

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		l.InfoCtx(ctx, "shutting down trust center servers...")
		span.AddEvent("Trust center servers shutting down")

		if err := httpsServer.Shutdown(shutdownCtx); err != nil {
			span.RecordError(err)
			l.ErrorCtx(ctx, "cannot shutdown HTTPS server", log.Error(err))
		}

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			span.RecordError(err)
			l.ErrorCtx(ctx, "cannot shutdown HTTP server", log.Error(err))
		}

		span.AddEvent("Trust center servers shutdown complete")
	}()

	if err := g.Wait(); err != nil {
		span.RecordError(err)
		return err
	}

	return ctx.Err()
}
