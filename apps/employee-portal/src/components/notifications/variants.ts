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

// The list height is capped by the popup's available space so the panel
// scrolls instead of leaving the viewport. `--available-height` comes from
// the popover positioner. Further pages load when the sentinel scrolls
// into view.
export const notificationsPanel = tv({
  slots: {
    popup: "flex w-80 max-h-(--available-height) flex-col overflow-hidden p-0",
    header: "flex items-center justify-between gap-2 px-3 pt-3 pb-2",
    actions: "flex items-center gap-2",
    action: "cursor-pointer text-sand-11 hover:text-sand-12",
    empty: "flex flex-col items-center gap-2 px-3 pt-2 pb-4 text-center",
    emptyIcon: "text-sand-11",
    list: "flex max-h-[min(24rem,calc(var(--available-height)-3.5rem))] min-h-0 flex-col overflow-y-auto overscroll-y-contain px-1 pb-1",
    sentinel: "h-px w-full shrink-0",
    item: "flex items-start gap-2 rounded-2 px-2 py-2 hover:bg-sand-3",
    body: "flex min-w-0 flex-col gap-0.5",
    dot: "mt-1.5 size-1.5 shrink-0 rounded-full",
    count: "pointer-events-none absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-gold-9 px-1 text-white",
    trigger: "data-popup-open:bg-sand-3",
    triggerWrap: "relative inline-flex",
  },
  variants: {
    unread: {
      true: { dot: "bg-gold-9" },
      false: {},
    },
  },
});
