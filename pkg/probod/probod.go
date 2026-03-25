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

package probod

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.probo.inc/probo/packages/emails"
	pemutil "go.probo.inc/probo/pkg/crypto/pem"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	proxyproto "github.com/pires/go-proxyproto"
	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/migrator"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/unit"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/awsconfig"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/keys"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/file"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oidc"
	"go.probo.inc/probo/pkg/mailer"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/trust"
	"go.probo.inc/probo/pkg/webhook"
	"golang.org/x/sync/errgroup"
)

type (
	Implm struct {
		cfg Config
	}

	// FullConfig represents the complete configuration file structure.
	// This is used by bootstrap to generate the YAML config file.
	FullConfig struct {
		Unit   UnitConfig `json:"unit"`
		Probod Config     `json:"probod"`
	}

	// UnitConfig contains unit framework configuration.
	UnitConfig struct {
		Metrics MetricsConfig `json:"metrics"`
		Tracing TracingConfig `json:"tracing"`
	}

	// MetricsConfig contains metrics server configuration.
	MetricsConfig struct {
		Addr string `json:"addr"`
	}

	// TracingConfig contains tracing configuration.
	TracingConfig struct {
		Addr          string `json:"addr"`
		MaxBatchSize  int    `json:"max-batch-size"`
		BatchTimeout  int    `json:"batch-timeout"`
		ExportTimeout int    `json:"export-timeout"`
		MaxQueueSize  int    `json:"max-queue-size"`
	}

	// ESignConfig contains electronic signature configuration.
	ESignConfig struct {
		TSAURL string `json:"tsa-url"`
	}

	// Config represents the probod application configuration.
	Config struct {
		BaseURL       string              `json:"base-url"`
		EncryptionKey string              `json:"encryption-key"`
		Pg            PgConfig            `json:"pg"`
		Api           APIConfig           `json:"api"`
		Auth          AuthConfig          `json:"auth"`
		TrustCenter   TrustCenterConfig   `json:"trust-center"`
		AWS           AWSConfig           `json:"aws"`
		Notifications NotificationsConfig `json:"notifications"`
		Connectors    []ConnectorConfig   `json:"connectors"`
		OpenAI        OpenAIConfig        `json:"openai"`
		ChromeDPAddr  string              `json:"chrome-dp-addr"`
		CustomDomains CustomDomainsConfig `json:"custom-domains"`
		SCIMBridge    SCIMBridgeConfig    `json:"scim-bridge"`
		ESign         ESignConfig         `json:"esign"`
	}

	// TrustCenterConfig contains trust center server configuration.
	TrustCenterConfig struct {
		HTTPAddr      string              `json:"http-addr"`
		HTTPSAddr     string              `json:"https-addr"`
		ProxyProtocol ProxyProtocolConfig `json:"proxy-protocol"`
	}
)

var (
	_ unit.Configurable = (*Implm)(nil)
	_ unit.Runnable     = (*Implm)(nil)
)

func New() *Implm {
	return &Implm{
		cfg: Config{
			BaseURL: "http://localhost:8080",
			Api: APIConfig{
				Addr: "localhost:8080",
			},
			Pg: PgConfig{
				Addr:     "localhost:5432",
				Username: "postgres",
				Password: "postgres",
				Database: "probod",
				PoolSize: 100,
			},
			ChromeDPAddr: "localhost:9222",
			Auth: AuthConfig{
				Password: PasswordConfig{
					Pepper:     "this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes",
					Iterations: 1000000,
				},
				Cookie: CookieConfig{
					Name:     "SSID",
					Secret:   "this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes",
					Duration: 24,
					Domain:   "localhost",
					Secure:   true,
				},
				DisableSignup:                       false,
				InvitationConfirmationTokenValidity: 3600,
				PasswordResetTokenValidity:          3600,
				MagicLinkTokenValidity:              900,
				SAML: SAMLConfig{
					SessionDuration:                   604800,
					CleanupIntervalSeconds:            86400,
					DomainVerificationIntervalSeconds: 60,
					DomainVerificationResolverAddr:    "8.8.8.8:53",
				},
			},
			TrustCenter: TrustCenterConfig{
				HTTPAddr:  ":80",
				HTTPSAddr: ":443",
			},
			AWS: AWSConfig{
				Region: "us-east-1",
				Bucket: "probod",
			},
			Notifications: NotificationsConfig{
				Mailer: MailerConfig{
					MailerInterval: 60,
					SenderEmail:    "no-reply@notification.getprobo.com",
					SenderName:     "Probo",
					SMTP: SMTPConfig{
						Addr: "localhost:1025",
					},
				},
				Slack: SlackConfig{
					SenderInterval: 60,
				},
				Webhook: WebhookConfig{
					SenderInterval: 5,
					CacheTTL:       86400,
				},
				SigningNotificationInterval: 600,
			},
			CustomDomains: CustomDomainsConfig{
				RenewalInterval:   3600,
				ProvisionInterval: 30,
				ResolverAddr:      "8.8.8.8:53",
				ACME: ACMEConfig{
					Directory: "https://acme-v02.api.letsencrypt.org/directory",
					Email:     "admin@getprobo.com",
					KeyType:   "EC256",
				},
			},
			SCIMBridge: SCIMBridgeConfig{
				SyncInterval: 60, // 15 minutes
				PollInterval: 30, // 30 seconds
			},
			ESign: ESignConfig{
				TSAURL: "http://timestamp.digicert.com",
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

	// Parse config values that need conversion from strings to complex types
	baseURL, err := baseurl.Parse(impl.cfg.BaseURL)
	if err != nil {
		rootSpan.RecordError(err)
		return fmt.Errorf("cannot parse base URL: %w", err)
	}

	var encryptionKey cipher.EncryptionKey
	if err := encryptionKey.UnmarshalText([]byte(impl.cfg.EncryptionKey)); err != nil {
		rootSpan.RecordError(err)
		return fmt.Errorf("cannot parse encryption key: %w", err)
	}

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

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = impl.cfg.AWS.UsePathStyle
	})

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

	llmClient, err := buildLLMClient(impl.cfg.OpenAI, l.Named("llm"), tp, r)
	if err != nil {
		return fmt.Errorf("cannot create LLM client: %w", err)
	}

	fileManagerService := filemanager.NewService(s3Client)

	var samlCert *x509.Certificate
	var samlKey *rsa.PrivateKey
	if impl.cfg.Auth.SAML.Certificate != "" && impl.cfg.Auth.SAML.PrivateKey != "" {
		// Decode certificate
		certBlock, _ := pem.Decode([]byte(impl.cfg.Auth.SAML.Certificate))
		if certBlock == nil {
			return fmt.Errorf("cannot decode SAML certificate PEM block")
		}
		var err error
		samlCert, err = x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return fmt.Errorf("cannot parse SAML certificate: %w", err)
		}

		// Decode private key
		signer, err := pemutil.DecodePrivateKey([]byte(impl.cfg.Auth.SAML.PrivateKey))
		if err != nil {
			return fmt.Errorf("cannot decode SAML private key: %w", err)
		}
		var ok bool
		samlKey, ok = signer.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("SAML private key is not an RSA key")
		}
	}

	if err := emails.UploadStaticAssets(
		ctx,
		s3Client,
		impl.cfg.AWS.Bucket,
	); err != nil {
		return fmt.Errorf("cannot upload email static assets: %w", err)
	}

	iamService, err := iam.NewService(
		ctx,
		pgClient,
		fileManagerService,
		hp,
		iam.Config{
			DisableSignup:                  impl.cfg.Auth.DisableSignup,
			InvitationTokenValidity:        time.Duration(impl.cfg.Auth.InvitationConfirmationTokenValidity) * time.Second,
			PasswordResetTokenValidity:     time.Duration(impl.cfg.Auth.PasswordResetTokenValidity) * time.Second,
			MagicLinkTokenValidity:         time.Duration(impl.cfg.Auth.MagicLinkTokenValidity) * time.Second,
			SessionDuration:                time.Duration(impl.cfg.Auth.Cookie.Duration) * time.Hour,
			Bucket:                         impl.cfg.AWS.Bucket,
			TokenSecret:                    impl.cfg.Auth.Cookie.Secret,
			BaseURL:                        baseURL,
			EncryptionKey:                  encryptionKey,
			Certificate:                    samlCert,
			PrivateKey:                     samlKey,
			Logger:                         l.Named("iam"),
			TracerProvider:                 tp,
			Registerer:                     r,
			ConnectorRegistry:              defaultConnectorRegistry,
			DomainVerificationInterval:     impl.cfg.Auth.SAML.DomainVerificationInterval(),
			DomainVerificationResolverAddr: impl.cfg.Auth.SAML.DomainVerificationResolverAddr,
			SCIMBridgeSyncInterval:         time.Duration(impl.cfg.SCIMBridge.SyncInterval) * time.Second,
			SCIMBridgePollInterval:         time.Duration(impl.cfg.SCIMBridge.PollInterval) * time.Second,
			GoogleOIDC: oidc.ProviderConfig{
				ClientID:     impl.cfg.Auth.Google.ClientID,
				ClientSecret: impl.cfg.Auth.Google.ClientSecret,
				Enabled:      impl.cfg.Auth.Google.Enabled,
			},
			MicrosoftOIDC: oidc.ProviderConfig{
				ClientID:     impl.cfg.Auth.Microsoft.ClientID,
				ClientSecret: impl.cfg.Auth.Microsoft.ClientSecret,
				Enabled:      impl.cfg.Auth.Microsoft.Enabled,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot create iam service: %w", err)
	}

	var accountKey crypto.Signer
	if impl.cfg.CustomDomains.ACME.AccountKey != "" {
		accountKey, err = pemutil.DecodePrivateKey([]byte(impl.cfg.CustomDomains.ACME.AccountKey))
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

	slackService := slack.NewService(
		pgClient,
		impl.cfg.GetSlackSigningSecret(),
		baseURL.String(),
		impl.cfg.Auth.Cookie.Secret,
		l.Named("slack"),
	)

	esignService := esign.NewService(
		pgClient,
		fileManagerService,
		html2pdfConverter,
		impl.cfg.ESign.TSAURL,
		impl.cfg.AWS.Bucket,
		l.Named("esign"),
	)

	mailmanService := mailman.NewService(pgClient, fileManagerService, impl.cfg.Auth.Cookie.Secret, baseURL, impl.cfg.AWS.Bucket, encryptionKey, l)

	proboService, err := probo.NewService(
		ctx,
		encryptionKey,
		pgClient,
		s3Client,
		impl.cfg.AWS.Bucket,
		baseURL.String(),
		impl.cfg.Auth.Cookie.Secret,
		llmClient,
		impl.cfg.OpenAI.ModelName,
		impl.cfg.OpenAI.Temperature,
		impl.cfg.OpenAI.MaxTokens,
		html2pdfConverter,
		acmeService,
		fileManagerService,
		l.Named("probo"),
		slackService,
		iamService,
		esignService,
		time.Duration(impl.cfg.Auth.InvitationConfirmationTokenValidity)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("cannot create probo service: %w", err)
	}

	trustService := trust.NewService(
		pgClient,
		s3Client,
		impl.cfg.AWS.Bucket,
		baseURL.String(),
		impl.cfg.GetSlackSigningSecret(),
		iamService,
		esignService,
		html2pdfConverter,
		fileManagerService,
		l,
		slackService,
	)

	fileService := file.NewService(pgClient, fileManagerService)

	serverHandler, err := server.NewServer(
		server.Config{
			AllowedOrigins:    impl.cfg.Api.Cors.AllowedOrigins,
			ExtraHeaderFields: impl.cfg.Api.ExtraHeaderFields,
			Probo:             proboService,
			File:              fileService,
			IAM:               iamService,
			Trust:             trustService,
			ESign:             esignService,
			Mailman:           mailmanService,
			Slack:             slackService,
			ConnectorRegistry: defaultConnectorRegistry,
			BaseURL:           baseURL,

			CustomDomainCname: impl.cfg.CustomDomains.CnameTarget,
			TokenSecret:       impl.cfg.Auth.Cookie.Secret,
			Logger:            l.Named("http.server"),
			Cookie: securecookie.Config{
				Name:     impl.cfg.Auth.Cookie.Name,
				Domain:   impl.cfg.Auth.Cookie.Domain,
				Path:     "/",
				MaxAge:   int(time.Duration(impl.cfg.Auth.Cookie.Duration) * time.Hour),
				Secret:   impl.cfg.Auth.Cookie.Secret,
				Secure:   impl.cfg.Auth.Cookie.Secure,
				HTTPOnly: true,
				SameSite: http.SameSiteLaxMode,
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
	sendingWorker := mailer.NewSendingWorker(
		pgClient,
		fileManagerService,
		impl.cfg.Notifications.Mailer.SenderName,
		impl.cfg.Notifications.Mailer.SenderEmail,
		mailer.SMTPConfig{
			Addr:        impl.cfg.Notifications.Mailer.SMTP.Addr,
			User:        impl.cfg.Notifications.Mailer.SMTP.User,
			Password:    impl.cfg.Notifications.Mailer.SMTP.Password,
			TLSRequired: impl.cfg.Notifications.Mailer.SMTP.TLSRequired,
		},
		l.Named("sending-worker"),
		mailer.WithSendingWorkerSMTPTimeout(time.Second*10),
		mailer.WithSendingWorkerInterval(time.Duration(impl.cfg.Notifications.Mailer.MailerInterval)*time.Second),
	)
	wg.Go(
		func() {
			if err := sendingWorker.Run(mailerCtx); err != nil {
				cancel(fmt.Errorf("sending worker crashed: %w", err))
			}
		},
	)

	slackSenderCtx, stopSlackSender := context.WithCancel(context.Background())
	slackSender := slack.NewSender(pgClient, l.Named("slack-sender"), encryptionKey, slack.Config{
		Interval: time.Duration(impl.cfg.Notifications.Slack.SenderInterval) * time.Second,
	})
	wg.Go(
		func() {
			if err := slackSender.Run(slackSenderCtx); err != nil {
				cancel(fmt.Errorf("slack sender crashed: %w", err))
			}
		},
	)

	webhookSenderCtx, stopWebhookSender := context.WithCancel(context.Background())
	webhookSender := webhook.NewSender(pgClient, l.Named("webhook-sender"), webhook.Config{
		Interval:      time.Duration(impl.cfg.Notifications.Webhook.SenderInterval) * time.Second,
		CacheTTL:      time.Duration(impl.cfg.Notifications.Webhook.CacheTTL) * time.Second,
		EncryptionKey: encryptionKey,
	})
	wg.Go(
		func() {
			if err := webhookSender.Run(webhookSenderCtx); err != nil {
				cancel(fmt.Errorf("webhook sender crashed: %w", err))
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

	iamServiceCtx, stopIAMService := context.WithCancel(context.Background())
	wg.Go(
		func() {
			if err := iamService.Run(iamServiceCtx); err != nil {
				cancel(fmt.Errorf("iam service crashed: %w", err))
			}
		},
	)

	esignServiceCtx, stopESignService := context.WithCancel(context.Background())
	wg.Go(
		func() {
			if err := esignService.Run(esignServiceCtx, trustService.EmailPresenterConfigByOrganizationID); err != nil {
				cancel(fmt.Errorf("esign service crashed: %w", err))
			}
		},
	)

	mailingListWorker := mailman.NewMailingListWorker(mailmanService, pgClient, l.Named("mailing-list-worker"))
	mailingListWorkerCtx, stopMailingListWorker := context.WithCancel(context.Background())
	wg.Go(
		func() {
			if err := mailingListWorker.Run(mailingListWorkerCtx); err != nil {
				cancel(fmt.Errorf("mailing list worker crashed: %w", err))
			}
		},
	)

	signingNotificationWorker := probo.NewSigningNotificationWorker(
		pgClient,
		fileManagerService,
		impl.cfg.AWS.Bucket,
		baseURL.String(),
		impl.cfg.Auth.Cookie.Secret,
		time.Duration(impl.cfg.Auth.InvitationConfirmationTokenValidity)*time.Second,
		l.Named("signing-notification-worker"),
		probo.WithSigningNotificationWorkerInterval(time.Duration(impl.cfg.Notifications.SigningNotificationInterval)*time.Second),
	)
	signingNotificationWorkerCtx, stopSigningNotificationWorker := context.WithCancel(context.Background())
	wg.Go(
		func() {
			if err := signingNotificationWorker.Run(signingNotificationWorkerCtx); err != nil {
				cancel(fmt.Errorf("signing notification worker crashed: %w", err))
			}
		},
	)

	trustCenterServerCtx, stopTrustCenterServer := context.WithCancel(context.Background())
	defer stopTrustCenterServer()
	wg.Go(
		func() {
			if err := impl.runTrustCenterServer(
				trustCenterServerCtx,
				l,
				r,
				tp,
				pgClient,
				serverHandler.TrustCenterHandler(),
				acmeService,
				proboService,
				encryptionKey,
			); err != nil {
				cancel(fmt.Errorf("trust center server crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopApiServer()
	stopTrustCenterServer()
	stopWebhookSender()
	stopESignService()
	stopMailingListWorker()
	stopSigningNotificationWorker()
	stopExportJobExporter()
	stopIAMService()
	stopMailer()
	stopSlackSender()

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

	if len(impl.cfg.Api.ProxyProtocol.TrustedProxies) > 0 {
		policy := proxyproto.TrustProxyHeaderFrom(parseIPs(impl.cfg.Api.ProxyProtocol.TrustedProxies)...)

		listener = &proxyproto.Listener{
			Listener:          listener,
			ReadHeaderTimeout: 10 * time.Second,
			ConnPolicy:        policy,
		}

		l.Info("using proxy protocol", log.Any("trusted-proxies", impl.cfg.Api.ProxyProtocol.TrustedProxies))
	}
	defer func() { _ = listener.Close() }()

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

func newTrustCenterHTTPRedirectHandler(proboService *probo.Service, l *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only redirect HTTP requests (no TLS)
		if r.TLS != nil {
			httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))
			return
		}

		domain := r.Host
		if domain == "" {
			httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))
			return
		}

		// Check if this domain is a trust center domain
		_, err := proboService.LoadOrganizationByDomain(ctx, domain)
		if err != nil {
			// Not a trust center domain, return 404
			httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))
			return
		}

		// This is a trust center domain, redirect to HTTPS
		base, err := baseurl.Parse("https://" + domain)
		if err != nil {
			httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))
			return
		}

		httpsURL := base.WithPath(r.URL.Path).WithQueryValues(r.URL.Query()).MustString()
		l.InfoCtx(
			ctx,
			"HTTP request to trust center custom domain, redirecting to HTTPS",
			log.String("domain", domain),
			log.String("path", r.URL.Path),
			log.String("to", httpsURL),
		)
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	})
}

func (impl *Implm) runTrustCenterServer(
	ctx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
	pgClient *pg.Client,
	trustRouter http.Handler,
	acmeService *certmanager.ACMEService,
	proboService *probo.Service,
	encryptionKey cipher.EncryptionKey,
) error {
	tracer := tp.Tracer("go.probo.inc/probo/pkg/probod")
	ctx, span := tracer.Start(ctx, "probod.runTrustCenterServer")
	defer span.End()

	certSelector := certmanager.NewSelector(pgClient, encryptionKey)

	warmer := certmanager.NewCacheStore(pgClient, encryptionKey, l)
	if err := warmer.WarmCache(ctx); err != nil {
		span.RecordError(err)
		l.ErrorCtx(ctx, "cannot warm certificate cache", log.Error(err))
	}

	renewalInterval := time.Duration(impl.cfg.CustomDomains.RenewalInterval) * time.Second
	if renewalInterval == 0 {
		renewalInterval = time.Hour
	}

	renewer := certmanager.NewRenewer(pgClient, acmeService, encryptionKey, renewalInterval, l)

	certProvisioningInterval := time.Duration(impl.cfg.CustomDomains.ProvisionInterval) * time.Second
	if certProvisioningInterval == 0 {
		certProvisioningInterval = 30 * time.Second
	}
	certProvisioner := certmanager.NewProvisioner(pgClient, acmeService, encryptionKey, impl.cfg.CustomDomains.CnameTarget, impl.cfg.CustomDomains.CAAIssuerDomain, certProvisioningInterval, impl.cfg.CustomDomains.ResolverAddr, l)

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
		l.Named("http_acme_handler"),
	)

	httpRedirectHandler := newTrustCenterHTTPRedirectHandler(proboService, l.Named("http_redirect"))

	httpServer := httpserver.NewServer(
		impl.cfg.TrustCenter.HTTPAddr,
		httpACMEHandler.Handle(httpRedirectHandler),
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
			defer func() { _ = listener.Close() }()

			if len(impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies) > 0 {
				policy := proxyproto.TrustProxyHeaderFrom(parseIPs(impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies)...)

				listener = &proxyproto.Listener{
					Listener:          listener,
					ReadHeaderTimeout: 10 * time.Second,
					ConnPolicy:        policy,
				}

				l.Info("using proxy protocol for trust center HTTP server", log.Any("trusted-proxies", impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies))
			}

			if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("cannot serve http requests: %w", err)
			}
			return nil
		},
	)

	acmeHandler := certmanager.NewACMEChallengeHandler(
		pgClient,
		l.Named("acme_handler"),
	)

	handler := acmeHandler.Handle(trustRouter)

	ignoreTLSHandshakeErrors := func(level log.Level, msg string, attrs []log.Attr) bool {
		return strings.Contains(msg, "tls: no certificates configured") ||
			strings.Contains(msg, "client sent an HTTP request to an HTTPS server") ||
			strings.Contains(msg, "tls: client offered only unsupported versions") ||
			strings.Contains(msg, "EOF") ||
			strings.Contains(msg, " i/o timeout") ||
			strings.Contains(msg, "tls: first record does not look like a TLS handshake") ||
			strings.Contains(msg, "tls: client requested unsupported application protocols") ||
			strings.Contains(msg, "read: connection reset by peer") ||
			strings.Contains(msg, "tls: unsupported SSLv2 handshake received") ||
			strings.Contains(msg, "tls: no cipher suite supported by both client and server") ||
			strings.Contains(msg, "tls: received record with version")
	}
	httpServerLogger := l.Named("", log.SkipMatch(ignoreTLSHandshakeErrors))
	httpsServer := httpserver.NewServer(
		impl.cfg.TrustCenter.HTTPSAddr,
		handler,
		httpserver.WithLogger(httpServerLogger),
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
				if errors.Is(err, coredata.ErrResourceNotFound) {
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
			defer func() { _ = listener.Close() }()

			if len(impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies) > 0 {
				policy := proxyproto.TrustProxyHeaderFrom(parseIPs(impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies)...)

				listener = &proxyproto.Listener{
					Listener:          listener,
					ReadHeaderTimeout: 10 * time.Second,
					ConnPolicy:        policy,
				}

				l.Info("using proxy protocol for trust center HTTPS server", log.Any("trusted-proxies", impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies))
			}

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

// parseIPs converts a slice of string IP addresses to net.IP.
// Invalid IPs are skipped.
func parseIPs(strs []string) []net.IP {
	ips := make([]net.IP, 0, len(strs))
	for _, s := range strs {
		if ip := net.ParseIP(s); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips
}
