package brand

import (
	"embed"
)

//go:embed assets
var staticAssets embed.FS

// Logical brand assets. These filenames must exist under assets/; NewAssets
// verifies their presence at startup so a rename or removal fails fast instead
// of silently producing a 404 in the consumers (e.g. emails).
const (
	PoweredByLogo     = "probo-gray-small.png"
	SenderCompanyLogo = "probo.png"
)

// requiredAssets are validated to exist whenever an Assets is constructed.
var requiredAssets = []string{PoweredByLogo, SenderCompanyLogo}
