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

export const bindingsPage = tv({
  slots: {
    main: "mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32",
  },
});

export const bindingsEmpty = tv({
  slots: {
    frame: [
      "relative flex min-h-128 flex-col items-center justify-center gap-6",
      "overflow-hidden rounded-5 border border-sand-3 bg-sand-1 px-8 py-12",
    ],
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 z-0 h-full",
      "bg-[radial-gradient(ellipse_70%_52%_at_50%_-8%,rgb(230_255_3/0.72)_0%,rgb(230_255_3/0.28)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    content: "relative z-1 flex w-full flex-col items-center gap-6",
    copy: "flex w-full flex-col items-center gap-4",
    icon: "size-8 shrink-0 [&_svg]:size-8",
    description: "text-sand-a11",
  },
});

export const bindingsList = tv({
  slots: {
    body: "flex flex-col gap-3",
    frame: "overflow-hidden rounded-5 border border-sand-3 bg-sand-1",
    table: "rounded-none",
    header: "sr-only",
  },
});

export const bindingListItem = tv({
  slots: {
    accountCell: "h-12 p-0",
    row: "flex h-full items-center gap-8 px-8",
    title: "min-w-0 flex-1 truncate",
    meta: "min-w-0 truncate text-sand-a11",
    actionCell: "w-32 p-0 pr-8",
  },
});

export const bindConfirmCard = tv({
  slots: {
    frame: "flex flex-col items-center gap-8 p-16",
    header: "flex w-full flex-col items-center gap-4 text-center",
    icon: "size-8 shrink-0 [&_svg]:size-8",
    body: "flex w-full flex-col gap-8",
    fields: "grid w-full grid-cols-1 overflow-hidden rounded-5 border border-sand-4 sm:grid-cols-2",
    field: [
      "flex flex-col gap-2 border-sand-4 p-8",
      "max-sm:not-last:border-b",
      "sm:odd:border-r",
    ],
    value: "break-all",
    action: "w-full",
  },
});
