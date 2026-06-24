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

package browser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSitemapXML(t *testing.T) {
	t.Parallel()

	t.Run(
		"valid urlset with multiple URLs",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.com/</loc></url>
  <url><loc>https://example.com/about</loc></url>
  <url><loc>https://example.com/contact</loc></url>
</urlset>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			require.Len(t, urls, 3)
			assert.Equal(t, "https://example.com/", urls[0])
			assert.Equal(t, "https://example.com/about", urls[1])
			assert.Equal(t, "https://example.com/contact", urls[2])
		},
	)

	t.Run(
		"valid sitemapindex with sitemap locations",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>https://example.com/sitemap-pages.xml</loc></sitemap>
  <sitemap><loc>https://example.com/sitemap-posts.xml</loc></sitemap>
</sitemapindex>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			require.Len(t, urls, 2)
			assert.Equal(t, "https://example.com/sitemap-pages.xml", urls[0])
			assert.Equal(t, "https://example.com/sitemap-posts.xml", urls[1])
		},
	)

	t.Run(
		"empty urlset returns empty slice",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
</urlset>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			assert.Empty(t, urls)
		},
	)

	t.Run(
		"malformed XML returns error",
		func(t *testing.T) {
			t.Parallel()

			xml := `<urlset><url><loc>https://example.com/</loc></url`

			_, err := parseSitemapXML(strings.NewReader(xml))

			assert.Error(t, err)
		},
	)

	t.Run(
		"urlset without namespace",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset>
  <url><loc>https://example.com/page1</loc></url>
  <url><loc>https://example.com/page2</loc></url>
</urlset>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			require.Len(t, urls, 2)
			assert.Equal(t, "https://example.com/page1", urls[0])
			assert.Equal(t, "https://example.com/page2", urls[1])
		},
	)

	t.Run(
		"trims whitespace in loc elements",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset>
  <url><loc>  https://example.com/padded  </loc></url>
</urlset>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			require.Len(t, urls, 1)
			assert.Equal(t, "https://example.com/padded", urls[0])
		},
	)

	t.Run(
		"skips empty loc elements",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset>
  <url><loc></loc></url>
  <url><loc>https://example.com/valid</loc></url>
  <url><loc>   </loc></url>
</urlset>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			require.Len(t, urls, 1)
			assert.Equal(t, "https://example.com/valid", urls[0])
		},
	)

	t.Run(
		"empty reader returns empty slice",
		func(t *testing.T) {
			t.Parallel()

			urls, err := parseSitemapXML(strings.NewReader(""))

			require.NoError(t, err)
			assert.Empty(t, urls)
		},
	)

	t.Run(
		"urlset with additional elements besides loc",
		func(t *testing.T) {
			t.Parallel()

			xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/page</loc>
    <lastmod>2024-01-01</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>`

			urls, err := parseSitemapXML(strings.NewReader(xml))

			require.NoError(t, err)
			require.Len(t, urls, 1)
			assert.Equal(t, "https://example.com/page", urls[0])
		},
	)
}
