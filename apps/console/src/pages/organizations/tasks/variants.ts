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

export const taskDetailsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
    titleRow: "flex flex-wrap items-start justify-between gap-3",
    title: "flex min-w-0 flex-1 items-center gap-2",
    actions: "flex shrink-0 flex-wrap items-center gap-2",
    body: "grid grid-cols-1 items-start gap-6 lg:grid-cols-[1fr_360px]",
  },
});

export const taskDetailsPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex flex-col gap-2",
    titleRow: "flex flex-wrap items-start justify-between gap-3",
    title: "flex min-w-0 items-center gap-2",
    body: "grid grid-cols-1 items-start gap-6 lg:grid-cols-[1fr_360px]",
    description: "flex flex-col gap-2",
  },
});

export const taskPropertiesSection = tv({
  slots: {
    root: "flex flex-col",
    row: "grid grid-cols-[7.5rem_minmax(0,1fr)] items-center gap-3 border-b border-sand-6 py-2.5 last:border-b-0",
    value: "flex min-w-0 items-center gap-2",
  },
});

export const taskNameField = tv({
  slots: {
    root: "min-w-0 flex-1",
    input: [
      "w-full border-0 bg-transparent p-0 text-6 font-medium text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
  },
});

export const taskDescriptionSection = tv({
  slots: {
    root: "min-w-0",
    textarea: [
      "w-full field-sizing-content resize-none overflow-hidden border-0 bg-transparent p-0 text-2 text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
  },
});

export const taskDurationField = tv({
  slots: {
    root: "grid grid-cols-[4.5rem_minmax(0,1fr)_auto] items-center gap-2",
  },
});

export const createTaskDialog = tv({
  slots: {
    form: "flex flex-col gap-4",
    fields: "flex flex-col gap-4",
    value: "flex items-center gap-2",
  },
});
