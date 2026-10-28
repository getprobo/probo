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

package tasksync

import (
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

var (
	ErrLinearNotConnected          = errors.New("linear connector is not connected")
	ErrLinearReconnectRequired     = errors.New("linear connector must be reconnected with write scopes")
	ErrTaskAlreadyLinked           = errors.New("task is already linked to an external issue")
	ErrTaskNotLinked               = errors.New("task is not linked to an external issue")
	ErrLinearTeamNotFound          = errors.New("linear team was not found")
	ErrLinearTeamIDRequired        = errors.New("linear team id is required")
	ErrLinearIssueIDRequired       = errors.New("linear issue id is required")
	ErrLinearIssueNotFound         = errors.New("linear issue was not found")
	ErrLinearDefaultTeamRequired   = errors.New("a default Linear team is required")
	ErrLinearSettingsRequired      = errors.New("a Linear team or an issue import status is required")
	ErrAsanaNotConnected           = errors.New("asana connector is not connected")
	ErrAsanaProjectIDRequired      = errors.New("asana project id is required")
	ErrAsanaProjectNotFound        = errors.New("asana project was not found")
	ErrAsanaTaskIDRequired         = errors.New("asana task id is required")
	ErrAsanaTaskNotFound           = errors.New("asana task was not found")
	ErrAsanaDefaultProjectRequired = errors.New("a default Asana project is required")
	ErrAsanaSettingsRequired       = errors.New("an Asana project or a task import status is required")
	ErrAsanaMappingInvalid         = errors.New("asana project mapping is invalid")
)

type LinearReconnectRequiredError struct {
	Scopes []string
}

func (e *LinearReconnectRequiredError) Error() string {
	if len(e.Scopes) == 0 {
		return ErrLinearReconnectRequired.Error()
	}

	return formatMissingScopes(e.Scopes)
}

func (e *LinearReconnectRequiredError) Is(target error) bool {
	return target == ErrLinearReconnectRequired
}

func NewLinearReconnectRequiredError(scopes []string) error {
	return &LinearReconnectRequiredError{Scopes: append([]string(nil), scopes...)}
}

func requiredTaskSyncScopes() []string {
	return []string{"read", "write", "issues:create"}
}

func missingTaskSyncScopes(granted []string) []string {
	have := make(map[string]struct{}, len(granted))
	for _, scope := range granted {
		have[scope] = struct{}{}
	}

	var missing []string

	for _, scope := range requiredTaskSyncScopes() {
		if _, ok := have[scope]; !ok {
			missing = append(missing, scope)
		}
	}

	return missing
}

func formatMissingScopes(scopes []string) string {
	return fmt.Sprintf("missing Linear OAuth scopes: %v", scopes)
}

func LinearPublishClientMessage(err error) (string, bool) {
	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return "task not found", true
	case errors.Is(err, ErrLinearNotConnected):
		return ErrLinearNotConnected.Error(), true
	case errors.Is(err, ErrLinearReconnectRequired):
		if reconnect, ok := errors.AsType[*LinearReconnectRequiredError](err); ok {
			return reconnect.Error(), true
		}

		return ErrLinearReconnectRequired.Error(), true
	case errors.Is(err, ErrTaskAlreadyLinked),
		errors.Is(err, coredata.ErrResourceAlreadyExists):
		return ErrTaskAlreadyLinked.Error(), true
	case errors.Is(err, ErrLinearTeamIDRequired):
		return ErrLinearTeamIDRequired.Error(), true
	case errors.Is(err, ErrLinearTeamNotFound):
		return ErrLinearTeamNotFound.Error(), true
	default:
		return "could not publish the task to Linear", false
	}
}
