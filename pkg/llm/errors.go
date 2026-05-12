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

package llm

import (
	"fmt"
	"time"
)

type (
	ErrRateLimit struct {
		RetryAfter time.Duration
		Err        error
	}

	ErrContextLength struct {
		MaxTokens int
		Err       error
	}

	ErrContentFilter struct {
		Err error
	}

	ErrAuthentication struct {
		Err error
	}

	// ErrStreamingRequired is returned by a provider when a non-streaming
	// request must be retried with the streaming endpoint (e.g. Anthropic
	// requires streaming for responses that may take longer than 10 minutes).
	ErrStreamingRequired struct {
		Err error
	}
)

func (e *ErrRateLimit) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited (retry after %s): %v", e.RetryAfter, e.Err)
	}
	return fmt.Sprintf("rate limited: %v", e.Err)
}

func (e *ErrRateLimit) Unwrap() error { return e.Err }

func (e *ErrContextLength) Error() string {
	if e.MaxTokens > 0 {
		return fmt.Sprintf("context length exceeded (max %d tokens): %v", e.MaxTokens, e.Err)
	}
	return fmt.Sprintf("context length exceeded: %v", e.Err)
}

func (e *ErrContextLength) Unwrap() error { return e.Err }

func (e *ErrContentFilter) Error() string {
	return fmt.Sprintf("content filtered: %v", e.Err)
}

func (e *ErrContentFilter) Unwrap() error { return e.Err }

func (e *ErrAuthentication) Error() string {
	return fmt.Sprintf("authentication failed: %v", e.Err)
}

func (e *ErrAuthentication) Unwrap() error { return e.Err }

func (e *ErrStreamingRequired) Error() string {
	return fmt.Sprintf("streaming is required: %v", e.Err)
}

func (e *ErrStreamingRequired) Unwrap() error { return e.Err }
