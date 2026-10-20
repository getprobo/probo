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

export const sourceListItem = tv({
  slots: {
    card: "flex h-full min-w-0 flex-col gap-3",
    header: "flex items-center gap-4",
    identity: "flex min-w-0 flex-1 flex-col gap-0.5",
    title: "min-w-0 truncate",
    logo: "size-8 shrink-0",
    actions: "flex flex-wrap items-center justify-end gap-2",
    organizationSelect: "w-44 shrink-0",
  },
});

export const connectionIssue = tv({
  slots: {
    root: "min-w-0 flex-1",
    body: "flex flex-wrap items-center gap-2",
    copy: "flex min-w-48 flex-1 flex-col gap-0.5",
  },
});

export const organizationsEmpty = tv({
  slots: {
    root: "flex max-w-80 flex-col gap-2",
    copy: "flex flex-col gap-1",
  },
});

export const manualOrgInput = tv({
  slots: {
    field: "max-w-40",
  },
});

export const addableConnectorCard = tv({
  slots: {
    title: "min-w-0 truncate",
    controls: "flex items-center gap-1",
    checkbox: "",
    menu: "pointer-events-auto relative z-1",
    badges: "mt-auto flex flex-wrap items-center gap-2",
    frame: "",
  },
  variants: {
    selectable: {
      true: {
        frame: [
          "pointer-events-none select-none",
          "has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-inset has-[:focus-visible]:ring-sand-8",
          "[&>button]:focus-visible:ring-0 [&>button]:focus-visible:ring-offset-0",
        ],
        checkbox: "pointer-events-none",
      },
    },
  },
});

export const csvSourcePage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    intro: "flex min-w-0 flex-col gap-2",
    form: "flex flex-col gap-4",
    actions: "flex items-center justify-end gap-2",
  },
});

export const sourcesPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    actions: "flex shrink-0 items-center gap-2",
    list: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    section: "flex flex-col gap-3",
    sectionTitle: "flex items-center gap-2.5",
    grid: "grid grid-cols-3 gap-3 max-xl:grid-cols-2 max-lg:grid-cols-1",
    empty: "flex flex-col items-center py-8 text-center",
    pager: "flex justify-center",
  },
});
