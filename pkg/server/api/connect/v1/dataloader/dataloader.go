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

package dataloader

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vikstrous/dataloadgen"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
)

type (
	ctxKey struct{ name string }

	Loaders struct {
		AvatarFileForIdentity *dataloadgen.Loader[gid.GID, *coredata.File]
		AvatarFileForProfile  *dataloadgen.Loader[gid.GID, *coredata.File]
	}

	batchFetcher struct {
		iam *iam.Service
	}
)

var loadersKey = &ctxKey{name: "dataloaders"}

func FromContext(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func NewMiddleware(iamSvc *iam.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				f := &batchFetcher{iam: iamSvc}
				ctx := context.WithValue(r.Context(), loadersKey, f.newLoaders())
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

func (f *batchFetcher) newLoaders() *Loaders {
	return &Loaders{
		AvatarFileForIdentity: dataloadgen.NewMappedLoader(f.fetchAvatarFilesForIdentities),
		AvatarFileForProfile:  dataloadgen.NewMappedLoader(f.fetchAvatarFilesForProfiles),
	}
}

func (f *batchFetcher) fetchAvatarFilesForIdentities(
	ctx context.Context,
	keys []gid.GID,
) (map[gid.GID]*coredata.File, error) {
	files, err := f.iam.AccountService.AvatarFiles(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load identity avatars: %w", err)
	}

	return files, nil
}

func (f *batchFetcher) fetchAvatarFilesForProfiles(
	ctx context.Context,
	keys []gid.GID,
) (map[gid.GID]*coredata.File, error) {
	files, err := f.iam.AccountService.AvatarFilesForProfiles(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("cannot batch load profile avatars: %w", err)
	}

	return files, nil
}
