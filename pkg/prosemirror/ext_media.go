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

package prosemirror

// ImageAttrs contains attributes for image nodes.
// FileID is set for an uploaded image. Src is set for a remote image.
// Width is the display width as a percentage of the column.
type ImageAttrs struct {
	FileID *string `json:"fileId,omitempty"`
	Src    string  `json:"src"`
	Alt    *string `json:"alt,omitempty"`
	Title  *string `json:"title,omitempty"`
	Width  *int    `json:"width,omitempty"`
}

// ImageAttrs parses and returns the image attributes.
func (n Node) ImageAttrs() (ImageAttrs, error) {
	return decodeAttrs(n.Attrs, ImageAttrs{}, "image")
}

// FileAttrs contains attributes for file attachment nodes.
type FileAttrs struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

// FileAttrs parses and returns the file attachment attributes.
func (n Node) FileAttrs() (FileAttrs, error) {
	return decodeAttrs(n.Attrs, FileAttrs{}, "file")
}
