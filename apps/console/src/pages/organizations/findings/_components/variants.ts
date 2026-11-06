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

export const findingColumnWidth = {
  reference: "6.5rem",
  title: "16rem",
  kind: "11rem",
  status: "8.5rem",
  priority: "6.5rem",
  owner: "10rem",
  due: "8rem",
  actions: "3rem",
} as const;

export const findingsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-wrap items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-1",
    title: "flex min-w-0 items-center gap-2.5",
    actions: "flex items-center gap-2",
    results: "flex flex-col gap-4",
    list: "flex flex-col gap-4",
    tableFilters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-44 shrink-0",
    pager: "flex justify-end",
    empty: "flex flex-col items-center gap-3 px-6 py-12 text-center",
    emptyActions: "flex flex-wrap items-center justify-center gap-2",
    nameCell: "min-w-0 max-w-0 whitespace-normal [overflow-wrap:anywhere]",
    name: "w-full min-w-0 whitespace-normal [overflow-wrap:anywhere] line-clamp-2",
    owner: "inline-flex min-w-0 max-w-full items-center gap-2",
    ownerName: "min-w-0 truncate",
    form: "flex flex-col gap-4",
    formRow: "grid gap-4 sm:grid-cols-2",
    narrative: "grid gap-4 md:grid-cols-2",
    narrativeWide: "md:col-span-2",
    links: "flex flex-col gap-2",
    linkRow: "flex items-center justify-between gap-2",
    linkText: "flex min-w-0 flex-col",
    side: "flex flex-col gap-4",
    section: "flex flex-col gap-3",
    pendingLink: "flex flex-col gap-2",
    pendingActions: "flex justify-end gap-2",
    back: "self-start",
    ownerIcon: "size-4 shrink-0 text-sand-8",
    skeletonTitle: "w-40",
    skeletonCount: "w-4",
    skeletonIntro: "w-80",
    skeletonAction: "w-28",
  },
  variants: {
    pending: {
      true: {
        list: "pointer-events-none opacity-50",
      },
    },
  },
});

export const findingDescription = tv({
  slots: {
    input: [
      "border-0 bg-transparent shadow-none focus-within:ring-0",
      "[&_textarea]:min-h-32 [&_textarea]:bg-transparent [&_textarea]:px-0 [&_textarea]:py-1",
    ],
  },
});

export const auditFilterSelect = tv({
  slots: {
    loadMore: "w-full",
  },
});

export const findingTitleField = tv({
  slots: {
    heading: "min-w-0",
    input: [
      "w-full min-w-0 bg-transparent text-6 font-medium text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
  },
});

export const findingKindIcon = tv({
  base: "size-4 shrink-0",
  variants: {
    kind: {
      MAJOR_NONCONFORMITY: "text-red-11",
      MINOR_NONCONFORMITY: "text-amber-11",
      OBSERVATION: "text-sand-11",
      EXCEPTION: "text-sand-11",
    },
  },
});

export const findingStatusIcon = tv({
  base: "size-4 shrink-0",
  variants: {
    status: {
      OPEN: "text-red-11",
      IN_PROGRESS: "text-amber-11",
      CLOSED: "text-green-11",
      MITIGATED: "text-green-11",
      RISK_ACCEPTED: "text-amber-11",
      FALSE_POSITIVE: "text-sand-10",
    },
  },
});

export const findingIdentifiedIcon = tv({
  base: "size-4 shrink-0 text-indigo-11",
});

export const findingRisk = tv({
  slots: {
    root: "flex w-full flex-col items-start gap-2",
    field: "w-full",
    icon: "size-4 shrink-0 text-amber-11",
  },
});
