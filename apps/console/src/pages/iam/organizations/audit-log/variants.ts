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
    toolbar: "flex flex-wrap items-center justify-between gap-2",
    tools: "flex flex-wrap items-center gap-2",
    action: "w-80 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0 max-sm:w-full",
    field: "w-44 shrink-0 max-sm:w-full",
  },
});

export const auditLogList = tv({
  slots: {
    root: "flex flex-col gap-4",
    toolbar: "flex flex-wrap items-center justify-between gap-2",
    tools: "flex flex-wrap items-center gap-2",
    action: "w-80 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0 max-sm:w-full",
    field: "w-44 shrink-0 max-sm:w-full",
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
    actor: "flex min-w-0 items-center gap-2",
    truncate: "min-w-0 truncate",
    action: "block max-w-full truncate",
  },
});

export const auditLogEntryDialog = tv({
  slots: {
    details: "flex flex-col gap-3",
    row: "flex flex-col gap-0.5",
    value: "min-w-0 break-all",
    actor: "flex min-w-0 flex-wrap items-center gap-2",
    target: "flex min-w-0 flex-col items-start gap-0.5",
  },
});

export const exportAuditLogDialog = tv({
  slots: {
    fields: "flex flex-col gap-4",
  },
});
