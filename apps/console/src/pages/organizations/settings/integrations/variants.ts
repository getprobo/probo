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

export const integrationsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
  },
});

export const integrationsList = tv({
  slots: {
    root: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-48 shrink-0",
    section: "flex flex-col gap-3",
    sectionTitle: "flex items-center gap-2.5",
    grid: "grid grid-cols-3 gap-3 max-xl:grid-cols-2 max-lg:grid-cols-1",
    marketplaceGrid: "grid grid-cols-4 gap-3 max-xl:grid-cols-2 max-lg:grid-cols-1",
    empty: "flex flex-col items-center py-8 text-center",
  },
});

export const connectorProviderStack = tv({
  slots: {
    root: "relative min-w-0",
    deck: "isolate grid",
    slide: "relative col-start-1 row-start-1 min-w-0",
    page: "pointer-events-none absolute right-3 bottom-3 z-1",
    next: "absolute bottom-2 left-1/2 z-1 -translate-x-1/2 translate-y-1/2",
  },
});

export const connectorCard = tv({
  slots: {
    card: "relative flex h-full min-w-0 flex-col",
    person: "flex items-center gap-4 px-4 py-4 pr-12",
    identity: "flex min-w-0 flex-1 flex-col gap-0.5",
    title: "min-w-0 truncate",
    menu: "absolute top-3 right-3",
    meta: "flex flex-col gap-1.5 px-4 py-3",
    metaRow: "flex flex-wrap items-center justify-between gap-2",
    controls: "flex items-center gap-1",
    usedBy: "flex items-center gap-2",
    usedByIcons: "flex items-center gap-1.5",
    connectActions: "ml-auto",
    organizationSelect: "w-44 shrink-0",
  },
});
