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

export const accessInventoryGrantsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
  },
});

export const accessInventoryGrantsPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    tools: "flex flex-wrap items-center justify-between gap-2",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-44 shrink-0",
  },
});

export const accessInventoryGrantsList = tv({
  slots: {
    root: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-44 shrink-0",
    include: "flex items-center gap-2",
    results: "transition-opacity",
    empty: "flex flex-col items-center py-8 text-center",
    pager: "flex justify-center",
    principal: "flex min-w-0 flex-col gap-0.5",
    resource: "flex min-w-0 flex-col gap-0.5",
  },
  variants: {
    pending: {
      true: {
        results: "opacity-60",
      },
    },
    disappeared: {
      true: {
        principal: "opacity-60",
        resource: "opacity-60",
      },
    },
  },
});
