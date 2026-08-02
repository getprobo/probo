// Command test-client runs a minimal OIDC authorization-code login against the
// Keycloak "govrly-test" realm and prints the resulting ID token claims.
//
// Usage:
//
//	OIDC_ISSUER=https://<service>.onrender.com/realms/govrly-test \
//	OIDC_CLIENT_ID=govrly-api \
//	OIDC_CLIENT_SECRET=<from admin console> \
//	go run .
//
// Then open http://localhost:3000/ in a browser.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const listenAddr = "localhost:3000"

func main() {
	issuer := mustEnv("OIDC_ISSUER")
	clientID := mustEnv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET") // empty for a public client
	redirectURL := envOr("OIDC_REDIRECT_URL", "http://"+listenAddr+"/callback")
	scopes := strings.Fields(envOr("OIDC_SCOPES", "openid profile email"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatalf("discovery failed for %s: %v", issuer, err)
	}
	log.Printf("discovered issuer %s", provider.Endpoint().AuthURL)

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	// One-shot login: the callback handler reports the outcome and we exit.
	done := make(chan error, 1)

	// PKCE + CSRF/replay protection. Regenerated on every visit to "/".
	var (
		state        = randomString()
		nonce        = randomString()
		codeVerifier = oauth2.GenerateVerifier()
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		state, nonce, codeVerifier = randomString(), randomString(), oauth2.GenerateVerifier()
		authURL := oauthCfg.AuthCodeURL(state,
			oidc.Nonce(nonce),
			oauth2.S256ChallengeOption(codeVerifier),
		)
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			failf(w, done, "authorization error: %s: %s", errParam, r.URL.Query().Get("error_description"))
			return
		}
		if r.URL.Query().Get("state") != state {
			failf(w, done, "state mismatch — possible CSRF")
			return
		}

		token, err := oauthCfg.Exchange(r.Context(), r.URL.Query().Get("code"),
			oauth2.VerifierOption(codeVerifier),
		)
		if err != nil {
			failf(w, done, "code exchange failed: %v", err)
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			failf(w, done, "no id_token in token response")
			return
		}
		idToken, err := verifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			failf(w, done, "id_token verification failed: %v", err)
			return
		}
		if idToken.Nonce != nonce {
			failf(w, done, "nonce mismatch — possible token replay")
			return
		}

		var claims map[string]any
		if err := idToken.Claims(&claims); err != nil {
			failf(w, done, "cannot decode claims: %v", err)
			return
		}
		pretty, _ := json.MarshalIndent(claims, "", "  ")

		fmt.Println("\n=== ID token claims ===")
		fmt.Println(string(pretty))
		fmt.Println("=== token metadata ===")
		fmt.Printf("token_type:     %s\n", token.TokenType)
		fmt.Printf("expiry:         %s\n", token.Expiry.Format(time.RFC3339))
		fmt.Printf("refresh_token:  %t\n", token.RefreshToken != "")
		fmt.Printf("acr:            %v\n", claims["acr"])
		fmt.Printf("amr:            %v\n", claims["amr"])

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>Login OK</h1><p>Subject: %s</p><pre>%s</pre>",
			htmlEscape(idToken.Subject), htmlEscape(string(pretty)))

		done <- nil
	})

	srv := &http.Server{Addr: listenAddr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		log.Printf("open http://%s/ to start the login", listenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			done <- fmt.Errorf("server: %w", err)
		}
	}()

	var runErr error
	select {
	case runErr = <-done:
	case <-ctx.Done():
		log.Println("interrupted")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	if runErr != nil {
		log.Fatal(runErr)
	}
	log.Println("done")
}

func failf(w http.ResponseWriter, done chan<- error, format string, args ...any) {
	err := fmt.Errorf(format, args...)
	http.Error(w, err.Error(), http.StatusBadRequest)
	done <- err
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func randomString() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("cannot read randomness: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
