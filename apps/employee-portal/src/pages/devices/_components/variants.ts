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
    body: "grid grid-cols-1 gap-4 md:grid-cols-4",
    stepper: "flex list-none flex-col gap-2 md:col-span-1",
    stage: "min-w-0 md:col-span-3",
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
    copy: "flex min-w-0 flex-1 flex-col gap-1",
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
    frame: "flex flex-col items-center gap-8",
    header: "flex w-full flex-col items-center gap-4 text-center",
    icon: "size-8 shrink-0 text-sand-12 [&_svg]:size-8",
    body: "flex w-full flex-col items-center gap-6",
  },
});

export const reviewGrid = tv({
  slots: {
    root: "grid w-full grid-cols-1 overflow-hidden rounded-5 border border-sand-3 sm:grid-cols-2",
    cell: [
      "flex flex-col gap-8 border-sand-3 p-8",
      "max-sm:not-last:border-b",
      "sm:odd:border-r sm:[&:nth-child(-n+2)]:border-b",
    ],
    heading: "flex items-center gap-3",
    headingIcon: "size-4 shrink-0 text-sand-12 [&_svg]:size-4",
    items: "flex flex-col gap-3",
    item: "flex items-center gap-3",
    itemIcon: "size-4 shrink-0 text-sand-12",
  },
});

export const downloadList = tv({
  slots: {
    item: "h-12 gap-8 px-8",
    meta: "w-24 shrink-0",
  },
});
