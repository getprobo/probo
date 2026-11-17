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
      "grid grid-cols-1 items-start gap-6",
      "lg:h-full lg:min-h-0 lg:items-stretch lg:overflow-hidden lg:grid-cols-[1fr_calc(280px+0.5rem)]",
    ],
    main: "flex min-w-0 flex-col gap-6",
    side: "relative min-w-0 lg:min-h-0",
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
