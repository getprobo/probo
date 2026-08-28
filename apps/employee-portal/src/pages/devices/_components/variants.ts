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

export const registerDevicePage = tv({
  slots: {
    main: "mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32",
    body: "grid grid-cols-1 gap-4 md:grid-cols-[15.25rem_minmax(0,1fr)]",
    stepper: "flex w-full list-none flex-col gap-2",
    stage: "min-w-0",
  },
});

export const progressStep = tv({
  slots: {
    root: [
      "flex w-full items-start gap-3 rounded-5 p-3 text-left",
      "outline-none focus-visible:ring-2 focus-visible:ring-sand-8",
    ],
    badge: "flex size-5 shrink-0 items-center justify-center rounded-full text-1 font-bold",
    icon: "size-3.5",
    copy: "flex min-w-0 flex-1 flex-col justify-center gap-1",
    description: "text-sand-a9",
  },
  variants: {
    state: {
      complete: {
        root: "cursor-pointer border-0 bg-transparent",
        badge: "bg-sand-12 text-sand-2",
      },
      current: {
        root: "bg-sand-a3",
        badge: "bg-sand-12 text-sand-2",
      },
      upcoming: {
        badge: "bg-sand-a3 text-sand-12",
      },
    },
  },
});

export const registerDeviceCard = tv({
  slots: {
    frame: "flex flex-col items-center gap-8 p-16",
    header: "flex w-full flex-col items-center gap-4 text-center",
    icon: "size-8 shrink-0 text-sand-12 [&_svg]:size-8",
    body: "flex w-full flex-col items-center gap-8",
  },
});

export const reviewGrid = tv({
  slots: {
    root: "grid w-full grid-cols-1 overflow-hidden rounded-5 border border-sand-4 sm:grid-cols-2",
    cell: [
      "flex flex-col gap-8 border-sand-4 p-8",
      "max-sm:not-last:border-b",
      "sm:odd:border-r sm:[&:nth-child(-n+2)]:border-b",
    ],
    heading: "flex items-center gap-3",
    headingIcon: "size-4 shrink-0 text-sand-12 [&_svg]:size-4",
    items: "flex flex-col gap-2",
    item: "flex items-center gap-3",
    itemIcon: "size-4 shrink-0 text-sand-11",
  },
});

export const downloadList = tv({
  slots: {
    item: "h-12 gap-8 px-8",
    meta: "w-24 shrink-0",
  },
});

export const devicesPage = tv({
  slots: {
    main: "mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32",
  },
});

export const devicesEmpty = tv({
  slots: {
    frame: [
      "relative flex min-h-128 flex-col items-center justify-center gap-6",
      "overflow-hidden rounded-5 border border-sand-3 bg-sand-1 px-8 py-12",
    ],
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 z-0 h-full",
      "bg-[radial-gradient(ellipse_70%_52%_at_50%_-8%,rgb(230_255_3_/_0.72)_0%,rgb(230_255_3_/_0.28)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    content: "relative z-1 flex w-full flex-col items-center gap-6",
    copy: "flex w-full flex-col items-center gap-4",
    icon: "size-8 shrink-0 text-sand-a9 [&_svg]:size-8",
    description: "text-sand-a11",
    actions: "flex items-center gap-3",
  },
});

export const devicesList = tv({
  slots: {
    body: "flex flex-col gap-3",
    frame: "overflow-hidden rounded-5 border border-sand-3 bg-sand-1 transition-opacity duration-150",
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

export const deviceListItem = tv({
  slots: {
    // Size 2 cells are h-11 px-3 py-3; Figma is a 48px row with px-8, a 32px
    // gap between the title and the metadata cluster, and gap-4 inside it.
    cell: "h-12 p-0",
    row: "flex h-full items-center gap-8 px-8",
    title: "min-w-0 flex-1 truncate",
    meta: "flex shrink-0 items-center gap-4",
    timestamp: "flex items-center gap-[3px]",
    timestampLabel: "text-sand-a8",
    timestampValue: "w-32 text-sand-a11",
    os: "w-24 text-sand-a11",
    status: "flex w-32 items-center gap-1",
    pipWrap: "flex size-4 shrink-0 items-center justify-center",
    pip: "size-1.5 rounded-full",
    statusLabel: "text-sand-a11",
  },
  variants: {
    connected: {
      true: {
        pip: "bg-green-8 ring-4 ring-green-3",
      },
      false: {
        pip: "bg-sand-8 ring-4 ring-sand-3",
      },
    },
  },
});

export const addManuallyPage = tv({
  slots: {
    main: "mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32",
    creating: "flex items-center gap-2",
    errorActions: "flex items-center gap-3",
  },
});

export const enrollmentInstructions = tv({
  slots: {
    root: "flex flex-col gap-10",
    token: "flex flex-col gap-3",
    install: "flex flex-col gap-4",
    group: "flex flex-col gap-3",
  },
});

export const copyableCodeBlock = tv({
  slots: {
    root: "overflow-hidden",
    toolbar: "flex items-center justify-end border-b border-sand-a3 px-2 py-1",
    pre: "overflow-x-auto whitespace-pre bg-sand-3 p-4 font-mono text-2 text-sand-12",
  },
});
