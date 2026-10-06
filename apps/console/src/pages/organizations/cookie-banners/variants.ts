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

export const cookieBannerConfigLayout = tv({
  slots: {
    root: "flex flex-col gap-6",
    lead: "flex min-w-0 items-center gap-2 whitespace-normal",
    title: "min-w-0 truncate",
    meta: "flex shrink-0 items-center gap-1.5",
    id: "font-mono",
    version: "shrink-0 font-mono",
  },
});

export const cookieBannerLifecycleSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    row: "flex flex-wrap items-center justify-between gap-3",
    status: "flex items-center gap-2",
    actions: "flex flex-wrap items-center justify-end gap-2",
  },
});

export const cookieBannerInstallSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    snippetWrap: "relative",
    snippet: "overflow-x-auto pr-9 font-mono text-2 text-sand-12",
    snippetCopy: "absolute top-0 right-0",
  },
});

export const cookieBannerSettingsSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    card: "flex flex-col gap-6",
    block: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    fields: "flex flex-col gap-4",
    field: "flex flex-col gap-1",
    pair: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 *:min-w-0",
    toggle: "flex items-start gap-3",
    toggleCopy: "flex flex-col gap-1",
    actions: "flex justify-end",
  },
});

export const cookieBannerTranslationsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    toolbar: "flex items-center gap-3",
    language: "w-64 max-sm:w-full",
    form: "flex flex-col gap-8",
    section: "flex flex-col gap-4",
    split: "grid grid-cols-2 gap-6 max-lg:grid-cols-1",
    fields: "flex flex-col gap-4",
    pair: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 *:min-w-0",
    preview: "flex items-start justify-center rounded-3 bg-sand-3 p-6",
    categoryGrid: "grid grid-cols-2 gap-4 max-sm:grid-cols-1",
    actions: "flex justify-end",
  },
});

export const cookieBannerList = tv({
  slots: {
    root: "flex flex-col gap-4",
    views: "flex flex-wrap items-center",
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
    views: "flex flex-wrap items-center gap-2",
    view: "h-8 w-28 animate-pulse rounded-2 bg-sand-3",
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
    title: "min-w-0 truncate font-mono",
    info: "relative z-1 shrink-0 pointer-events-auto",
    detail: "flex flex-col gap-1",
    date: "whitespace-nowrap",
    actions: "flex items-center gap-1",
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
    root: "w-full min-w-36",
  },
});

export const trackerPatternDetailHeader = tv({
  slots: {
    root: "flex flex-col gap-4",
    back: "self-start",
    bar: "flex items-start justify-between gap-4",
    titleRow: "flex min-w-0 flex-1 flex-wrap items-center gap-2",
    title: "min-w-0",
    badges: "flex flex-wrap items-center gap-1.5",
    actions: "flex shrink-0 items-center gap-2",
  },
});

export const trackerMaxAgeField = tv({
  slots: {
    root: "flex w-full gap-2",
    value: "min-w-0 flex-1",
    unit: "w-36 shrink-0",
  },
});

export const trackerPatternPropertiesSection = tv({
  slots: {
    root: "flex flex-col gap-6",
    fields: "flex flex-col gap-4",
    pair: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 *:min-w-0",
    sourceId: "flex min-w-0 items-center gap-2",
    sourceIdText: "min-w-0 break-all font-mono",
  },
});

export const trackerPatternDetectedTrackersSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
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

export const detectedTrackerListItem = tv({
  slots: {
    identifier: "min-w-0 max-w-xs break-all font-mono",
    url: "min-w-0 max-w-xs break-all font-mono",
    date: "whitespace-nowrap",
  },
});

export const cookieBannerDisplaySection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex items-start justify-between gap-4",
    heading: "flex flex-col gap-1",
    empty: "flex flex-col items-center gap-1 py-8 text-center",
    title: "flex min-w-0 items-center gap-2",
    description: "min-w-0 line-clamp-2",
    actions: "flex shrink-0 items-center gap-1",
  },
});

export const cookieBannerThemeSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex items-start justify-between gap-4",
    heading: "flex flex-col gap-1",
    fields: "flex flex-col gap-4",
    colors: "grid grid-cols-3 gap-4 max-sm:grid-cols-1",
    color: "flex flex-col gap-1.5",
    colorRow: "flex items-center gap-2",
    swatch: "h-8 w-10 shrink-0 cursor-pointer rounded-2 border border-sand-6 bg-transparent p-0.5",
    texts: "grid grid-cols-2 gap-4 max-sm:grid-cols-1",
    preview: "flex min-h-70 items-end justify-center bg-sand-3 p-8",
    snippet: "overflow-x-auto font-mono text-2 text-sand-12",
    snippetBar: "flex items-center justify-between",
  },
});

export const categoryDrawer = tv({
  slots: {
    form: "flex min-h-0 flex-1 flex-col",
    fields: "flex flex-col gap-4",
    field: "flex flex-col gap-1",
    checks: "flex flex-col gap-2",
    check: "flex items-center gap-2",
    checkLabel: "font-mono",
    footerActions: "flex flex-row justify-end gap-2",
  },
});

export const categoryCreateDialog = tv({
  slots: {
    form: "flex flex-col",
    fields: "flex flex-col gap-4",
    field: "flex flex-col gap-1",
  },
});
