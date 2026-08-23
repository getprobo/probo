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

import { tv } from "tailwind-variants";

// Mirrors compliance-portal documents layout with console v1 tokens.
export const accessEntriesLayout = tv({
  slots: {
    page: "flex w-full flex-col gap-8 pb-28",
    results: "flex w-full flex-col gap-8",
  },
});

export const accessEntrySection = tv({
  slots: {
    root: "flex flex-col gap-3",
    header: "flex items-center gap-2.5",
    title: "text-sm font-medium text-txt-primary uppercase tracking-wide",
    count: "text-sm text-txt-tertiary",
  },
});

export const accessEntryList = tv({
  slots: {
    root: "list-none overflow-hidden rounded-[10px] border border-border-low bg-level-1",
    item: "flex flex-wrap items-center gap-4 border-b border-border-low px-4 py-3 last:border-b-0",
    content: "flex w-64 min-w-0 shrink-0 flex-col gap-0.5 max-md:flex-1",
    flags: "flex min-w-0 flex-1 flex-wrap items-center gap-1",
    // Fixed tracks so Admin/MFA/Auth/Last login line up across rows; the last
    // track matches EntryDecisionActions' w-36 Select.
    trailing: "ml-auto grid shrink-0 grid-cols-[3.5rem_3rem_8rem_5.5rem_9rem] items-center gap-x-4 max-md:ml-0 max-md:w-full max-md:grid-cols-2 max-md:gap-y-3",
    status: "flex min-w-0 flex-col items-center gap-0.5 overflow-hidden",
    statusLabel: "w-full truncate text-center text-[10px] font-medium uppercase tracking-wide text-txt-tertiary",
  },
  variants: {
    inactive: {
      true: { item: "opacity-70 line-through text-txt-tertiary" },
      false: {},
    },
  },
  defaultVariants: {
    inactive: false,
  },
});

export const accessEntriesSelectionBar = tv({
  slots: {
    bar: "fixed inset-x-0 bottom-0 z-20 border-t border-border-low bg-level-1/80 px-8 py-4 backdrop-blur max-md:px-4",
    inner: "mx-auto flex w-full flex-wrap items-center justify-between gap-4",
    actions: "flex flex-wrap items-center gap-2",
  },
});
