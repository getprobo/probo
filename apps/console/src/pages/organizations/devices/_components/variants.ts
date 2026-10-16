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

export const devicesPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    header: "flex flex-wrap items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-1",
    title: "flex min-w-0 items-center gap-2.5",
    actions: "flex items-center gap-2",
    more: "flex justify-center",
    results: "flex flex-col gap-4",
    list: "flex flex-col gap-4",
    sections: "flex flex-col gap-8",
    section: "flex flex-col gap-4",
    views: "flex flex-wrap items-center",
    tools: "flex flex-wrap items-center justify-between gap-2",
    filters: "flex flex-wrap items-center justify-end gap-2",
    tableFilters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
    platformFilter: "w-max shrink-0",
    versionFilter: "w-max shrink-0",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    footer: "flex items-center justify-end gap-4",
  },
  variants: {
    pending: {
      true: {
        section: "pointer-events-none opacity-50",
        list: "pointer-events-none opacity-50",
      },
    },
  },
});

export const deviceCardFields = tv({
  slots: {
    root: "grid grid-cols-3 gap-4 max-lg:grid-cols-2 max-sm:grid-cols-1 *:min-w-0",
    property: "flex min-w-0 flex-col gap-1",
  },
});

export const postureChecksPopover = tv({
  slots: {
    trigger: "inline-flex cursor-pointer border-0 bg-transparent p-0",
    popup: "w-max",
    list: "m-0 flex list-none flex-col p-0 text-1",
    item: "flex items-center gap-1",
  },
});

export const postureReport = tv({
  slots: {
    time: "inline-flex cursor-pointer items-center gap-2 whitespace-nowrap border-0 bg-transparent p-0 text-inherit",
    correlation: "inline-flex min-w-0 items-center gap-1",
    checks: "grid grid-cols-2 gap-x-8 gap-y-2 py-2 ps-6",
    check: "flex items-center gap-4",
    checkLabel: "min-w-40",
  },
});

export const enrollment = tv({
  slots: {
    root: "flex flex-col gap-4",
    copyField: "flex flex-col gap-1",
    copyValue: "flex min-w-0 items-center gap-1",
    copyCode: "min-w-0 flex-1 truncate",
  },
});
