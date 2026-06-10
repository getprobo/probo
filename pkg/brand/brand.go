package brand

import (
	"embed"
	"io/fs"
)

var (
	//go:embed assets
	staticAssets embed.FS

	DefaultPoweredByLogoPath     string = "/api/files/v1/static/probo-gray-small.png"
	DefaultSenderCompanyLogoPath string = "/api/files/v1/static/probo.png"
)

func NewAssets() (fs.FS, error) {
	return fs.Sub(staticAssets, "assets")
}
