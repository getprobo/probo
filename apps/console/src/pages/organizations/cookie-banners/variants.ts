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

export const cookieBannerPage = tv({
  base: "flex flex-col gap-6",
});

export const cookieBannerPageHeader = tv({
  slots: {
    root: "flex flex-col gap-2",
  },
});

export const cookieBannerList = tv({
  slots: {
    root: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
    results: "transition-opacity",
    pager: "flex justify-center",
    empty: "flex flex-col items-center gap-1 py-8 text-center",
  },
  variants: {
    pending: {
      true: {
        results: "opacity-60",
      },
    },
  },
});

export const cookieBannerListSkeleton = tv({
  slots: {
    root: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
  },
});

export const trackerPatternListItem = tv({
  slots: {
    name: "flex min-w-0 flex-col gap-0.5",
    heading: "flex min-w-0 items-center gap-2",
    title: "min-w-0 truncate",
    description: "min-w-0 line-clamp-1",
    date: "whitespace-nowrap",
    actions: "flex items-center gap-1",
    edit: "flex flex-col gap-3",
    editFields: "flex items-end gap-2",
    editField: "flex min-w-0 flex-1 flex-col gap-1",
  },
  variants: {
    excluded: {
      true: {
        name: "opacity-50",
        title: "line-through",
      },
    },
  },
  defaultVariants: {
    excluded: false,
  },
});

export const trackerResourceListItem = tv({
  slots: {
    origin: "flex min-w-0 flex-col gap-0.5",
    title: "min-w-0 truncate",
    description: "min-w-0 line-clamp-1",
    path: "min-w-0 truncate font-mono",
    date: "whitespace-nowrap",
    actions: "flex items-center gap-1",
    edit: "flex items-end gap-2",
    editField: "min-w-0 flex-1",
  },
  variants: {
    excluded: {
      true: {
        origin: "opacity-50",
        title: "line-through",
        path: "opacity-50",
      },
    },
  },
  defaultVariants: {
    excluded: false,
  },
});

export const moveToCategorySelect = tv({
  slots: {
    root: "min-w-36",
  },
});
