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
