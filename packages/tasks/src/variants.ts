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

export const taskList = tv({
  slots: {
    root: "flex flex-col gap-4",
    group: "flex flex-col gap-2",
    sectionHeader: [
      "appearance-none cursor-pointer border-0 bg-transparent p-0 text-inherit",
      "flex w-full items-center gap-2 text-start outline-none",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
      "[&_[data-slot=caret-open]]:hidden",
      "data-panel-open:[&_[data-slot=caret-closed]]:hidden",
      "data-panel-open:[&_[data-slot=caret-open]]:block",
    ],
    caret: "size-4 shrink-0 text-sand-11",
    list: "divide-y divide-sand-a3",
    empty: "py-6",
  },
  variants: {
    offset: {
      true: { sectionHeader: "pt-4" },
      false: {},
    },
    spaced: {
      true: { sectionHeader: "pb-2" },
      false: {},
    },
  },
  defaultVariants: {
    offset: false,
    spaced: false,
  },
});

export const taskListRow = tv({
  base: "bg-sand-1",
  variants: {
    position: {
      only: "rounded-4 border border-sand-a3",
      first: "rounded-t-4 border-x border-t border-sand-a3",
      middle: "border-x border-t border-sand-a3",
      last: "rounded-b-4 border border-sand-a3",
    },
    placeholder: {
      true: "h-12 animate-pulse",
      false: "",
    },
  },
  defaultVariants: {
    position: "only",
    placeholder: false,
  },
});

export const taskListItem = tv({
  slots: {
    root: "relative flex items-center gap-3 px-6 py-3",
    main: "flex min-w-0 flex-1 items-center gap-3",
    title: "min-w-0 flex-1 truncate",
    titleLink: "hover:underline after:absolute after:inset-0 after:content-['']",
    metaGroup: "relative z-1 ml-auto flex h-6 shrink-0 items-center gap-3",
    meta: "flex h-6 shrink-0 items-center gap-1 text-sand-11 [&_svg]:size-3.5",
    metaIcon: "text-indigo-11",
    deadline: "flex h-6 shrink-0 items-center gap-1 text-sand-12 [&_svg]:size-3.5",
    deadlineIcon: "text-red-11",
    deadlineCustom: "p-1",
    chipTrigger: "appearance-none cursor-pointer border-0 bg-transparent p-0 text-inherit",
    created: "relative z-1 flex h-6 shrink-0 items-center text-sand-11",
    externalLink: "relative z-1 flex h-6 shrink-0 items-center",
    trailing: "relative z-1 flex h-6 shrink-0 items-center gap-2",
    assignee: "relative z-1 flex h-6 shrink-0 items-center",
    state: "relative z-1 flex h-6 shrink-0 items-center",
    iconTrigger: [
      "size-4! h-4! w-4! min-h-0! min-w-0! justify-center gap-0 p-0! px-0!",
      "hover:bg-transparent",
      "[&>:first-child]:min-w-0 [&>:first-child]:flex-none",
      "[&>:last-child]:hidden",
    ],
    stateTrigger: [
      "h-6! w-auto min-w-0 items-center gap-0 px-0!",
      "[&>:first-child]:flex [&>:first-child]:h-full [&>:first-child]:min-w-0 [&>:first-child]:items-center",
      "[&>:last-child]:hidden",
    ],
    stateOption: "flex items-center gap-2",
  },
  variants: {
    linked: {
      true: { root: "hover:bg-sand-2" },
      false: {},
    },
    interaction: {
      idle: { root: "" },
      grab: { root: "cursor-grab select-none" },
      grabbing: { root: "cursor-grabbing select-none" },
      dragging: { root: "cursor-grabbing opacity-40 select-none" },
      ghost: { root: "bg-sand-3 opacity-50 select-none" },
    },
  },
  defaultVariants: {
    linked: true,
    interaction: "idle",
  },
});

export const taskDetails = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
    titleRow: "flex flex-wrap items-start justify-between gap-3",
    title: "flex min-w-0 flex-1 items-center gap-2",
    actions: "flex shrink-0 flex-wrap items-center gap-2",
    body: "grid grid-cols-1 items-start gap-6 lg:grid-cols-[1fr_360px]",
    main: "flex min-w-0 flex-col gap-6",
  },
});

export const taskNameField = tv({
  slots: {
    root: "min-w-0 flex-1",
    input: [
      "w-full border-0 bg-transparent p-0 text-6 font-medium text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
  },
});

export const taskProperties = tv({
  slots: {
    root: "flex flex-col",
    row: "grid grid-cols-[7.5rem_minmax(0,1fr)] gap-3 border-b border-sand-6 py-2.5 last:border-b-0",
    label: "",
    value: "flex min-w-0 items-center gap-2",
  },
  variants: {
    align: {
      center: { row: "items-center" },
      start: { row: "items-start", label: "pt-1" },
    },
  },
  defaultVariants: {
    align: "center",
  },
});

export const taskEngagement = tv({
  slots: {
    root: "flex min-w-0 flex-col gap-4",
    panel: "min-w-0",
  },
});

export const taskCommentsEmpty = tv({
  slots: {
    root: "flex items-center gap-2 py-2",
    icon: "text-sand-a8 [&_svg]:size-4",
  },
});

export const taskDurationField = tv({
  slots: {
    root: "grid grid-cols-[4.5rem_minmax(0,1fr)_auto] items-center gap-2",
  },
});

export const taskCommentsSection = tv({
  slots: {
    root: "flex min-w-0 flex-col gap-3",
    actions: "flex justify-end",
    list: "flex flex-col gap-4",
    itemSkeleton: "h-24 animate-pulse rounded-2 bg-sand-3",
  },
});

export const taskCommentForm = tv({
  slots: {
    root: "flex flex-col gap-3",
    actions: "flex justify-end",
  },
});

export const taskCommentEditor = tv({
  base: [
    "flex-none! overflow-visible! bg-sand-1! py-3! pr-3! pl-14! shadow-2!",
    "[&_.tiptap>p:first-child]:mt-0",
  ],
});

export const taskCommentListItem = tv({
  slots: {
    root: "flex flex-col gap-2",
    header: "flex items-center gap-2",
    meta: "flex min-w-0 flex-1 flex-wrap items-baseline gap-x-2 gap-y-0.5",
    content: "min-w-0",
  },
});

export const taskDescription = tv({
  slots: {
    root: "min-w-0",
    editor: "min-h-40",
  },
});
