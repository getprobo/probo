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

// Documents page shell: a centered content column below the header band. The
// `busy` variant dims the current results while a filtered slice refetches.
export const documentsLayout = tv({
  slots: {
    page: "flex w-full flex-col items-center px-8 pt-8 pb-28 max-md:px-4",
    results: "flex w-full max-w-5xl flex-col gap-8 transition-opacity duration-150",
  },
  variants: {
    busy: {
      true: { results: "opacity-60" },
      false: {},
    },
  },
  defaultVariants: {
    busy: false,
  },
});

// Bottom selection action bar, shown while rows are selected. A fixed, blurred
// band whose inner column aligns with the results column above it.
export const selectionBar = tv({
  slots: {
    bar: "fixed inset-x-0 bottom-0 z-10 border-t border-sand-a3 bg-sand-1/80 px-8 py-4 backdrop-blur max-md:px-4",
    inner: "mx-auto flex w-full max-w-5xl items-center justify-between gap-4",
    actions: "flex items-center gap-2",
  },
});
