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

import type { PageLayout } from "@probo/ui/src/RichEditor/pagination";

// Pages match the PDF export: A4 with 2.5 cm margins (pkg/docgen/template.html),
// in CSS pixels at 96 dpi.
const pixelsPerCentimeter = 96 / 2.54;
const a4ShortSide = 794;
const a4LongSide = 1123;

export const documentPageMargin = Math.round(2.5 * pixelsPerCentimeter);
export const documentPageGap = 24;

export type DocumentPageSize = {
  width: number;
  height: number;
};

export function documentPageSize(orientation: string): DocumentPageSize {
  return orientation === "LANDSCAPE"
    ? { width: a4LongSide, height: a4ShortSide }
    : { width: a4ShortSide, height: a4LongSide };
}

export function documentPageLayout(size: DocumentPageSize): PageLayout {
  return {
    pageHeight: size.height,
    marginTop: documentPageMargin,
    marginBottom: documentPageMargin,
    gap: documentPageGap,
  };
}

export function documentPagesHeight(size: DocumentPageSize, pages: number): number {
  return pages * size.height + Math.max(0, pages - 1) * documentPageGap;
}
