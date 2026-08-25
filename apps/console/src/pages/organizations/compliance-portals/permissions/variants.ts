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

export const permissionsPage = tv({
  base: "flex flex-col gap-6",
});

export const permissionsPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    section: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
  },
});

export const ndaSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    body: "flex flex-col gap-3",
    empty: "flex flex-col items-center justify-center gap-3 rounded-4 border border-dashed border-sand-6 px-4 py-8",
    emptyCopy: "max-w-md text-center",
    errorCopy: "flex flex-col gap-0.5",
  },
});

export const ndaCard = tv({
  slots: {
    root: "flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between",
    copy: "flex min-w-0 flex-col gap-1",
    actions: "flex shrink-0 flex-wrap items-center gap-2",
  },
});

export const accessSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    more: "flex justify-center",
  },
});

export const accessListItem = tv({
  slots: {
    item: "",
    trailing: "relative z-2 flex shrink-0 items-center gap-3",
    counts: "flex flex-col items-end gap-0.5",
  },
  variants: {
    interactive: {
      true: {
        item: "cursor-pointer hover:bg-sand-2",
      },
    },
    inactive: {
      true: {
        item: "opacity-50",
      },
    },
  },
  defaultVariants: {
    interactive: false,
    inactive: false,
  },
});

export const accessListEmpty = tv({
  slots: {
    root: "flex flex-col items-center gap-5 py-8 text-center",
    icon: "text-sand-a8 [&_svg]:size-6",
    body: "flex max-w-xs flex-col items-center gap-2",
  },
});
