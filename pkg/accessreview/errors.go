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

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/aws/smithy-go"
	"golang.org/x/oauth2"

	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var (
	ErrCampaignMissingSources    = errors.New("access review campaign missing scope sources")
	ErrCampaignNotDraft          = errors.New("access review campaign not draft")
	ErrCampaignNotDeletable      = errors.New("access review campaign not deletable")
	ErrCampaignNotPendingActions = errors.New("access review campaign not pending actions")
	ErrCampaignCompleted         = errors.New("access review campaign completed")
	ErrCampaignCancelled         = errors.New("access review campaign cancelled")
	ErrMissingOAuthScopes        = errors.New("missing required OAuth scopes")
)

type (
	CampaignMissingSourcesError struct {
		CampaignID gid.GID
	}

	CampaignNotDraftError struct {
		CampaignID gid.GID
	}

	CampaignNotDeletableError struct {
		CampaignID gid.GID
	}

	CampaignNotPendingActionsError struct {
		CampaignID gid.GID
	}

	CampaignCompletedError struct {
		CampaignID gid.GID
	}

	CampaignCancelledError struct {
		CampaignID gid.GID
	}

	MissingOAuthScopesError struct {
		Scopes []string
	}

	// ProbeError carries the provider a credential probe failed for, so
	// callers log it as a field rather than parse it out of a message.
	ProbeError struct {
		Provider coredata.ConnectorProvider
		Err      error
	}
)

func NewCampaignMissingSourcesError(campaignID gid.GID) error {
	return &CampaignMissingSourcesError{CampaignID: campaignID}
}

func (e *CampaignMissingSourcesError) Error() string {
	return fmt.Sprintf(
		"access review campaign %q cannot be started: no scope sources configured",
		e.CampaignID,
	)
}

func (e *CampaignMissingSourcesError) Is(target error) bool {
	return target == ErrCampaignMissingSources
}

func NewCampaignNotDraftError(campaignID gid.GID) error {
	return &CampaignNotDraftError{CampaignID: campaignID}
}

func (e *CampaignNotDraftError) Error() string {
	return fmt.Sprintf("access review campaign %q is not in draft", e.CampaignID)
}

func (e *CampaignNotDraftError) Is(target error) bool {
	return target == ErrCampaignNotDraft
}

func NewCampaignNotDeletableError(campaignID gid.GID) error {
	return &CampaignNotDeletableError{CampaignID: campaignID}
}

func (e *CampaignNotDeletableError) Error() string {
	return fmt.Sprintf(
		"access review campaign %q cannot be deleted while it is in progress",
		e.CampaignID,
	)
}

func (e *CampaignNotDeletableError) Is(target error) bool {
	return target == ErrCampaignNotDeletable
}

func NewCampaignNotPendingActionsError(campaignID gid.GID) error {
	return &CampaignNotPendingActionsError{CampaignID: campaignID}
}

func (e *CampaignNotPendingActionsError) Error() string {
	return fmt.Sprintf(
		"access review campaign %q cannot be closed unless it is pending actions",
		e.CampaignID,
	)
}

func (e *CampaignNotPendingActionsError) Is(target error) bool {
	return target == ErrCampaignNotPendingActions
}

func NewCampaignCompletedError(campaignID gid.GID) error {
	return &CampaignCompletedError{CampaignID: campaignID}
}

func (e *CampaignCompletedError) Error() string {
	return fmt.Sprintf("access review campaign %q is already completed", e.CampaignID)
}

func (e *CampaignCompletedError) Is(target error) bool {
	return target == ErrCampaignCompleted
}

func NewCampaignCancelledError(campaignID gid.GID) error {
	return &CampaignCancelledError{CampaignID: campaignID}
}

func (e *CampaignCancelledError) Error() string {
	return fmt.Sprintf("access review campaign %q is already cancelled", e.CampaignID)
}

func (e *CampaignCancelledError) Is(target error) bool {
	return target == ErrCampaignCancelled
}

func NewMissingOAuthScopesError(scopes []string) error {
	return &MissingOAuthScopesError{Scopes: append([]string(nil), scopes...)}
}

func (e *MissingOAuthScopesError) Error() string {
	display := make([]string, len(e.Scopes))
	for i, scope := range e.Scopes {
		display[i] = strings.TrimPrefix(scope, "https://graph.microsoft.com/")
	}

	return "Missing required OAuth scopes: " + strings.Join(display, ", ")
}

func (e *MissingOAuthScopesError) Is(target error) bool {
	return target == ErrMissingOAuthScopes
}

func NewProbeError(prvdr coredata.ConnectorProvider, err error) error {
	return &ProbeError{Provider: prvdr, Err: err}
}

func (e *ProbeError) Error() string {
	return fmt.Sprintf("cannot probe %s connector: %v", e.Provider, e.Err)
}

func (e *ProbeError) Unwrap() error {
	return e.Err
}

// isNilError reports whether err is nil, or whether its chain holds a typed
// nil. A typed nil panics when something calls its Unwrap, which errors.AsType
// does while walking, so the classifiers below screen the chain first: they run
// while logging a failure, where a panic costs more than a coarse answer. Each
// node is checked before it is unwrapped, so the screen cannot trip over the
// value it is looking for.
func isNilError(err error) bool {
	if err == nil {
		return true
	}

	switch value := reflect.ValueOf(err); value.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if value.IsNil() {
			return true
		}
	}

	switch unwrapper := err.(type) {
	case interface{ Unwrap() error }:
		cause := unwrapper.Unwrap()
		if cause == nil {
			// A real error that simply wraps nothing, not a nil one.
			return false
		}

		return isNilError(cause)

	case interface{ Unwrap() []error }:
		if slices.ContainsFunc(unwrapper.Unwrap(), isNilError) {
			return true
		}
	}

	return false
}

// IsProviderVerdict reports whether err is the provider's answer rather than a
// failure on Probo's side. Only a rejected credential, a transport failure that
// reached the provider, and a refused token refresh qualify. Everything else a
// probe can return (settings that will not decode, a request that could not be
// built, a registry misconfiguration) is ours, so the default is to treat a
// failure as Probo's and report it in full.
func IsProviderVerdict(err error) bool {
	if isNilError(err) {
		return false
	}

	// Checked before any verdict: our own deadline expiring is never the
	// provider's answer, even when it is joined with one.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	if _, ok := errors.AsType[*provider.CredentialRejectedError](err); ok {
		return true
	}

	// ok can be true with a nil value when a custom As assigns one, so the
	// dereference below needs its own guard.
	if urlErr, ok := errors.AsType[*url.Error](err); ok {
		if urlErr == nil {
			return false
		}

		// url.Parse failures arrive as a *url.Error too, with nothing sent.
		return urlErr.Op != "parse"
	}

	if _, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
		return true
	}

	// A workload identity connector never makes an HTTP request of its own:
	// STS answers through the AWS SDK, so its API errors are the verdict.
	if _, ok := errors.AsType[smithy.APIError](err); ok {
		return true
	}

	return false
}

// ProbeFailureCode reduces a probe failure to a token safe to log. Probe
// errors wrap text Probo does not control (an OAuth error_description, a
// response body, a customer's self-hosted host), so anything unrecognised
// degrades to its Go type rather than being quoted.
func ProbeFailureCode(err error) string {
	if isNilError(err) {
		return "none"
	}

	// Guarded against a typed nil: this runs on a logging path, where a panic
	// would cost more than the lost detail.
	if rejected, ok := errors.AsType[*provider.CredentialRejectedError](err); ok && rejected != nil {
		return fmt.Sprintf("credential_rejected_%d", rejected.StatusCode)
	}

	if _, ok := errors.AsType[*url.Error](err); ok {
		return "transport_error"
	}

	// An AWS error code is a fixed identifier such as AccessDenied, so it is
	// safe to report where the surrounding message is not.
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr != nil {
		return fmt.Sprintf("aws_%s", apiErr.ErrorCode())
	}

	cause := err
	for {
		unwrapped := errors.Unwrap(cause)
		if unwrapped == nil {
			break
		}

		cause = unwrapped
	}

	return fmt.Sprintf("%T", cause)
}
