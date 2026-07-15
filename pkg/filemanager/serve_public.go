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

package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var ErrPublicFileNotFound = errors.New("public file not found")

func (s *Service) ServePublicFile(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	fileID gid.GID,
) error {
	file, err := s.GetPublicFile(ctx, fileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return ErrPublicFileNotFound
		}

		return err
	}

	conds := FileConditions{
		IfNoneMatch: r.Header.Get("If-None-Match"),
		IfRange:     r.Header.Get("If-Range"),
		Range:       r.Header.Get("Range"),
	}
	if ifModifiedSince := r.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
		if t, parseErr := http.ParseTime(ifModifiedSince); parseErr == nil {
			conds.IfModifiedSince = t
		}
	}

	obj, err := s.OpenFile(ctx, file, conds)
	if err != nil {
		return err
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Accept-Ranges", "bytes")

	if obj.ETag != "" {
		w.Header().Set("ETag", obj.ETag)
	}

	if obj.NotModified {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	if obj.RangeNotSatisfiable {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", file.FileSize))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)

		return nil
	}

	defer func() { _ = obj.Body.Close() }()

	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.ContentLength, 10))

	if !obj.LastModified.IsZero() {
		w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(http.TimeFormat))
	}

	if obj.PartialContent {
		w.Header().Set("Content-Range", obj.ContentRange)
		w.WriteHeader(http.StatusPartialContent)
	}

	if _, err := io.Copy(w, obj.Body); err != nil {
		return err
	}

	return nil
}
