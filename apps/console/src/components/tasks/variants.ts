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

export const tasksCard = tv({
  slots: {
    root: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
    list: "divide-y divide-sand-6",
    sectionHeader: "flex items-center gap-2 bg-sand-3 px-6 py-3",
    stateOption: "flex items-center gap-2",
    empty: "py-6",
  },
  variants: {
    dragging: {
      true: {
        sectionHeader: "border-2 border-dashed border-transparent hover:border-sand-8",
      },
    },
  },
  defaultVariants: {
    dragging: false,
  },
});

export const taskListItem = tv({
  slots: {
    root: "relative flex items-center gap-3 px-6 py-3 hover:bg-sand-2",
    main: "flex min-w-0 flex-1 items-center gap-3",
    title: "min-w-0 truncate",
    // The overlay stretches the title link across the whole row; siblings that
    // must stay clickable sit above it with `relative z-1`.
    titleLink: "hover:underline after:absolute after:inset-0 after:content-['']",
    meta: "flex shrink-0 items-center gap-1 text-sand-11 [&_svg]:size-3.5",
    externalLink: "relative z-1 shrink-0",
    assignee: "relative z-1 ml-auto shrink-0",
    state: "relative z-1 w-40 shrink-0",
    stateOption: "flex items-center gap-2",
  },
  variants: {
    // One variant rather than several booleans: tailwind-variants/lite has no
    // merge, so overlapping cursor/opacity classes would both be emitted.
    interaction: {
      idle: { root: "" },
      grab: { root: "cursor-grab select-none" },
      grabbing: { root: "cursor-grabbing select-none" },
      dragging: { root: "cursor-grabbing opacity-40 select-none" },
      ghost: { root: "bg-sand-3 opacity-50 select-none" },
    },
  },
  defaultVariants: {
    interaction: "idle",
  },
});
