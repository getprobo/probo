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

export const connectorCard = tv({
  slots: {
    card: "relative h-full min-w-0",
    identity: "flex min-w-0 flex-1 flex-col gap-2",
    name: "flex min-w-0 items-baseline gap-1.5",
    title: "min-w-0 truncate leading-tight",
    controls: "flex items-center gap-1",
    metaRow: "flex flex-wrap items-center justify-between gap-2",
    tags: "flex flex-wrap items-center gap-2",
    usedBy: "flex items-center gap-2",
    usedByIcons: "flex items-center gap-1.5",
    organizationSelect: "w-44 shrink-0",
    probeError: "pointer-events-auto flex flex-col gap-2 rounded-2 bg-sand-3 p-3",
    probeErrorHeader: "flex items-center justify-between gap-2",
    probeErrorRow: "flex min-w-0 items-start gap-1",
    probeErrorText: "min-w-0 flex-1",
  },
});

export const connectorDetailsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    title: "flex min-w-0 flex-wrap items-center gap-2",
    grid: "grid grid-cols-2 gap-3 max-lg:grid-cols-1",
  },
});

export const connectorAccountsDrawer = tv({
  slots: {
    heading: "flex min-w-0 flex-1 flex-col gap-1",
    title: "flex min-w-0 items-center gap-2.5",
    list: "flex flex-col gap-3",
    empty: "flex flex-col items-center py-8 text-center",
    actions: "flex items-center gap-2",
  },
});

export const connectorDetailsPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    title: "flex min-w-0 items-center gap-2",
  },
});
