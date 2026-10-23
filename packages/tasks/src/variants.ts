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
    section: "",
    sectionHeader: [
      "appearance-none cursor-pointer border-0 bg-gold-3 text-gold-12",
      "flex w-full items-center gap-2 px-3 py-1.5 text-start outline-none",
      "focus-visible:ring-2 focus-visible:ring-gold-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-2",
    ],
    caret: "size-4 shrink-0 text-gold-11",
    empty: "py-6",
  },
  variants: {
    offset: {
      true: { section: "pt-2" },
      false: {},
    },
    pinned: {
      true: { sectionHeader: "rounded-t-none rounded-b-4 shadow-2" },
      false: { sectionHeader: "rounded-4" },
    },
  },
  defaultVariants: {
    offset: false,
    pinned: false,
  },
});

export const taskListRow = tv({
  base: "box-border flex h-12 overflow-hidden bg-sand-1",
  variants: {
    position: {
      only: "rounded-4 border border-sand-a3",
      first: "rounded-t-4 border-x border-t border-sand-a3",
      middle: "border-x border-t border-sand-a3",
      last: "rounded-b-4 border border-sand-a3",
    },
  },
  defaultVariants: {
    position: "only",
  },
});

export const taskListItem = tv({
  slots: {
    root: "relative flex min-h-0 min-w-0 flex-1 items-center gap-3 px-6",
    main: "flex min-w-0 flex-1 items-center gap-3",
    title: "flex min-w-0 flex-1 items-center gap-1.5",
    titleText: "min-w-0 truncate",
    titleLink: "hover:underline after:absolute after:inset-0 after:content-['']",
    metaGroup: "relative z-1 ml-auto flex h-6 shrink-0 items-center gap-3",
    meta: "flex h-6 shrink-0 items-center gap-1 text-sand-11 [&_svg]:size-3.5",
    metaIcon: "text-indigo-11",
    deadline: "flex h-6 shrink-0 items-center gap-1 text-sand-12 [&_svg]:size-3.5",
    deadlineIcon: "text-red-11",
    deadlineCustom: "p-1",
    chipTrigger: "appearance-none cursor-pointer border-0 bg-transparent p-0 text-inherit",
    created: "relative z-1 grid h-6 shrink-0 items-center justify-items-end whitespace-nowrap text-1 text-sand-11 tabular-nums",
    more: "relative z-1 flex h-6 shrink-0 items-center",
    trailing: "relative z-1 flex h-6 shrink-0 items-center gap-2",
    assignee: "relative z-1 flex h-6 shrink-0 items-center",
    assigneeLabel: "min-w-0 truncate",
    assigneeIcon: "size-4 shrink-0 text-sand-8",
    assigneeAvatar: "!size-4",
    state: "relative z-1 flex h-6 shrink-0 items-center",
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
    field: {
      true: {
        chipTrigger: "rounded-2 !px-1.5 !py-0.5 hover:bg-sand-3",
        meta: "!h-auto !gap-2 [&_svg]:!size-4",
        deadline: "!h-auto !gap-2 !text-sand-11 [&_svg]:!size-4",
      },
      false: {},
    },
  },
  defaultVariants: {
    linked: true,
    interaction: "idle",
    field: false,
  },
});

export const taskDetails = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
    titleRow: "flex flex-wrap items-start justify-between gap-3",
    title: "flex min-w-0 flex-1 items-center gap-2",
    actions: "flex shrink-0 flex-wrap items-center gap-2",
    body: "grid grid-cols-1 items-start gap-6 lg:grid-cols-[1fr_280px]",
    main: "flex min-w-0 flex-col gap-6",
    editorCard: "border-0 bg-sand-1 !py-2 shadow-3",
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
    root: "flex flex-col items-start gap-3",
    recurrence: "inline-flex min-w-0 items-center gap-2",
    recurrenceIcon: "size-4 shrink-0 text-indigo-11",
    recurrenceField: "min-w-0 flex-1",
  },
});

export const taskEngagement = tv({
  slots: {
    root: "flex min-w-0 flex-col gap-4",
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
    list: "flex flex-col gap-4",
    inset: "pl-3",
    actions: "flex justify-end",
    itemSkeleton: "h-24 animate-pulse rounded-2 bg-sand-3",
  },
});

export const taskCommentForm = tv({
  slots: {
    root: "flex flex-col gap-3",
    actions: "flex justify-end",
    editor: [
      "min-h-8 flex-none! overflow-visible! bg-transparent! py-0! pr-0! shadow-none!",
      "[&_.tiptap>p:first-child]:mt-0",
    ],
  },
});

export const taskCommentEditor = tv({
  base: [
    "flex-none! overflow-visible! bg-transparent! py-0! pr-0! shadow-none!",
    "[&_.tiptap>p:first-child]:mt-0",
  ],
  variants: {
    editing: {
      true: "min-h-40",
      false: "",
    },
  },
  defaultVariants: {
    editing: false,
  },
});

export const taskActivityComment = tv({
  slots: {
    item: "before:!hidden",
    card: "group -ml-3 flex w-[calc(100%+0.75rem)] min-w-0 flex-1 flex-col",
    row: "flex min-w-0 items-start gap-3",
    body: "flex min-w-0 flex-1 flex-col gap-2",
    header: "flex min-w-0 items-center gap-2",
    identity: "flex min-w-0 flex-1 items-baseline gap-2",
    actions: [
      "flex shrink-0 items-center overflow-hidden",
      "w-0 opacity-0",
      "group-hover:w-auto group-hover:opacity-100",
      "group-focus-within:w-auto group-focus-within:opacity-100",
    ],
    save: "flex justify-end gap-2",
  },
});

export const taskDescription = tv({
  slots: {
    root: "min-w-0",
    editor: "min-h-40",
  },
});

export const taskListView = tv({
  base: "transition-opacity duration-150",
  variants: {
    busy: {
      true: "opacity-60",
      false: "",
    },
  },
  defaultVariants: {
    busy: false,
  },
});

export const taskListToolbar = tv({
  slots: {
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
    dueFilter: "w-48 shrink-0",
    stateOption: "flex items-center gap-2",
  },
});
