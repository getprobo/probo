// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// GetPublicFile loads a public file record by ID.
func (s *Service) GetPublicFile(
	ctx context.Context,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadPublicByID(ctx, conn, fileID); err != nil {
				return fmt.Errorf("cannot load public file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}
