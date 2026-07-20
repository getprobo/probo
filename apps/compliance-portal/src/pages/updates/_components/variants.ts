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

import { tv } from "tailwind-variants/lite";

// Updates list surface: a white card holding divider-separated rows. The card
// dims while a page change is loading. Slots are shared by the list and its
// skeleton so the loading placeholder matches the real layout.
export const updatesList = tv({
  slots: {
    card: "overflow-hidden rounded-5 border border-sand-3 bg-sand-1 transition-opacity duration-150",
    rows: "divide-y divide-sand-a2",
  },
  variants: {
    busy: {
      true: { card: "opacity-60" },
      false: {},
    },
  },
  defaultVariants: {
    busy: false,
  },
});

// Update detail article layout: the toolbar row (back link + subscribe), the
// centered content column, and the article body with its gold metadata row.
// Slots are shared by the detail page and its skeleton.
export const updateArticle = tv({
  slots: {
    toolbar: "flex w-full items-center justify-between gap-4 max-sm:flex-col max-sm:items-stretch",
    content: "flex w-full flex-col items-center px-8 py-8 max-md:px-4",
    article: "flex w-full max-w-2xl flex-col gap-4",
    meta: "flex items-center gap-1.5",
    metaIcon: "size-4 text-gold-9",
    body: "block whitespace-pre-wrap",
  },
});
