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

package objectstorage

import (
	"context"
	"io"
	"time"
)

type (
	Store interface {
		PutObject(
			ctx context.Context,
			bucket string,
			key string,
			body io.Reader,
			options PutOptions,
		) error
		UploadObject(
			ctx context.Context,
			bucket string,
			key string,
			body io.Reader,
			options PutOptions,
		) error
		HeadObject(ctx context.Context, bucket string, key string) (Properties, error)
		GetObject(
			ctx context.Context,
			bucket string,
			key string,
			options GetOptions,
		) (*Object, error)
		DeleteObject(ctx context.Context, bucket string, key string) error
		PresignGetObject(
			ctx context.Context,
			bucket string,
			key string,
			options PresignOptions,
		) (string, error)
	}

	PutOptions struct {
		ContentLength      *int64
		ContentType        string
		CacheControl       string
		ContentDisposition string
		Metadata           map[string]string
	}

	GetOptions struct {
		IfNoneMatch     string
		IfModifiedSince time.Time
		Range           string
		FileSize        int64
	}

	PresignOptions struct {
		Expires                    time.Duration
		ResponseCacheControl       string
		ResponseContentType        string
		ResponseContentDisposition string
	}

	Properties struct {
		ContentLength int64
		ETag          string
		LastModified  time.Time
	}

	Object struct {
		Body                io.ReadCloser
		ContentLength       int64
		ContentRange        string
		ETag                string
		LastModified        time.Time
		NotModified         bool
		PartialContent      bool
		RangeNotSatisfiable bool
	}
)
