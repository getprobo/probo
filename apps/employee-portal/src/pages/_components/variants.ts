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

export const getStartedCard = tv({
  slots: {
    frame: "relative flex h-full w-full flex-col overflow-hidden",
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 h-3/5",
      "bg-[radial-gradient(ellipse_85%_55%_at_50%_0%,rgb(230_255_3_/_0.72)_0%,rgb(230_255_3_/_0.28)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    header: "relative z-1 flex flex-col items-center gap-4 px-8 py-12 text-center",
    icon: "relative size-8 shrink-0 text-sand-a9",
    copy: "relative flex min-w-0 flex-col items-center gap-4",
    description: "max-w-64",
    steps: "relative z-1 mt-auto flex flex-col",
  },
});

export const getStartedStep = tv({
  slots: {
    row: "flex items-center gap-4 border-t border-sand-a2 px-8 py-6",
    lead: "flex min-w-0 flex-1 items-center gap-4",
    index: [
      "flex size-5 shrink-0 items-center justify-center rounded-full",
      "text-1 font-bold",
    ],
    copy: "flex min-w-0 flex-col gap-1",
    description: "text-sand-a9",
  },
  variants: {
    tone: {
      current: {
        index: "bg-sand-12 text-sand-2",
      },
      upcoming: {
        index: "bg-sand-a3 text-sand-12",
      },
    },
  },
  defaultVariants: {
    tone: "current",
  },
});

export const dashboardCard = tv({
  slots: {
    frame: "relative flex h-[300px] flex-col gap-8 overflow-hidden p-8",
    wash: [
      "pointer-events-none absolute inset-x-0 bottom-0 z-0 h-3/5",
      "bg-[radial-gradient(ellipse_85%_55%_at_50%_100%,rgb(230_255_3_/_0.72)_0%,rgb(230_255_3_/_0.28)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_top,black_0%,black_40%,transparent_100%)]",
    ],
    header: "relative z-1 flex items-start gap-4",
    icon: "size-8 shrink-0 text-sand-a9",
    copy: "flex min-w-0 flex-1 flex-col justify-center gap-1",
    description: "text-sand-a9",
    view: "shrink-0 font-medium",
    body: "relative z-1 flex flex-1 flex-col items-center justify-center gap-6",
    empty: "flex flex-col items-center gap-6",
    emptyIcon: "size-8 text-sand-a9",
    emptyLabel: "text-sand-9",
  },
  variants: {
    wash: {
      true: {},
      false: {
        wash: "hidden",
      },
    },
  },
  defaultVariants: {
    wash: false,
  },
});

export const pageHeader = tv({
  slots: {
    root: "flex flex-col gap-4",
    crumbs: "flex items-center gap-3",
    chevron: "size-3 shrink-0 text-sand-11",
  },
});

export const documentQueueSummary = tv({
  slots: {
    frame: "relative flex flex-col items-center justify-center gap-6 overflow-hidden border-b border-sand-a2 px-8 py-12",
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 z-0 h-full",
      "bg-[radial-gradient(ellipse_70%_52%_at_50%_-8%,rgb(230_255_3_/_0.72)_0%,rgb(230_255_3_/_0.28)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    content: "relative z-1 flex w-full flex-col items-center gap-6",
    copy: "flex w-full flex-col items-center gap-4",
    icon: "size-8 shrink-0 text-sand-a9 [&_svg]:size-8",
  },
});

export const employeeDocumentListItem = tv({
  slots: {
    row: "",
    // Size 2 cells are h-11 px-3; Figma rows are h-12 with px-8 on the ends.
    titleCell: "h-12 w-full max-w-0 py-0 pl-8",
    title: "relative z-1 block min-w-0 truncate",
    titleLink: [
      "block min-w-0 truncate",
      "after:absolute after:inset-0 after:-z-1",
      "outline-none focus-visible:after:bg-sand-2",
    ],
    metaCell: "relative z-1 h-12 whitespace-nowrap py-0",
    timestamp: "flex items-center gap-[3px]",
    timestampLabel: "text-sand-a8",
    timestampValue: "text-sand-a11",
    chip: "text-sand-a11",
    trailingCell: "relative z-1 h-12 whitespace-nowrap py-0 pr-8",
    chevron: "size-4 text-sand-11",
  },
  variants: {
    trailing: {
      action: {},
      chevron: {
        row: "relative isolate cursor-pointer",
      },
    },
  },
});

export const documentEmpty = tv({
  slots: {
    frame: "flex h-[200px] flex-col items-center justify-center gap-6",
    icon: "size-8 text-sand-a9 [&_svg]:size-8",
    copy: "flex flex-col items-center gap-2 text-center",
    label: "text-sand-9",
  },
});

export const documentListSection = tv({
  slots: {
    root: "flex flex-col gap-9",
    heading: "flex items-center gap-1",
    body: "flex flex-col gap-3",
    frame: "overflow-hidden rounded-5 border border-sand-a3 bg-sand-1 transition-opacity duration-150",
    table: "rounded-none",
    header: "sr-only",
    pager: "flex justify-center",
  },
  variants: {
    busy: {
      true: {
        frame: "opacity-60",
      },
      false: {},
    },
  },
  defaultVariants: {
    busy: false,
  },
});

export const queueTopBar = tv({
  slots: {
    // Inverse theme island: .dark / .light + color-scheme are set on the
    // header so stock outline buttons resolve the opposite scale.
    bar: "flex h-14 shrink-0 items-center justify-between bg-sand-1 px-8",
    start: "flex min-w-0 items-center gap-6",
    controls: "flex items-center gap-2",
    progress: "truncate",
  },
});

export const documentWorkspace = tv({
  slots: {
    root: [
      "flex h-full min-h-0 flex-1 overflow-hidden",
      "max-md:flex-col",
      "[view-transition-name:document-workspace]",
    ],
    request: [
      "flex w-full shrink-0 flex-col overflow-hidden gap-8 bg-sand-1 px-8 py-8",
      "min-h-0 max-md:max-h-[50%] md:h-full md:w-96 xl:w-128",
    ],
    history: "flex min-h-0 flex-1 flex-col gap-8",
    stage: "flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden bg-sand-3",
    loading: "grid min-h-0 flex-1 place-items-center",
    spinner: "size-6 animate-spin text-sand-a10",
  },
});

export const pdfPreview = tv({
  slots: {
    viewport: "min-h-0 flex-1 overflow-auto bg-sand-3",
    list: "flex flex-col items-center gap-4 px-8 py-8",
    loading: "grid place-items-center py-16 text-sand-a10",
    spinner: "size-6 animate-spin",
    page: "shadow-3",
  },
});

export const documentViewerToolbar = tv({
  slots: {
    // Container query: the stage is often narrower than the viewport (request
    // panel beside it), so viewport sm is too late to drop action labels.
    root: [
      "@container flex h-12 shrink-0 items-center justify-between gap-x-4",
      "bg-sand-1 px-6",
    ],
    start: "flex min-w-0 items-center gap-2",
    controls: "flex items-center gap-1",
    actions: "flex shrink-0 items-center gap-2",
    actionLabeled: "@max-xl:hidden",
    actionIcon: "hidden @max-xl:inline-flex",
    separator: "h-6",
  },
});

export const documentRequestPanel = tv({
  slots: {
    root: "flex shrink-0 flex-col gap-16",
    copy: "flex flex-col gap-4",
    eyebrow: "text-sand-9",
    status: "flex items-center gap-3",
    statusIcon: "size-4 shrink-0",
    actions: "flex w-full flex-col gap-6",
    actionRow: "flex w-full flex-col gap-3 sm:flex-row",
    consent: "text-sand-a9",
  },
  variants: {
    tone: {
      signed: {
        status: "text-indigo-11",
        statusIcon: "text-indigo-11",
      },
      approved: {
        status: "text-green-11",
        statusIcon: "text-green-11",
      },
      rejected: {
        status: "text-red-11",
        statusIcon: "text-red-11",
      },
      voided: {
        status: "text-sand-11",
        statusIcon: "text-sand-11",
      },
    },
  },
});

export const documentVersionHistory = tv({
  slots: {
    root: "flex min-h-0 w-full flex-1 flex-col gap-2",
    header: "flex shrink-0 items-center justify-between",
    frame: "relative min-h-0 flex-1 overflow-clip border-x border-sand-3 bg-sand-1",
    viewport: [
      "h-full overscroll-y-contain snap-y snap-mandatory outline-none",
      "mask-no-repeat",
      "mask-[linear-gradient(to_bottom,transparent,black_min(16px,var(--scroll-area-overflow-y-start)),black_calc(100%-min(16px,var(--scroll-area-overflow-y-end,16px))),transparent)]",
    ],
    list: "flex flex-col",
    scrollbar: [
      "flex w-4 flex-col items-center py-1",
      "opacity-0 transition-opacity pointer-events-none",
      "data-hovering:pointer-events-auto data-hovering:opacity-100",
      "data-scrolling:pointer-events-auto data-scrolling:opacity-100 data-scrolling:duration-0",
    ],
    thumb: "w-1 rounded-full bg-sand-a8",
    ghost: "h-[72px] shrink-0 opacity-40",
  },
});

export const documentVersionHistoryItem = tv({
  slots: {
    row: [
      "flex h-[72px] w-full shrink-0 snap-start items-center gap-4",
      "border-b border-sand-3 px-8 py-4 text-left last:border-b-0",
      "hover:bg-sand-3",
    ],
    radio: "flex size-5 shrink-0 items-center justify-center rounded-full bg-sand-a5",
    radioDot: "size-2.5 rounded-full bg-sand-12",
    copy: "flex min-w-0 flex-1 flex-col gap-1",
    meta: "flex min-w-0 items-center gap-1",
    current: "shrink-0",
  },
});
