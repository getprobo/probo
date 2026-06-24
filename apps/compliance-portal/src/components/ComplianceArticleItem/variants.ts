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

// Compliance article list row (Figma "Compliance Article Item"): a leading
// icon, a title (+ optional eyebrow), and right-aligned meta. Designed to sit
// inside a list container that draws the dividers (e.g. `divide-y`), so rows
// stay divider-agnostic whether or not each is wrapped in a link. Slots are
// shared by the row and its skeleton.
export const complianceArticleItem = tv({
  slots: {
    root: "flex w-full items-center gap-4 px-8 py-4",
    icon: "flex size-6 shrink-0 items-center justify-center text-gold-9 [&_svg]:size-full",
    content: "flex min-w-0 flex-1 flex-col gap-1",
    meta: "shrink-0",
    iconPlaceholder: "size-6 shrink-0 animate-pulse rounded-2 bg-sand-3",
  },
});
