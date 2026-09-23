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

export const webhooksSettingsPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
  },
});

export const webhooksSettingsPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    grid: "grid grid-cols-2 gap-3 max-md:grid-cols-1",
  },
});

export const webhookSubscriptionList = tv({
  slots: {
    root: "flex flex-col gap-4",
    results: "transition-opacity",
    grid: "grid grid-cols-2 gap-3 max-md:grid-cols-1",
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

export const webhookSubscriptionListItem = tv({
  slots: {
    card: "relative flex h-full flex-col gap-4",
    header: "flex items-center justify-between gap-3",
    lead: "flex min-w-0 flex-1 items-center gap-3",
    icon: "flex size-10 shrink-0 items-center justify-center rounded-2 bg-sand-3 text-sand-11 [&_svg]:size-5",
    endpoint: [
      "min-w-0 flex-1 break-all",
      "after:absolute after:inset-0 after:content-['']",
    ],
    action: "relative z-1",
    eventsSection: "flex flex-col gap-2",
    eventsHeading: "flex items-center gap-2",
    events: "flex flex-wrap items-center gap-1.5",
    eventsAll: "flex max-w-80 flex-wrap content-start gap-1.5",
    badge: "font-mono",
    activity: "mt-auto flex items-center gap-2",
  },
});

export const webhookEventTypeSelect = tv({
  slots: {
    popup: "flex w-80 flex-col gap-2",
    results: "flex max-h-64 flex-col gap-1 overflow-y-auto",
    row: "flex cursor-pointer items-center gap-2 rounded-2 px-2 py-1.5 text-2 text-sand-12 hover:bg-sand-3",
    label: "min-w-0 font-mono",
    empty: "px-2 py-1 text-1 text-sand-11",
  },
});

export const deleteWebhookSubscriptionDialog = tv({
  slots: {
    body: "flex flex-col gap-2",
  },
});

export const newWebhookSubscriptionPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    intro: "flex min-w-0 flex-col gap-2",
    form: "flex flex-col gap-4",
    actions: "flex justify-end",
  },
});

export const newWebhookSubscriptionPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    intro: "flex min-w-0 flex-col gap-2",
    form: "flex flex-col gap-4",
  },
});

export const webhookSubscriptionDetailPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    back: "self-start",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    title: "min-w-0 break-all",
    settings: "flex flex-col gap-4",
    endpointRow: "flex items-end gap-2",
    endpointField: "min-w-0 flex-1",
    eventsSection: "flex flex-col gap-2",
    eventsHeading: "flex items-center gap-2",
    history: "flex flex-col gap-3",
  },
});

export const webhookSubscriptionDetailPageSkeleton = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    settings: "flex flex-col gap-4",
    history: "flex flex-col gap-3",
  },
});

export const webhookSigningSecretField = tv({
  slots: {
    root: "flex flex-col gap-2",
    header: "flex items-center gap-2",
    value: "min-w-0 max-w-full break-all",
    actions: "flex shrink-0 items-center gap-1",
  },
});

export const webhookSubscriptionEventList = tv({
  slots: {
    root: "flex flex-col gap-4",
    results: "transition-opacity",
    empty: "flex flex-col items-center py-8 text-center",
    item: "items-stretch",
    row: "flex w-full min-w-0 flex-col",
    trigger: [
      "flex w-full items-center gap-3",
      "data-open:[&_svg]:rotate-180",
    ],
    lead: "flex min-w-0 items-center gap-2",
    trail: "ml-auto flex min-w-0 items-center gap-2",
    contentType: "min-w-0 truncate",
    caret: "size-4 shrink-0 text-sand-11 transition-transform duration-150",
    response:
      "mt-2 whitespace-pre-wrap break-all rounded-2 bg-sand-2 p-2 text-1 text-sand-12",
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
