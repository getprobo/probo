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

export const documentDetailsPage = tv({
  slots: {
    root: [
      "grid grid-cols-1 items-start",
      "transition-[grid-template-columns,column-gap,row-gap] duration-150 ease-out",
      "lg:h-full lg:min-h-0 lg:items-stretch lg:overflow-hidden",
    ],
    main: "flex min-w-0 flex-col gap-4 lg:min-h-0",
    backRow: "flex items-center justify-between gap-3",
    titleRow: "grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-3",
    title: "min-w-0 flex-1",
    // Below lg the cards sit under the document, so the row track collapses.
    // At lg the page grid collapses the column instead.
    sideFrame: [
      "grid min-w-0 overflow-hidden",
      "transition-[grid-template-rows] duration-150 ease-out",
      "lg:h-full lg:min-h-0",
    ],
    sideClip: "min-h-0 min-w-0 overflow-hidden lg:h-full lg:min-h-0",
    side: "relative min-w-0 lg:h-full lg:min-h-0",
    viewport: [
      "w-full overscroll-y-contain outline-none lg:h-full",
      "mask-no-repeat",
      "mask-[linear-gradient(to_bottom,transparent,black_min(16px,var(--scroll-area-overflow-y-start)),black_calc(100%-min(16px,var(--scroll-area-overflow-y-end,16px))),transparent)]",
    ],
    cards: "flex w-full !min-w-0 flex-col gap-4 pr-2 [&>*]:shrink-0",
    scrollbar: [
      "flex w-1 flex-col items-center py-1",
      "pointer-events-none opacity-0 transition-opacity",
      "data-hovering:pointer-events-auto data-hovering:opacity-100",
      "data-scrolling:pointer-events-auto data-scrolling:opacity-100 data-scrolling:duration-0",
    ],
    thumb: "w-1 rounded-full bg-sand-a8",
  },
  variants: {
    cards: {
      true: {
        root: "gap-6 lg:grid-cols-[1fr_calc(280px+0.5rem)]",
        sideFrame: "max-lg:grid-rows-[1fr]",
      },
      false: {
        root: "gap-y-0 lg:grid-cols-[1fr_0px] lg:gap-x-0",
        sideFrame: "max-lg:grid-rows-[0fr]",
      },
    },
  },
  defaultVariants: {
    cards: true,
  },
});

export const documentEditor = tv({
  slots: {
    frame: [
      "@container relative min-h-80 overflow-hidden rounded-4 border border-sand-a3 bg-sand-1",
      "lg:min-h-0 lg:flex-1",
    ],
    viewport: [
      "w-full overscroll-y-contain outline-none lg:h-full",
      "mask-no-repeat",
      "mask-[linear-gradient(to_bottom,transparent,black_min(16px,var(--scroll-area-overflow-y-start)),black_calc(100%-min(16px,var(--scroll-area-overflow-y-end,16px))),transparent)]",
    ],
    // The review column only fits once the frame is wide enough to keep a
    // readable line length beside it.
    content: "grid grid-cols-1 @4xl:grid-cols-[minmax(0,1fr)_18rem]",
    paper: "min-w-0 py-12 pr-8 pl-14",
    review: "hidden py-12 pr-4 @4xl:block",
    scrollbar: [
      "flex w-1 flex-col items-center py-1",
      "pointer-events-none opacity-0 transition-opacity",
      "data-hovering:pointer-events-auto data-hovering:opacity-100",
      "data-scrolling:pointer-events-auto data-scrolling:opacity-100 data-scrolling:duration-0",
    ],
    thumb: "w-1 rounded-full bg-sand-a8",
  },
});

export const documentProperties = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    field: "flex flex-col gap-1",
    value: "px-1.5",
  },
});

export const documentApprovals = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex min-w-0 items-center justify-between gap-2",
    titleLead: "flex min-w-0 items-center gap-2",
    titleIcon: "size-4 shrink-0 text-sand-a9",
    rows: "flex w-full flex-col items-stretch gap-2",
    row: "flex min-w-0 items-center justify-between gap-2",
    name: "min-w-0 flex-1 truncate",
    trailing: "flex shrink-0 items-center gap-1",
  },
});

export const documentSignatories = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex min-w-0 items-center justify-between gap-2",
    titleLead: "flex min-w-0 items-center gap-2",
    titleIcon: "size-4 shrink-0 text-sand-a9",
    rows: "flex w-full flex-col items-stretch gap-2",
    row: "flex min-w-0 items-center justify-between gap-2",
    name: "min-w-0 flex-1 truncate",
    trailing: "flex shrink-0 items-center gap-1",
  },
});

export const documentControls = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex min-w-0 items-center justify-between gap-2",
    names: "flex w-full flex-col items-stretch gap-2",
    name: "flex min-w-0 items-center justify-between gap-2",
    control: "min-w-0 flex-1 truncate !block",
  },
  variants: {
    pending: {
      true: { names: "opacity-60" },
      false: {},
    },
  },
  defaultVariants: {
    pending: false,
  },
});

export const documentReviewList = tv({
  slots: {
    column: "relative",
    card: "absolute inset-x-0 transition-[top] duration-150 ease-out",
  },
  variants: {
    placed: {
      true: {},
      false: { card: "invisible" },
    },
  },
  defaultVariants: {
    placed: false,
  },
});

export const documentReviewItem = tv({
  slots: {
    root: "",
    row: "flex min-w-0 items-start gap-3",
    body: "flex min-w-0 flex-1 flex-col gap-1",
    header: "flex min-w-0 items-baseline gap-2",
    name: "truncate",
    time: "shrink-0 whitespace-nowrap",
  },
  variants: {
    active: {
      true: { root: "border-sand-8! shadow-2" },
      false: {},
    },
  },
  defaultVariants: {
    active: false,
  },
});

export const documentCommentThreadListItem = tv({
  slots: {
    content: "text-2 [&_.tiptap>p:first-child]:mt-0 [&_.tiptap>p:last-child]:mb-0",
  },
});

export const documentSuggestionListItem = tv({
  slots: {
    changes: "flex min-w-0 flex-col gap-1",
    text: "line-clamp-4 block break-words whitespace-pre-line",
    deleted: "line-clamp-4 block break-words whitespace-pre-line line-through",
  },
});

export const documentTitleField = tv({
  slots: {
    root: "min-w-0 flex-1",
    input: [
      "w-full border-0 bg-transparent p-0 text-6 font-medium text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
    heading: "min-w-0",
  },
});
