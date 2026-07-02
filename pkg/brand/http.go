package brand

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"path"
)

// StaticPathPrefix is the single source of truth for the public URL path under
// which brand assets are served. It must stay in sync with the route mounted by
// the files API handler (/api -> /files/v1 -> /static/{file}).
const StaticPathPrefix = "/api/files/v1/static"

// StaticPath returns the public URL path that serves the named brand asset.
func StaticPath(name string) string {
	return path.Join(StaticPathPrefix, name)
}

// Assets provides access to the embedded brand assets together with their
// precomputed content-hash ETags.
type Assets struct {
	fs    fs.FS
	etags map[string]string
}

// NewAssets builds an Assets from the embedded brand files. It panics on error
// because the assets are embedded at build time and a failure here is a
// programming error, not a runtime condition.
func NewAssets() *Assets {
	assetsFS, err := fs.Sub(staticAssets, "assets")
	if err != nil {
		panic(fmt.Errorf("cannot open brand assets: %w", err))
	}

	for _, name := range requiredAssets {
		if _, err := fs.Stat(assetsFS, name); err != nil {
			panic(fmt.Errorf("missing required brand asset %q: %w", name, err))
		}
	}

	return &Assets{
		fs:    assetsFS,
		etags: mustComputeAssetETags(assetsFS),
	}
}

// Stat returns file information for the named asset.
func (a *Assets) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(a.fs, name)
}

// ServeAssets writes the named asset to the response. It delegates to
// http.ServeFileFS, which sets Content-Type, handles range requests, and honors
// a pre-set ETag response header for If-None-Match (304 Not Modified).
func (a *Assets) ServeAssets(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if etag, ok := a.etags[name]; ok {
		w.Header().Set("ETag", `"`+etag+`"`)
	}

	http.ServeFileFS(w, r, a.fs, name)
}

// mustComputeAssetETags precomputes content-hash ETags for every embedded
// brand asset, mirroring the MD5 walk used by the SPA static handler. It panics
// on error because the assets are embedded at build time and a failure here is
// a programming error, not a runtime condition.
func mustComputeAssetETags(assets fs.FS) map[string]string {
	etags := make(map[string]string)

	err := fs.WalkDir(
		assets,
		".",
		func(name string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			content, err := fs.ReadFile(assets, name)
			if err != nil {
				return fmt.Errorf("cannot read brand asset %q: %w", name, err)
			}

			hash := md5.Sum(content)
			etags[name] = hex.EncodeToString(hash[:])

			return nil
		},
	)
	if err != nil {
		panic(fmt.Errorf("cannot compute brand asset etags: %w", err))
	}

	return etags
}
