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

package console_v1

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/server/api/authn"
)

// The install callback is reachable without credentials, and
// httpserver.RenderError serializes err.Error() into the response body, so every
// branch below answers with one of these sentinels and sends the real cause to
// the logger. They are deliberately coarse: which check failed is exactly what a
// prober would like to learn.
var (
	errInstallNotFound             = errors.New("connector install is not available")
	errInstallInvalidCallback      = errors.New("invalid connector install callback")
	errInstallUnauthorizedCallback = errors.New("connector install callback is not authorized")
	errInstallForbiddenCallback    = errors.New("connector install callback is forbidden")
	errInstallInternal             = errors.New("internal error")
)

// Once the caller is known and authorized, the browser is mid-top-level
// navigation and a JSON body is the wrong answer: these go back as an `error`
// query param on the connections page, which already toasts it. They say what
// the customer can do and nothing about which check failed.
const (
	installMessageAlreadyUsed  = "This installation link was already used. Start the connection again from Probo."
	installMessageUnavailable  = "The provider could not be reached to verify the installation. Please try again in a moment."
	installMessageNotVerified  = "The installation could not be verified. Start the connection again from Probo."
	installMessageNotPermitted = "You no longer have permission to connect this provider for this organization."
	installMessageInternal     = "Something went wrong while finishing the installation. Please try again."
)

// installAuthorizationDenied reports whether the authorizer refused the caller,
// as opposed to failing to reach a decision at all. Authorize returns both
// through the same error, and they deserve opposite answers: a refusal is the
// caller's answer, while a Postgres failure reported as 403 would hide an
// outage behind a permissions message and claim a decision was taken that never
// was.
//
// Every refusal it can reach the caller with is enumerated here. They collapse
// to one opaque 403 rather than 401/403/404 apiece: this route is public, so
// which check refused — and whether the organization even exists — is exactly
// what a prober would like to learn.
func installAuthorizationDenied(err error) bool {
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return true
	}

	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		return true
	}

	if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
		return true
	}

	if _, ok := errors.AsType[*iam.ErrOrganizationNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*iam.ErrSessionNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*iam.ErrSessionExpired](err); ok {
		return true
	}

	if _, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
		return true
	}

	return false
}

// installAuthorizationFailure maps an authorizer error onto the answer it
// deserves.
func installAuthorizationFailure(err error) (int, error) {
	if installAuthorizationDenied(err) {
		return http.StatusForbidden, errInstallForbiddenCallback
	}

	return http.StatusInternalServerError, errInstallInternal
}

// installClaimedWorkTimeout bounds EVERYTHING done while the single-use claim
// is held — the vendor call, the second authorization and the write — not just
// the vendor call. A claim untouched for longer than
// coredata.InstallStateStaleAfter is reclaimable by the next request, so any
// step that could outlive that window would be reclaimed from under itself and
// the ceremony would run twice. Bounding only the slowest step is not enough:
// the deadline has to cover their sum.
//
// Pinned by TestInstallClaimedWorkFitsClaimWindow.
const installClaimedWorkTimeout = 30 * time.Second

// handleConnectorInstallComplete finishes an app-install ceremony: the vendor
// top-level-redirects the customer's browser here with Probo's signed state and
// its own proof of the installed tenant.
//
// The ordering below is load-bearing. The state is validated before Postgres is
// touched; the identity is matched and the organization re-authorized before
// anything is claimed, so a rejected callback costs the legitimate owner nothing
// and reveals nothing about the tenant; and the claim is taken before the
// outbound verification, so a replay cannot race a live completion.
func handleConnectorInstallComplete(
	logger *log.Logger,
	iamSvc *iam.Service,
	baseURL *baseurl.BaseURL,
	proboSvc *probo.Service,
	providerRegistry *provider.Registry,
	installStateKey string,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var p coredata.ConnectorProvider
		if err := p.UnmarshalText([]byte(strings.ToUpper(chi.URLParam(r, "provider")))); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, errInstallInvalidCallback)
			return
		}

		reg, ok := providerRegistry.Get(p)
		if !ok || !reg.SupportsInstall() {
			httpserver.RenderError(w, http.StatusNotFound, errInstallNotFound)
			return
		}

		if !providerRegistry.ManagedConnectorReady(p) {
			httpserver.RenderError(w, http.StatusNotFound, errInstallNotFound)
			return
		}

		state := r.URL.Query().Get(reg.Install.StateParam)
		if state == "" {
			httpserver.RenderError(w, http.StatusBadRequest, errInstallInvalidCallback)
			return
		}

		payload, err := connector.ValidateInstallState(installStateKey, state)
		if err != nil {
			// Warn, not Error: an expired state is an ordinary customer outcome.
			// The query carries the state and the vendor's proof, so only the
			// path is logged.
			logger.WarnCtx(
				ctx,
				"rejecting connector install callback with an invalid state",
				log.String("provider", string(p)),
				log.String("path", r.URL.Path),
			)
			httpserver.RenderError(w, http.StatusBadRequest, errInstallInvalidCallback)

			return
		}

		// A state minted for one vendor cannot be spent at another's callback.
		if payload.Data.Provider != string(p) {
			logger.WarnCtx(
				ctx,
				"rejecting connector install callback whose state names another provider",
				log.String("provider", string(p)),
				log.String("path", r.URL.Path),
			)
			httpserver.RenderError(w, http.StatusBadRequest, errInstallInvalidCallback)

			return
		}

		organizationID := payload.Data.OrganizationID
		scope := coredata.NewScopeFromObjectID(organizationID)

		// The state is spendable only by the human it was minted for. Verifying
		// the vendor's proof establishes that the browser came from a real
		// install of that tenant; it says nothing about which Probo organization
		// the tenant belongs in. Without this, anyone holding
		// ActionConnectorInitiate anywhere could hand the vendor's own install
		// link to an administrator of an unrelated tenant and capture it on a
		// wholly genuine vendor token.
		//
		// One opaque answer for "no cookie" and "wrong identity" alike, and
		// never the identity GIDs in the log.
		identity := authn.IdentityFromContext(ctx)
		if identity == nil || identity.ID != payload.Data.IdentityID {
			logger.WarnCtx(
				ctx,
				"rejecting connector install callback carrying no matching identity",
				log.String("provider", string(p)),
				log.String("path", r.URL.Path),
			)
			httpserver.RenderError(w, http.StatusUnauthorized, errInstallUnauthorizedCallback)

			return
		}

		// A valid session is not a live permission: nothing tears a session down
		// when a membership role changes, so the decision the initiate leg took
		// minutes ago is taken again here rather than trusted.
		session := authn.SessionFromContext(ctx)
		if session == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, errInstallUnauthorizedCallback)
			return
		}

		if _, err := iamSvc.Authorizer.Authorize(
			ctx,
			iam.AuthorizeParams{
				Principal: identity.ID,
				Resource:  organizationID,
				Session:   &session.ID,
				Action:    probo.ActionConnectorInitiate,
			},
		); err != nil {
			// The authorizer's own error is logged and never rendered: this
			// route is public, and the reason a decision went against the
			// caller is not a public fact.
			status, rendered := installAuthorizationFailure(err)

			logger.WarnCtx(
				ctx,
				"rejecting unauthorized connector install callback",
				log.String("provider", string(p)),
				log.String("path", r.URL.Path),
				log.Error(err),
			)
			httpserver.RenderError(w, status, rendered)

			return
		}

		// Claimed before the outbound call, so a refresh or a link-preview
		// prefetch cannot run a second verification in parallel with this one.
		processingToken, err := proboSvc.Connectors.ClaimInstallState(ctx, scope, organizationID, state)
		if err != nil {
			if errors.Is(err, probo.ErrInstallStateAlreadyUsed) {
				redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, installMessageAlreadyUsed)
				return
			}

			logger.ErrorCtx(ctx, "cannot claim connector install state", log.Error(err))
			redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, installMessageInternal)

			return
		}

		// From here the claim is held, and every exit has to say what becomes
		// of it: burn what a retry cannot fix, release what it can. Returning
		// with it still held would answer the customer's own retry with
		// "already completed", which is the one thing that is definitely false.
		//
		// One deadline covers everything done under the claim, not just the
		// vendor call: the claim turns reclaimable once it has gone untouched
		// for coredata.InstallStateStaleAfter, so it is the SUM of the steps
		// below that has to stay inside that window. The ledger transitions
		// deliberately keep using ctx, since context.WithoutCancel drops this
		// deadline and they must outlive it.
		//
		claimedCtx, cancelClaimed := context.WithTimeout(ctx, installClaimedWorkTimeout)
		defer cancelClaimed()

		// Verify with the same credential the persisted connector will use, so
		// what the ceremony proves is what the driver can later do.
		managedKey, _ := providerRegistry.ManagedAPIKey(p)

		appID, _ := providerRegistry.ManagedResourceID(p)

		httpClient, err := providerRegistry.NewAPIKeyConnection(p, managedKey).Client(ctx)
		if err != nil {
			releaseInstallState(ctx, logger, proboSvc, scope, organizationID, state, processingToken)
			logger.ErrorCtx(ctx, "cannot build connector install verification client", log.Error(err))
			redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, installMessageInternal)

			return
		}

		resourceID, err := reg.Install.Verify(claimedCtx, httpClient, appID, r.URL.Query())
		if err != nil {
			// Transient failures release the claim so the customer's remaining
			// window still works; everything else burns it, so a forged proof
			// gets exactly one attempt. The classification is the provider's,
			// taken on the vendor's status and never on an error string.
			if errors.Is(err, provider.ErrInstallVerificationTransient) {
				releaseInstallState(ctx, logger, proboSvc, scope, organizationID, state, processingToken)
				logger.WarnCtx(
					ctx,
					"connector install verification is temporarily unavailable",
					log.String("provider", string(p)),
					log.Error(err),
				)
				redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, installMessageUnavailable)

				return
			}

			burnInstallState(ctx, logger, proboSvc, scope, organizationID, state, processingToken)
			logger.WarnCtx(
				ctx,
				"rejecting connector install callback whose proof did not verify",
				log.String("provider", string(p)),
				log.Error(err),
			)
			redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, installMessageNotVerified)

			return
		}

		// Authorized a second time, after the vendor answered. The first check
		// guards the claim; this one guards the write. Nothing tears a session
		// down when a membership is revoked, so a decision taken before a
		// network round-trip is not the decision that should authorize an
		// insert. The residual window shrinks to the microseconds between here
		// and the transaction, from however long the vendor took to reply.
		if _, err := iamSvc.Authorizer.Authorize(
			claimedCtx,
			iam.AuthorizeParams{
				Principal: identity.ID,
				Resource:  organizationID,
				Session:   &session.ID,
				Action:    probo.ActionConnectorInitiate,
			},
		); err != nil {
			// A denial is final for this state; a storage failure is not, and
			// burning on one would cost the customer a ceremony they completed
			// correctly.
			message := installMessageNotPermitted

			if installAuthorizationDenied(err) {
				burnInstallState(ctx, logger, proboSvc, scope, organizationID, state, processingToken)
			} else {
				releaseInstallState(ctx, logger, proboSvc, scope, organizationID, state, processingToken)

				message = installMessageInternal
			}

			logger.WarnCtx(
				ctx,
				"rejecting connector install callback whose authorization lapsed mid-flight",
				log.String("provider", string(p)),
				log.String("path", r.URL.Path),
				log.Error(err),
			)
			redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, message)

			return
		}

		// The key is empty by design: (*provider.Registry).APIKeyFor substitutes
		// the Probo-held key at every use, so anything stored here is discarded.
		cnnctr, err := proboSvc.Connectors.CompleteInstall(
			claimedCtx,
			scope,
			probo.CompleteConnectorInstallRequest{
				OrganizationID:  organizationID,
				Provider:        p,
				SettingsKey:     reg.Install.SettingsResourceKey,
				ResourceID:      resourceID,
				Connection:      providerRegistry.NewAPIKeyConnection(p, ""),
				State:           state,
				ProcessingToken: processingToken,
			},
		)
		if err != nil {
			// The burn lives inside that transaction, so it rolled back with
			// it: the claim is still held and a retry is the right answer.
			releaseInstallState(ctx, logger, proboSvc, scope, organizationID, state, processingToken)
			logger.ErrorCtx(ctx, "cannot complete connector install", log.Error(err))
			redirectInstallOutcome(w, r, logger, baseURL, safeRedirect, organizationID, installMessageInternal)

			return
		}

		// Rebuilt server-side from the state's organization rather than copied
		// from a callback parameter: this flow has one entry point, so there is
		// no continue URL to honour and none to validate.
		redirectURL, err := baseURL.
			WithPath("/organizations/"+organizationID.String()+"/access-reviews/connections").
			WithQuery("connector_id", cnnctr.ID.String()).
			WithQuery("provider", string(p)).
			String()
		if err != nil {
			logger.ErrorCtx(ctx, "cannot build connector install redirect URL", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errInstallInternal)

			return
		}

		safeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
	}
}

// redirectInstallOutcome sends an authorized customer back to their connections
// page carrying a human-readable failure, rather than answering a top-level
// browser navigation with a JSON error body. It is only reachable once the
// identity and the organization are established; the pre-authorization branches
// render instead, because a redirect there would tell an anonymous prober which
// organization the state named.
func redirectInstallOutcome(
	w http.ResponseWriter,
	r *http.Request,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	safeRedirect *saferedirect.SafeRedirect,
	organizationID gid.GID,
	message string,
) {
	redirectURL, err := baseURL.
		WithPath("/organizations/"+organizationID.String()+"/access-reviews/connections").
		WithQuery("error", message).
		String()
	if err != nil {
		logger.ErrorCtx(r.Context(), "cannot build connector install error redirect URL", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInstallInternal)

		return
	}

	safeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
}

// releaseInstallState hands the claim back for a failure a retry could fix.
//
// context.WithoutCancel: the ledger transition must outlive a browser that hung
// up mid-flight, and outlive the verification timeout that may have caused it.
func releaseInstallState(
	ctx context.Context,
	logger *log.Logger,
	proboSvc *probo.Service,
	scope coredata.Scoper,
	organizationID gid.GID,
	state string,
	processingToken string,
) {
	if err := proboSvc.Connectors.ReleaseInstallState(
		context.WithoutCancel(ctx),
		scope,
		organizationID,
		state,
		processingToken,
	); err != nil {
		logger.ErrorCtx(ctx, "cannot release connector install state", log.Error(err))
	}
}

// burnInstallState spends the claim for good, so a refused proof gets exactly
// one attempt.
func burnInstallState(
	ctx context.Context,
	logger *log.Logger,
	proboSvc *probo.Service,
	scope coredata.Scoper,
	organizationID gid.GID,
	state string,
	processingToken string,
) {
	if err := proboSvc.Connectors.BurnInstallState(
		context.WithoutCancel(ctx),
		scope,
		organizationID,
		state,
		processingToken,
	); err != nil {
		logger.ErrorCtx(ctx, "cannot burn connector install state", log.Error(err))
	}
}
