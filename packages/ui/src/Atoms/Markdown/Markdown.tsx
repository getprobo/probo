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

import { isValidElement } from "react";
import ReactMarkdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import remarkGfm from "remark-gfm";

import { MermaidDiagram } from "./MermaidDiagram";

type Props = {
  content: string;
  /**
   * Absolute http(s) image origins allowed in addition to data: URIs, path-
   * relative srcs, and the page origin ('self'). Align with CSP img-src peers
   * (e.g. AppOrigin, file-storage origin, https://www.google.com).
   */
  allowedImageOrigins?: readonly string[];
};

const ABSOLUTE_SCHEME_RE = /^[a-zA-Z][a-zA-Z0-9+.-]*:/;

export function isAllowedMarkdownImageSrc(
  src: string,
  allowedImageOrigins: readonly string[] = [],
): boolean {
  if (src.startsWith("data:")) {
    return true;
  }

  // Path-relative / root-relative — not protocol-relative ("//host/...").
  if (!src.startsWith("//") && !ABSOLUTE_SCHEME_RE.test(src)) {
    return true;
  }

  let url: URL;
  try {
    const base
      = typeof globalThis.location?.href === "string"
        ? globalThis.location.href
        : "http://localhost/";
    url = new URL(src, base);
  } catch {
    return false;
  }

  if (url.protocol !== "http:" && url.protocol !== "https:") {
    return false;
  }

  const allowed = new Set(
    allowedImageOrigins.map(origin => origin.replace(/\/$/, "")),
  );
  if (typeof globalThis.location?.origin === "string") {
    allowed.add(globalThis.location.origin);
  }

  return allowed.has(url.origin);
}

export function Markdown({ content, allowedImageOrigins = [] }: Props) {
  return (
    <div className="prose prose-neutral">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw]}
        components={{
          a: ({ href, children, ...props }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              {...props}
            >
              {children}
            </a>
          ),
          img: ({ src, alt, ...props }) => {
            if (typeof src !== "string" || src === "") {
              return null;
            }

            if (!isAllowedMarkdownImageSrc(src, allowedImageOrigins)) {
              return null;
            }

            return <img src={src} alt={alt ?? ""} {...props} />;
          },
          pre: ({ children, ...props }) => {
            const child = isValidElement<{
              className?: string;
              children?: string;
            }>(children)
              ? children
              : null;

            if (
              child?.type === "code"
              && child.props.className === "language-mermaid"
              && typeof child.props.children === "string"
            ) {
              return <MermaidDiagram chart={child.props.children} />;
            }

            return (
              <pre
                className="border border-border-solid rounded p-4 bg-transparent font-mono text-sm overflow-x-auto text-inherit"
                {...props}
              >
                {children}
              </pre>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
