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

export const scimPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    section: "flex flex-col gap-4",
    grid: "grid grid-cols-1 gap-3 md:grid-cols-3",
    exportFields: "flex flex-col gap-3",
  },
});

export const scimSetupCard = tv({
  slots: {
    frame: "relative flex h-full flex-col gap-6 overflow-hidden px-6 pt-10 pb-6",
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 z-0 h-3/5",
      "bg-[radial-gradient(ellipse_85%_55%_at_50%_0%,color-mix(in_srgb,var(--color-lime-9)_72%,transparent)_0%,color-mix(in_srgb,var(--color-lime-9)_28%,transparent)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    header: "relative z-1 flex items-start gap-4",
    icon: "size-8 shrink-0 text-sand-a9 [&_svg]:size-8",
    copy: "flex min-w-0 flex-1 flex-col justify-center gap-1",
    description: "text-sand-a9",
    body: "relative z-1 mt-auto flex flex-col items-stretch",
  },
});

export const scimProviderCard = tv({
  slots: {
    actions: "flex flex-wrap items-center gap-2",
    body: "flex flex-col gap-4",
    callout: "flex flex-col gap-1 rounded-2 bg-red-3 p-3",
    hint: "flex-1",
  },
});

export const scimConfiguration = tv({
  slots: {
    root: "flex flex-col gap-4",
    fields: "flex flex-col gap-3",
    field: "flex flex-col gap-1",
    fieldValue: "flex min-w-0 items-center gap-1",
    code: "min-w-0 flex-1 break-all",
    actions: "flex flex-wrap items-center gap-2",
    effects: "flex flex-col gap-3",
    effectsList: "list-disc ps-5",
  },
});

export const scimEventList = tv({
  slots: {
    root: "flex flex-col gap-4",
    results: "transition-opacity",
    empty: "flex flex-col items-center py-8 text-center",
    item: "relative items-stretch",
    row: "flex w-full min-w-0 flex-col",
    trigger: [
      "flex w-full items-center gap-3 text-start",
      // Overlay stretches the trigger across the padded ListItem. The panel
      // sits above it so request/response copy stays clickable when open.
      "after:absolute after:inset-0 after:content-['']",
      "data-open:[&_svg]:rotate-180",
    ],
    lead: "relative flex min-w-0 flex-1 flex-wrap items-center gap-2",
    trail: "relative ml-auto flex shrink-0 items-center gap-2",
    path: "min-w-0 truncate font-mono",
    caret: "relative size-4 shrink-0 text-sand-11 transition-transform duration-150",
    panel: "relative z-1 mt-2 flex flex-col gap-3",
    meta: "flex flex-wrap gap-4",
    metaField: "flex min-w-0 flex-col gap-0.5",
    block: "flex flex-col gap-1",
    blockHeading: "text-1 text-sand-11",
    responseWrap: "relative",
    response:
      "whitespace-pre-wrap break-all rounded-2 bg-sand-2 p-2 pr-9 text-1 text-sand-12",
    responseCopy: "absolute top-1.5 right-1.5",
    pager: "flex justify-center",
  },
  variants: {
    pending: {
      true: {
        results: "opacity-60",
      },
    },
  },
});
