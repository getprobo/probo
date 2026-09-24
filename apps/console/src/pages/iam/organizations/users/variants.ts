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

export const usersPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
  },
});

export const usersPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
    grid: "grid grid-cols-3 gap-3 max-xl:grid-cols-2 max-lg:grid-cols-1",
  },
});

export const usersList = tv({
  slots: {
    root: "flex flex-col gap-4",
    tools: "flex flex-wrap items-center justify-between gap-2",
    search: "w-80 max-sm:min-w-0 max-sm:w-full",
    filters: "flex flex-wrap items-center justify-end gap-2",
    filter: "w-40 shrink-0",
    results: "transition-opacity",
    grid: "grid grid-cols-3 gap-3 max-xl:grid-cols-2 max-lg:grid-cols-1",
    empty: "flex flex-col items-center py-8 text-center",
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

export const userListItem = tv({
  slots: {
    card: "relative flex h-full min-w-0 flex-col",
    overlay: "absolute inset-0 z-0",
    person: "flex items-center gap-4 px-4 py-4 pr-12",
    avatar: "relative shrink-0",
    source: "pointer-events-none absolute -right-0.5 -bottom-0.5 z-1 origin-bottom-right scale-75",
    menu: "absolute top-3 right-3 z-1",
    identity: "flex min-w-0 flex-1 flex-col gap-0",
    kind: "min-w-0 truncate leading-none",
    title: "min-w-0 truncate leading-tight",
    email: "min-w-0 truncate leading-tight",
    meta: "flex flex-col gap-1.5 px-4 py-3",
    metaRow: "flex flex-wrap items-center justify-between gap-2",
    role: "relative z-1 flex items-center gap-1.5",
    roleSelect: "w-32 shrink-0",
    contract: "min-w-0",
  },
  variants: {
    inactive: {
      true: {
        person: "opacity-50",
        meta: "opacity-50",
      },
    },
  },
  defaultVariants: {
    inactive: false,
  },
});

export const userPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    header: "flex items-start justify-between gap-4",
    identity: "flex min-w-0 flex-1 flex-col gap-1",
    titleRow: "flex min-w-0 flex-wrap items-center gap-2",
    title: "min-w-0 truncate",
    email: "min-w-0 truncate",
    actions: "flex shrink-0 items-center gap-2",
  },
});

export const userIdentitySection = tv({
  slots: {
    root: "flex flex-col gap-4",
    person: "flex items-start gap-4 max-sm:flex-col",
    avatar: "flex w-32 shrink-0 flex-col gap-2",
    dropzone: "relative",
    source: "pointer-events-none absolute -right-0.5 -bottom-0.5 z-1 origin-bottom-right scale-75",
    hint: "text-pretty",
    fields: "flex min-w-0 flex-1 flex-col gap-3",
  },
  variants: {
    inactive: {
      true: {
        person: "opacity-50",
      },
    },
  },
  defaultVariants: {
    inactive: false,
  },
});

export const userPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    heading: "flex min-w-0 flex-1 flex-col gap-1",
    identity: "flex flex-col gap-4",
    person: "flex items-start gap-4 max-sm:flex-col",
    fields: "flex min-w-0 flex-1 flex-col gap-3",
    properties: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    dates: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 [&>*]:min-w-0",
  },
});

export const userPropertiesSection = tv({
  slots: {
    root: "flex flex-col gap-4",
    intro: "flex flex-col gap-1",
    fields: "flex flex-col gap-4",
    dates: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 [&>*]:min-w-0",
  },
});

export const userEmailsField = tv({
  slots: {
    root: "flex flex-col gap-2",
    row: "flex items-center gap-2",
    field: "min-w-0 flex-1",
    add: "self-start",
  },
});

export const newUserPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    intro: "flex min-w-0 flex-col gap-2",
    form: "flex flex-col gap-4",
    dates: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 [&>*]:min-w-0",
    actions: "flex justify-end",
  },
});

export const newUserPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    intro: "flex min-w-0 flex-col gap-2",
    form: "flex flex-col gap-4",
    dates: "grid grid-cols-2 gap-3 max-sm:grid-cols-1 [&>*]:min-w-0",
  },
});
