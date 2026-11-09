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

package richtext

import (
	"go.probo.inc/probo/pkg/prosemirror/html"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
	"go.probo.inc/probo/pkg/richtext/attachment"
	"go.probo.inc/probo/pkg/richtext/mention"
)

var (
	// DocumentHTML renders DocumentSchema nodes to HTML.
	DocumentHTML = html.NewRenderer(
		DocumentSchema,
		append(html.Nodes(), append(attachment.HTMLNodes(), mention.HTMLNodes()...)...),
		html.Marks(),
	)

	// DocumentMarkdown renders DocumentSchema nodes to Markdown.
	DocumentMarkdown = markdown.NewRenderer(
		DocumentSchema,
		DocumentHTML,
		append(markdown.Nodes(), append(attachment.MarkdownNodes(), mention.MarkdownNodes()...)...),
		markdown.Marks(),
	)

	// CommentHTML renders CommentSchema nodes to HTML.
	CommentHTML = html.NewRenderer(
		CommentSchema,
		append(html.Nodes(), mention.HTMLNodes()...),
		html.Marks(),
	)

	// CommentMarkdown renders CommentSchema nodes to Markdown.
	CommentMarkdown = markdown.NewRenderer(
		CommentSchema,
		CommentHTML,
		append(markdown.Nodes(), mention.MarkdownNodes()...),
		markdown.Marks(),
	)
)
