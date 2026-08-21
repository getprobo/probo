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

export const hostingPage = tv({
  base: "flex flex-col gap-8",
});

export const hostingPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-8",
    section: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
  },
});

export const statusSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    headingRow: "flex items-center justify-between",
    grid: "grid grid-cols-3 gap-4 max-lg:grid-cols-2 max-sm:grid-cols-1",
  },
});

export const hostingCard = tv({
  slots: {
    // Ghost Card has no chrome; the frame owns border + fill so tone can
    // swap hue without fighting Card's sand-a3 (tv/lite has no merge).
    frame: "flex h-full flex-col overflow-hidden rounded-5 border bg-sand-1",
    header: "relative flex w-full items-center justify-between gap-3 overflow-hidden p-8 max-md:p-4",
    wash: "pointer-events-none absolute top-1/2 left-1/2 aspect-square w-full -translate-x-1/2 -translate-y-1/2 opacity-10 blur-[14px]",
    fade: "pointer-events-none absolute inset-0 bg-linear-to-b from-sand-1/0 to-sand-1",
    icon: "relative z-1 flex size-12 items-center justify-center overflow-hidden rounded-3",
    control: "relative z-1 shrink-0",
    body: "flex flex-1 flex-col gap-2 px-8 pb-8 max-md:px-4 max-md:pb-4",
  },
  variants: {
    tone: {
      sand: {
        frame: "border-sand-a3",
        wash: "hidden",
        fade: "hidden",
        icon: "bg-sand-1 text-sand-9",
      },
      green: {
        frame: "border-green-6",
        wash: "bg-green-9",
        icon: "bg-transparent text-green-11",
      },
      sky: {
        frame: "border-sky-6",
        wash: "bg-sky-9",
        icon: "bg-transparent text-sky-11",
      },
      amber: {
        frame: "border-amber-6",
        wash: "bg-amber-9",
        icon: "bg-transparent text-amber-11",
      },
      red: {
        frame: "border-red-6",
        wash: "bg-red-9",
        icon: "bg-transparent text-red-11",
      },
    },
    wide: {
      true: {
        frame: "col-span-3 max-lg:col-span-2 max-sm:col-span-1",
      },
    },
  },
  defaultVariants: {
    tone: "sand",
    wide: false,
  },
});

export const domainsSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    grid: "grid grid-cols-3 gap-4 max-lg:grid-cols-2 max-sm:grid-cols-1",
  },
});

export const domainCard = tv({
  slots: {
    lead: "relative z-1 flex min-w-0 flex-1 items-center gap-3",
    subtitle: "min-w-0 truncate",
    copy: "flex flex-col gap-2",
    heading: "flex items-center justify-between gap-3",
    identity: "flex min-w-0 flex-wrap items-center gap-2",
    dns: "flex flex-col gap-3",
    record: "flex flex-col gap-2 rounded-4 bg-sand-3 p-4",
    recordHeader: "flex items-center justify-between gap-2",
    recordField: "flex flex-col gap-1",
    recordValue: "flex min-w-0 items-center gap-2",
    code: "min-w-0 flex-1 break-all",
  },
});

export const customDomainForm = tv({
  slots: {
    form: "flex flex-1 flex-col",
    field: "w-full max-w-lg",
    actions: "self-start",
  },
});

export const domainFormDialog = tv({
  slots: {
    form: "flex flex-col gap-4",
    fields: "flex flex-col gap-4",
    examples: "rounded-4 bg-sand-3 p-4",
  },
});
