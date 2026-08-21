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

export const hostingPage = tv({
  base: "flex flex-col gap-8",
});

export const hostingPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-8",
    section: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    empty: "flex flex-col items-center gap-5 py-8 text-center",
  },
});

export const statusSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    headingRow: "flex items-center justify-between",
    cardBody: "flex flex-col gap-4",
    row: "flex items-center justify-between gap-4",
    rowText: "flex min-w-0 flex-col gap-1",
  },
});

export const domainsSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    empty: "flex flex-col items-center gap-5 py-8 text-center",
    emptyIcon: "text-sand-a8 [&_svg]:size-6",
  },
});

export const domainCard = tv({
  slots: {
    identity: "flex flex-wrap items-center gap-2",
    actions: "flex shrink-0 items-center gap-2",
  },
});

export const domainFormDialog = tv({
  slots: {
    form: "flex flex-col gap-4",
    fields: "flex flex-col gap-4",
    examples: "rounded-4 bg-sand-3 p-4",
  },
});

export const domainDetailsDialog = tv({
  slots: {
    titleRow: "flex flex-wrap items-center gap-3",
    body: "flex flex-col gap-6",
    calloutBody: "flex flex-col gap-1",
    dns: "flex flex-col gap-4",
    records: "flex flex-col gap-3",
    record: "flex flex-col gap-2 rounded-4 bg-sand-3 p-4",
    recordHeader: "flex items-center justify-between",
    recordField: "flex flex-col gap-1",
    recordValue: "flex min-w-0 items-center gap-2",
    code: "min-w-0 flex-1 break-all",
  },
});
