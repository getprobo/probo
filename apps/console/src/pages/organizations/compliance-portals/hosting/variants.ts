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
  base: "flex flex-col gap-6",
});

export const hostingPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    section: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
  },
});

export const statusSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    headingRow: "flex items-center justify-between",
    grid: "grid grid-cols-3 gap-3 max-lg:grid-cols-2 max-sm:grid-cols-1",
  },
});

export const domainsSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    grid: "grid grid-cols-3 gap-3 max-lg:grid-cols-2 max-sm:grid-cols-1",
  },
});

export const domainCard = tv({
  slots: {
    lead: "relative z-1 flex min-w-0 flex-1 items-center gap-3",
    subtitle: "min-w-0 truncate",
    copy: "flex flex-col gap-2",
    heading: "flex items-center justify-between gap-3",
    identity: "flex min-w-0 flex-wrap items-center gap-2",
    dns: "flex flex-col gap-2",
    record: "flex flex-col gap-2 rounded-2 bg-sand-3 p-3",
    recordHeader: "flex items-center justify-between gap-2",
    recordField: "flex flex-col gap-1",
    recordValue: "flex min-w-0 items-center gap-1",
    code: "min-w-0 flex-1 break-all",
    wide: "col-span-3 max-lg:col-span-2 max-sm:col-span-1",
  },
});

export const customDomainForm = tv({
  slots: {
    form: "flex flex-1 flex-col",
    field: "w-full max-w-lg",
    actions: "self-start",
  },
});

export const deleteDomainDialog = tv({
  slots: {
    form: "flex flex-col gap-4",
    fields: "flex flex-col gap-4",
  },
});
