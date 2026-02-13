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
	"go.probo.inc/probo/pkg/agents"
	"go.probo.inc/probo/pkg/awsconfig"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/keys"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailer"
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
		cfg config
	}

	config struct {
		BaseURL       *baseurl.BaseURL     `json:"base-url"`
		EncryptionKey cipher.EncryptionKey `json:"encryption-key"`
		Pg            pgConfig             `json:"pg"`
		Api           apiConfig            `json:"api"`
		Auth          authConfig           `json:"auth"`
		TrustCenter   trustCenterConfig    `json:"trust-center"`
		AWS           awsConfig            `json:"aws"`
		Notifications notificationsConfig  `json:"notifications"`
		Connectors    []connectorConfig    `json:"connectors"`
		OpenAI        openaiConfig         `json:"openai"`
		ChromeDPAddr  string               `json:"chrome-dp-addr"`
		CustomDomains customDomainsConfig  `json:"custom-domains"`
		SCIMBridge    scimBridgeConfig     `json:"scim-bridge"`
	}

	trustCenterConfig struct {
		HTTPAddr      string              `json:"http-addr"`
		HTTPSAddr     string              `json:"https-addr"`
		ProxyProtocol proxyProtocolConfig `json:"proxy-protocol"`
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
				PasswordResetTokenValidity:          3600,
				MagicLinkTokenValidity:              900,
				SAML: samlConfig{
					SessionDuration:                   604800,
					CleanupIntervalSeconds:            86400,
					DomainVerificationIntervalSeconds: 60,
					DomainVerificationResolverAddr:    "8.8.8.8:53",
				},
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
				Webhook: webhookConfig{
					SenderInterval: 5,
					CacheTTL:       86400,
				},
			},
			CustomDomains: customDomainsConfig{
				RenewalInterval:   3600,
				ProvisionInterval: 30,
				ResolverAddr:      "8.8.8.8:53",
				ACME: acmeConfig{
					Directory: "https://acme-v02.api.letsencrypt.org/directory",
					Email:     "admin@getprobo.com",
					KeyType:   "EC256",
				},
			},
			SCIMBridge: scimBridgeConfig{
				SyncInterval: 60, // 15 minutes
				PollInterval: 30, // 30 seconds
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

	agentConfig := agents.Config{
		OpenAIAPIKey: impl.cfg.OpenAI.APIKey,
		Temperature:  impl.cfg.OpenAI.Temperature,
		ModelName:    impl.cfg.OpenAI.ModelName,
	}

	agent := agents.NewAgent(l.Named("agent"), agentConfig)

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
			BaseURL:                        impl.cfg.BaseURL,
			EncryptionKey:                  impl.cfg.EncryptionKey,
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
		impl.cfg.BaseURL.String(),
		impl.cfg.EncryptionKey,
		impl.cfg.Auth.Cookie.Secret,
		l.Named("slack"),
	)

	proboService, err := probo.NewService(
		ctx,
		impl.cfg.EncryptionKey,
		pgClient,
		s3Client,
		impl.cfg.AWS.Bucket,
		impl.cfg.BaseURL.String(),
		impl.cfg.Auth.Cookie.Secret,
		agentConfig,
		html2pdfConverter,
		acmeService,
		fileManagerService,
		l.Named("probo"),
		slackService,
		iamService,
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
		impl.cfg.GetSlackSigningSecret(),
		iamService,
		html2pdfConverter,
		fileManagerService,
		l,
		slackService,
	)

	serverHandler, err := server.NewServer(
		server.Config{
			AllowedOrigins:    impl.cfg.Api.Cors.AllowedOrigins,
			ExtraHeaderFields: impl.cfg.Api.ExtraHeaderFields,
			Probo:             proboService,
			IAM:               iamService,
			Trust:             trustService,
			Slack:             slackService,
			ConnectorRegistry: defaultConnectorRegistry,
			BaseURL:           impl.cfg.BaseURL,
			Agent:             agent,
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
	mailer := mailer.NewMailer(
		pgClient,
		l,
		mailer.Config{
			SenderEmail: impl.cfg.Notifications.Mailer.SenderEmail,
			SenderName:  impl.cfg.Notifications.Mailer.SenderName,
			Addr:        impl.cfg.Notifications.Mailer.SMTP.Addr,
			User:        impl.cfg.Notifications.Mailer.SMTP.User,
			Password:    impl.cfg.Notifications.Mailer.SMTP.Password,
			TLSRequired: impl.cfg.Notifications.Mailer.SMTP.TLSRequired,
			Timeout:     time.Second * 10,
			Interval:    time.Duration(impl.cfg.Notifications.Mailer.MailerInterval) * time.Second,
		},
	)
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

	webhookSenderCtx, stopWebhookSender := context.WithCancel(context.Background())
	webhookSender := webhook.NewSender(pgClient, l.Named("webhook-sender"), webhook.Config{
		Interval:      time.Duration(impl.cfg.Notifications.Webhook.SenderInterval) * time.Second,
		CacheTTL:      time.Duration(impl.cfg.Notifications.Webhook.CacheTTL) * time.Second,
		EncryptionKey: impl.cfg.EncryptionKey,
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

	trustCenterServerCtx, stopTrustCenterServer := context.WithCancel(context.Background())
	defer stopTrustCenterServer()
	wg.Go(
		func() {
			if err := impl.runTrustCenterServer(trustCenterServerCtx, l, r, tp, pgClient, serverHandler.TrustCenterHandler(), acmeService, proboService); err != nil {
				cancel(fmt.Errorf("trust center server crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopMailer()
	stopSlackSender()
	stopWebhookSender()
	stopExportJobExporter()
	stopIAMService()
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

	if len(impl.cfg.Api.ProxyProtocol.TrustedProxies) > 0 {
		policy := proxyproto.TrustProxyHeaderFrom(impl.cfg.Api.ProxyProtocol.TrustedProxies...)

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
		httpsURL := "https://" + r.Host + r.URL.RequestURI()
		l.InfoCtx(ctx, "HTTP request to trust center custom domain, redirecting to HTTPS",
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
	certProvisioner := certmanager.NewProvisioner(pgClient, acmeService, impl.cfg.EncryptionKey, impl.cfg.CustomDomains.CnameTarget, certProvisioningInterval, impl.cfg.CustomDomains.ResolverAddr, l)

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
				policy := proxyproto.TrustProxyHeaderFrom(impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies...)

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
		impl.cfg.EncryptionKey,
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
				policy := proxyproto.TrustProxyHeaderFrom(impl.cfg.TrustCenter.ProxyProtocol.TrustedProxies...)

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
