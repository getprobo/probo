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

export const auditLogPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
  },
});

export const auditLogPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    filters: "grid grid-cols-4 gap-3 max-lg:grid-cols-2 max-sm:grid-cols-1",
  },
});

export const auditLogList = tv({
  slots: {
    root: "flex flex-col gap-4",
    filters: "grid grid-cols-4 gap-3 max-lg:grid-cols-2 max-sm:grid-cols-1",
    filter: "min-w-0",
    results: "transition-opacity",
    empty: "flex flex-col items-center py-8 text-center",
    pager: "flex justify-center",
  },
  variants: {
    pending: {
      true: {
        results: "opacity-60",
      },
    },
  },
});

export const auditLogListItem = tv({
  slots: {
    open: [
      "min-w-0 appearance-none border-0 bg-transparent p-0 text-left font-inherit text-sand-12 pointer-events-auto cursor-pointer",
      "after:absolute after:inset-0 after:content-['']",
      "outline-none focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    actor: "flex min-w-0 flex-col items-start gap-1",
    target: "flex min-w-0 flex-col items-start gap-0.5",
    truncate: "min-w-0 max-w-full truncate",
    mono: "min-w-0 max-w-full truncate font-mono",
  },
});

export const auditLogEntryDialog = tv({
  slots: {
    details: "flex flex-col gap-3",
    row: "flex flex-col gap-0.5",
    value: "min-w-0 break-all",
    actor: "flex min-w-0 flex-wrap items-center gap-2",
  },
});

export const exportAuditLogDialog = tv({
  slots: {
    fields: "flex flex-col gap-4",
  },
});
