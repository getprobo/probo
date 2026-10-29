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

export const tasksPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    actions: "flex shrink-0 items-center gap-2",
    list: "flex flex-col gap-4",
    userFilter: "w-52 shrink-0 max-sm:w-full",
  },
});

export const taskDetailsPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
    titleRow: "flex flex-wrap items-start justify-between gap-3",
    title: "flex min-w-0 items-center gap-2",
    body: "grid grid-cols-1 items-start gap-6 lg:grid-cols-[1fr_280px]",
    main: "flex min-w-0 flex-col gap-6",
    engagement: "flex flex-col gap-4",
  },
});

export const taskCommentsEmpty = tv({
  slots: {
    root: "flex items-center gap-2 py-2",
    icon: "text-sand-a8 [&_svg]:size-4",
  },
});

export const taskInternalControls = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    names: "flex w-full flex-col items-stretch gap-2",
    name: "flex min-w-0 items-center justify-between gap-2",
    control: "min-w-0 flex-1 truncate !block",
  },
});

export const taskAsanaField = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex items-center gap-2",
    titleIcon: "size-4 shrink-0",
    link: "flex w-full flex-col items-stretch gap-2",
    issue: "hover:underline",
    action: "w-full",
    skeleton: "w-full",
  },
});

export const taskAsanaPublishField = tv({
  slots: {
    root: "flex flex-col",
    action: "w-full",
  },
  variants: {
    spacing: {
      compact: { root: "gap-2" },
      relaxed: { root: "gap-4" },
    },
  },
  defaultVariants: {
    spacing: "compact",
  },
});

export const asanaMappingEditor = tv({
  slots: {
    root: "flex flex-col gap-3",
    row: "flex flex-wrap items-end gap-3",
    control: "min-w-48 flex-1",
    incoming: "flex items-center gap-2 pb-2",
  },
});

export const taskLinearField = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex items-center gap-2",
    titleIcon: "size-4 shrink-0",
    link: "flex w-full flex-col items-stretch gap-2",
    issue: "hover:underline",
  },
});

export const taskLinearPublishField = tv({
  slots: {
    root: "flex flex-col",
  },
  variants: {
    spacing: {
      compact: { root: "gap-2" },
      relaxed: { root: "gap-4" },
    },
  },
  defaultVariants: {
    spacing: "compact",
  },
});

export const tasksSettingsDialog = tv({
  slots: {
    sections: "flex flex-col gap-6",
    section: "flex flex-col gap-3",
    intro: "flex flex-col gap-1",
    list: "flex flex-col gap-3",
    row: "flex items-start justify-between gap-4",
    defaultTeam: "flex flex-col gap-2 border-t border-sand-6 pt-3",
    defaultTeamControl: "block w-full",
    issueImport: "flex flex-col gap-2",
    identity: "flex min-w-0 flex-col gap-1",
    heading: "flex flex-wrap items-center gap-2",
  },
});
