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
    grid: "grid grid-cols-3 gap-3 max-lg:grid-cols-2 max-sm:grid-cols-1",
    errorCopy: "flex flex-col gap-0.5",
  },
});

export const ndaCard = tv({
  slots: {
    frame: "",
    controls: "flex items-center gap-1",
    actions: "flex flex-wrap items-center gap-2 pt-1",
  },
  variants: {
    dashed: {
      true: {
        frame: "border-dashed",
      },
    },
  },
  defaultVariants: {
    dashed: false,
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
    item: "relative hover:bg-sand-2",
    hit: "absolute inset-0 z-0 outline-none focus-visible:bg-sand-2",
    body: "relative z-1 flex min-w-0 flex-1 items-center gap-4 pointer-events-none",
    avatar: "shrink-0",
    main: "flex min-w-0 flex-1 items-center gap-3",
    identity: "min-w-0",
    name: "min-w-0 truncate",
    email: "min-w-0 truncate",
    trailing: "relative z-1 flex shrink-0 items-center gap-3 pointer-events-none",
    request: "pointer-events-auto",
    nda: "flex shrink-0 items-center gap-1.5 whitespace-nowrap [&_svg]:size-4",
    joined: "whitespace-nowrap",
  },
  variants: {
    inactive: {
      true: {
        item: "opacity-50",
      },
    },
  },
  defaultVariants: {
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

export const visitorPage = tv({
  slots: {
    root: "flex flex-col gap-6 pb-28",
    back: "self-start",
    hero: "grid grid-cols-3 gap-3 max-lg:grid-cols-1",
    profile: "flex h-full min-w-0 flex-col overflow-hidden rounded-4 border border-sand-a3 bg-sand-1",
    person: "flex items-center gap-5 px-5 py-4",
    identity: "flex min-w-0 flex-1 flex-col justify-center gap-1",
    joined: "px-5 py-4",
  },
});

export const visitorPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6 pb-28",
    hero: "grid grid-cols-3 gap-3 max-lg:grid-cols-1",
    profile: "min-w-0",
    nda: "col-span-2 max-lg:col-span-1",
    section: "flex flex-col gap-3",
  },
});

export const electronicSignatureSection = tv({
  slots: {
    root: "col-span-2 h-full min-w-0 max-lg:col-span-1",
    popup: "min-w-lg max-w-lg! p-4!",
    copy: "flex items-center gap-2",
    description: "min-w-0 flex-1",
    trigger: "shrink-0",
    event: "min-w-0",
    timestamp: "shrink-0 whitespace-nowrap tabular-nums",
  },
});

export const documentAccessList = tv({
  slots: {
    root: "flex flex-col gap-3",
    heading: "flex flex-wrap items-center justify-between gap-2",
    titleRow: "flex min-w-0 items-center gap-2",
    title: "min-w-0 truncate",
    meta: "truncate",
    badge: "shrink-0",
    trailing: "flex shrink-0 items-center gap-2",
  },
});

export const documentAccessSelectionBar = tv({
  slots: {
    bar: "fixed inset-x-0 bottom-0 z-2 border-t border-sand-a3 bg-sand-1/80 px-8 py-4 backdrop-blur max-md:px-4",
    inner: "mx-auto flex w-full flex-wrap items-center justify-between gap-4",
    actions: "flex flex-wrap items-center gap-2",
  },
});
