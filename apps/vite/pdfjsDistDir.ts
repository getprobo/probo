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

import { createRequire } from "node:module";
import { dirname } from "node:path";

import type { Alias } from "vite";

// Hoisted pdfjs-dist (pdf-parse) differs from react-pdf's copy, and its modern
// build calls Promise.try with no polyfill. Resolve to react-pdf's legacy build.
export function pdfjsDistDir(importMetaUrl: string): string {
  const reactPdfRequire = createRequire(
    createRequire(importMetaUrl).resolve("react-pdf/package.json"),
  );
  return dirname(reactPdfRequire.resolve("pdfjs-dist/package.json"));
}

export function pdfjsAliases(importMetaUrl: string): Alias[] {
  const dir = pdfjsDistDir(importMetaUrl);
  return [
    { find: /^pdfjs-dist$/, replacement: `${dir}/legacy/build/pdf.mjs` },
    { find: /^pdfjs-dist\/(build|web)\//, replacement: `${dir}/legacy/$1/` },
    { find: /^pdfjs-dist\//, replacement: `${dir}/` },
  ];
}

// Dev dep optimization prebundles this import and drops the `?url` default export.
export const pdfjsWorkerUrlId = "pdfjs-dist/build/pdf.worker.min.mjs?url";
