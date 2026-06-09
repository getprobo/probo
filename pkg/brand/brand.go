package brand

import (
	"embed"
	"io/fs"
)

var (
	//go:embed assets
	staticAssets embed.FS

	Assets fs.FS

	DefaultPoweredByLogoPath     string = "/api/files/v1/static/probo-gray-small.png"
	DefaultSenderCompanyLogoPath string = "/api/files/v1/static/probo.png"
)

func init() {
	var err error

	Assets, err = fs.Sub(staticAssets, "assets")
	if err != nil {
		panic(err)
	}
}
