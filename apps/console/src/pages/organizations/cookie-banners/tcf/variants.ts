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

export const tcfPage = tv({
  base: "flex flex-col gap-6",
});

export const tcfPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-8",
    intro: "flex flex-col gap-3",
    title: "w-40",
    description: "w-96",
    stats: "mt-6 grid grid-cols-4 gap-3 max-lg:grid-cols-2",
    list: "flex flex-col gap-4",
    tools: "flex w-full items-center justify-between gap-2",
    search: "w-60 max-sm:min-w-0 max-sm:flex-1",
    filter: "w-48 shrink-0",
  },
});

export const tcfSection = tv({
  slots: {
    root: "flex flex-col gap-8",
    intro: "flex flex-col gap-3",
    list: "flex flex-col gap-4",
    tools: "flex w-full items-center justify-between gap-2",
    search: "w-60 max-sm:min-w-0 max-sm:flex-1",
    filter: "w-48 shrink-0",
    pager: "flex justify-center pt-4",
    results: "flex flex-col transition-opacity",
  },
  variants: {
    pending: {
      true: {
        results: "opacity-60",
      },
    },
  },
});

export const gvlVendorStats = tv({
  slots: {
    root: "mt-6 grid grid-cols-4 gap-3 max-lg:grid-cols-2",
    card: "flex h-full flex-col items-start gap-2",
    comparison: "flex items-baseline gap-4",
    valueRow: "flex items-baseline gap-1.5",
    value: "font-mono",
    footer: "mt-auto pt-3",
  },
});

export const gvlVendorListItem = tv({
  slots: {
    name: "min-w-0 truncate",
    meta: "min-w-0 truncate",
    trailing: "ml-auto flex shrink-0 items-center",
  },
});

export const gvlVendorListEmpty = tv({
  slots: {
    root: "flex flex-col items-center gap-5 py-8 text-center",
    icon: "text-sand-a8 [&_svg]:size-6",
    body: "flex max-w-xs flex-col items-center gap-2",
  },
});
