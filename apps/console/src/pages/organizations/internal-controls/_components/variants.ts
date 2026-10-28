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

export const internalControlsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    header: "flex flex-wrap items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-1",
    title: "flex min-w-0 items-center gap-2.5",
    actions: "flex items-center gap-2",
    results: "flex flex-col gap-4",
    list: "flex flex-col gap-4",
    views: "flex flex-wrap items-center",
    tools: "flex flex-wrap items-center justify-between gap-2",
    tableFilters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-44 shrink-0",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    pager: "flex justify-end",
    empty: "flex flex-col items-center gap-3 px-6 py-12 text-center",
    emptyActions: "flex flex-wrap items-center justify-center gap-2",
    identity: "flex min-w-0 flex-col items-start gap-1",
    owner: "inline-flex min-w-0 items-center gap-2",
    due: "inline-flex items-center gap-1",
  },
  variants: {
    pending: {
      true: {
        list: "pointer-events-none opacity-50",
      },
    },
  },
});
