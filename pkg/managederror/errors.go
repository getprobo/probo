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

package managederror

import (
	"errors"
	"fmt"
)

const (
	CodeUnauthenticated = "UNAUTHENTICATED"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeTenantNotFound  = "TENANT_NOT_FOUND"
	CodeSessionExpired  = "SESSION_EXPIRED"

	CodeNotFound      = "NOT_FOUND"
	CodeAlreadyExists = "ALREADY_EXISTS"
	CodeConflict      = "CONFLICT"

	CodeInvalidInput    = "INVALID_INPUT"
	CodeValidationError = "VALIDATION_ERROR"

	CodeBusinessRuleViolation = "BUSINESS_RULE_VIOLATION"
	CodeOperationNotAllowed   = "OPERATION_NOT_ALLOWED"

	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeTimeout             = "TIMEOUT"

	CodeFileNotSupported = "FILE_NOT_SUPPORTED"
	CodeFileTooLarge     = "FILE_TOO_LARGE"
	CodeUploadFailed     = "UPLOAD_FAILED"
)

type ErrorManaged struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e *ErrorManaged) Error() string {
	return e.Message
}

func NewErrorManaged(errorCode, message string) *ErrorManaged {
	return &ErrorManaged{
		Code:    errorCode,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

func NewErrorManagedf(errorCode, format string, args ...any) *ErrorManaged {
	return &ErrorManaged{
		Code:    errorCode,
		Message: fmt.Sprintf(format, args...),
		Details: make(map[string]interface{}),
	}
}

func (e *ErrorManaged) WithDetail(key string, value any) *ErrorManaged {
	e.Details[key] = value
	return e
}

func (e *ErrorManaged) WithField(field string, message string) *ErrorManaged {
	if e.Details["fields"] == nil {
		e.Details["fields"] = make(map[string]string)
	}
	e.Details["fields"].(map[string]string)[field] = message
	return e
}

func NewUnauthenticatedError() *ErrorManaged {
	return NewErrorManaged(CodeUnauthenticated, "Authentication required")
}

func NewUnauthorizedError() *ErrorManaged {
	return NewErrorManaged(CodeUnauthorized, "Access denied")
}

func NewForbiddenError(message string) *ErrorManaged {
	return NewErrorManaged(CodeForbidden, message)
}

func NewTenantNotFoundError() *ErrorManaged {
	return NewErrorManaged(CodeTenantNotFound, "Tenant not found")
}

func NewSessionExpiredError() *ErrorManaged {
	return NewErrorManaged(CodeSessionExpired, "Session expired")
}

func NewNotFoundError(resource string) *ErrorManaged {
	return NewErrorManagedf(CodeNotFound, "%s not found", resource)
}

func NewAlreadyExistsError(resource string) *ErrorManaged {
	return NewErrorManagedf(CodeAlreadyExists, "%s already exists", resource)
}

func NewConflictError(message string) *ErrorManaged {
	return NewErrorManaged(CodeConflict, message)
}

func NewInvalidInputError(message string) *ErrorManaged {
	return NewErrorManaged(CodeInvalidInput, message)
}

func NewValidationError(message string) *ErrorManaged {
	return NewErrorManaged(CodeValidationError, message)
}

func NewBusinessRuleError(message string) *ErrorManaged {
	return NewErrorManaged(CodeBusinessRuleViolation, message)
}

func NewOperationNotAllowedError(message string) *ErrorManaged {
	return NewErrorManaged(CodeOperationNotAllowed, message)
}

func NewInternalServerError(message string) *ErrorManaged {
	return NewErrorManaged(CodeInternalServerError, message)
}

func NewServiceUnavailableError() *ErrorManaged {
	return NewErrorManaged(CodeServiceUnavailable, "Service temporarily unavailable")
}

func NewTimeoutError() *ErrorManaged {
	return NewErrorManaged(CodeTimeout, "Operation timed out")
}

func NewFileNotSupportedError(fileType string) *ErrorManaged {
	return NewErrorManagedf(CodeFileNotSupported, "File type '%s' is not supported", fileType)
}

func NewFileTooLargeError(maxSize string) *ErrorManaged {
	return NewErrorManagedf(CodeFileTooLarge, "File size exceeds maximum allowed size of %s", maxSize)
}

func NewUploadFailedError(message string) *ErrorManaged {
	return NewErrorManaged(CodeUploadFailed, message)
}

func IsErrorManaged(err error) bool {
	var managedErr *ErrorManaged
	return errors.As(err, &managedErr)
}

func GetErrorManaged(err error) *ErrorManaged {
	var managedErr *ErrorManaged
	if errors.As(err, &managedErr) {
		return managedErr
	}
	return nil
}
