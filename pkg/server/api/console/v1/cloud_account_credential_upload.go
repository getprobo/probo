// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package console_v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authz"
)

const (
	// cloudAccountCredentialUploadMaxBodyBytes caps the multipart
	// body. Even Azure JWT bundles fit comfortably under 16 KiB; any
	// caller exceeding this is either malicious or misconfigured.
	cloudAccountCredentialUploadMaxBodyBytes = 16 * 1024

	// cloudAccountCredentialUploadMaxMemoryBytes caps in-memory
	// allocation during multipart parsing. Set lower than the body
	// cap so disk spilling never kicks in for legitimate payloads.
	cloudAccountCredentialUploadMaxMemoryBytes = 16 * 1024
)

// handleCloudAccountCredentialUpload implements the
// POST /cloud-accounts/credentials/upload route.
//
// The credential body never lands in any access log: kit/httpserver
// only logs request metadata (method, path, status, size) and
// never the body. The upload form-field name "payload" is
// documented here as the carrier for the secret bytes.
//
// The handler distinguishes "first attach" (cloud-account row in
// PENDING_VERIFICATION with an empty credential envelope) from
// "rotation" (any other state) and authorises against
// ActionCloudAccountCreate or ActionCloudAccountRotateCredentials
// accordingly.
func handleCloudAccountCredentialUpload(
	logger *log.Logger,
	proboSvc *probo.Service,
	authorize authz.AuthorizeFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Cap body size before ParseMultipartForm reads anything.
		r.Body = http.MaxBytesReader(w, r.Body, cloudAccountCredentialUploadMaxBodyBytes)

		if err := r.ParseMultipartForm(cloudAccountCredentialUploadMaxMemoryBytes); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot parse multipart form: %w", err))
			return
		}

		cloudAccountIDStr := r.FormValue("cloud_account_id")
		if cloudAccountIDStr == "" {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing cloud_account_id"))
			return
		}

		cloudAccountID, err := gid.ParseGID(cloudAccountIDStr)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot parse cloud_account_id: %w", err))
			return
		}

		_, payloadHeader, err := r.FormFile("payload")
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing payload form file: %w", err))
			return
		}
		if payloadHeader.Size > cloudAccountCredentialUploadMaxBodyBytes {
			httpserver.RenderError(w, http.StatusRequestEntityTooLarge, errors.New("payload too large"))
			return
		}

		payload, err := readMultipartPayload(r)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot read payload: %w", err))
			return
		}

		ctx := r.Context()
		prb := proboSvc.WithTenant(cloudAccountID.TenantID())

		// Load the row to (a) decide attach-vs-rotate authz action
		// and (b) read the row's organization_id for the authorize
		// resource. Use GetMetadata so the credential envelope
		// stays encrypted.
		account, err := prb.CloudAccounts.GetMetadata(ctx, cloudAccountID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				httpserver.RenderError(w, http.StatusNotFound, errors.New("cloud account not found"))
				return
			}
			logger.ErrorCtx(ctx, "cannot load cloud account", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal error"))
			return
		}

		action := probo.ActionCloudAccountRotateCredentials
		if isFirstAttach(account) {
			action = probo.ActionCloudAccountCreate
		}

		if err := authorize(ctx, account.OrganizationID, action); err != nil {
			// authorize already returns gqlerror values for the
			// graphql path; for the REST path we surface a 403
			// with no body details so the response shape stays
			// uniform.
			httpserver.RenderError(w, http.StatusForbidden, errors.New("forbidden"))
			return
		}

		req := probo.RotateCloudAccountCredentialsRequest{
			CloudAccountID: cloudAccountID,
			Provider:       account.Provider,
			CredentialKind: account.CredentialKind,
		}
		switch account.Provider {
		case coredata.CloudAccountProviderGCP:
			req.GCPServiceAccountJSON = payload
		case coredata.CloudAccountProviderAzure:
			req.AzureClientSecret = string(payload)
		default:
			// AWS uses no credential body -- the install assets
			// flow + GraphQL createCloudAccount already carry the
			// role ARN + external id metadata. Reject explicitly
			// so a misrouted upload doesn't silently no-op.
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("provider %q does not support credential upload", account.Provider))
			return
		}

		updated, err := prb.CloudAccounts.RotateCredentials(ctx, req)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				httpserver.RenderError(w, http.StatusNotFound, errors.New("cloud account not found"))
				return
			}
			if errors.Is(err, cloudaccount.ErrCredentialsInvalid) {
				httpserver.RenderError(w, http.StatusBadRequest, errors.New("credentials invalid"))
				return
			}
			logger.ErrorCtx(ctx, "cannot attach cloud account credentials", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal error"))
			return
		}

		verifyResult, verifyErr := prb.CloudAccounts.Verify(ctx, updated.ID)
		if verifyErr != nil {
			// Verify failure is not fatal for the upload itself
			// (the row is now in PENDING_VERIFICATION with the
			// new envelope persisted). Surface the error string
			// in lastProbeError so the operator can see it.
			logger.ErrorCtx(ctx, "cannot verify cloud account after upload", log.Error(verifyErr))
			fallback := verifyErr.Error()
			httpserver.RenderJSON(w, http.StatusOK, cloudAccountCredentialUploadResponse{
				Status:         coredata.CloudAccountStatusPendingVerification,
				LastProbeError: &fallback,
			})
			return
		}

		httpserver.RenderJSON(w, http.StatusOK, cloudAccountCredentialUploadResponse{
			Status:         verifyResult.Status,
			LastProbeError: verifyResult.LastProbeError,
		})
	}
}

// cloudAccountCredentialUploadResponse is the JSON body returned to
// the wizard.
type cloudAccountCredentialUploadResponse struct {
	Status         coredata.CloudAccountStatus `json:"status"`
	LastProbeError *string                     `json:"lastProbeError,omitempty"`
}

// readMultipartPayload reads the "payload" form file fully into
// memory and returns the bytes. Body size has already been capped
// by MaxBytesReader on the request.
func readMultipartPayload(r *http.Request) ([]byte, error) {
	f, _, err := r.FormFile("payload")
	if err != nil {
		return nil, fmt.Errorf("cannot open payload form file: %w", err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read payload form file: %w", err)
	}
	return buf, nil
}

// isFirstAttach returns true when the row has never had a
// credential envelope attached. Used to pick between
// ActionCloudAccountCreate and ActionCloudAccountRotateCredentials.
//
// The row is in PENDING_VERIFICATION immediately after
// createCloudAccount; once the first upload + Verify completes, it
// transitions to VERIFIED or ERRORED. A row stuck in
// PENDING_VERIFICATION with an empty encrypted_credentials column
// is a fresh row awaiting its first credentials upload.
func isFirstAttach(account *coredata.CloudAccount) bool {
	if account.Status != coredata.CloudAccountStatusPendingVerification {
		return false
	}
	return len(account.EncryptedCredentials) == 0
}
