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

// Package provider defines the interface for identity providers that can be
// used as a source of truth for SCIM synchronization.
package provider

import (
	"context"

	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
)

type (
	// Photo is one user's directory picture. Absent means the directory has
	// no picture for that user, and the bridge clears the stored version.
	// A zero Photo with a nil error means the picture could not be read,
	// and the bridge keeps the previous version.
	Photo struct {
		ETag   string
		Body   []byte
		MIME   string
		Absent bool
	}

	Provider interface {
		Name() string
		ListUsers(ctx context.Context) (scimclient.Users, error)
	}

	// PhotoSource is a provider that can read a directory picture. Providers
	// that cannot read pictures do not implement it. The bridge still
	// provisions the user when a picture cannot be read.
	PhotoSource interface {
		UserPhoto(ctx context.Context, user scimclient.User) (Photo, error)
	}
)
