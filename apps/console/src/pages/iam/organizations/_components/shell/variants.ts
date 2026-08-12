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

// The organization frame, shared by OrganizationLayout and its skeleton so the
// loading state matches the loaded one.
//
// The frame owns the viewport: the page scrolls inside `content` rather than on
// the document, which is what keeps the top bar and both nav columns in place
// without any of them being `position: fixed`.
export const organizationLayout = tv({
  slots: {
    root: "flex h-dvh flex-col bg-sand-2",
    body: "flex min-h-0 flex-1",
    content: "min-w-0 flex-1 overflow-y-auto transition-[padding] duration-300",
    // Preserves the page metrics the v1 Layout established, so pages that have
    // not been converted yet keep their current measure and rhythm.
    contentInner: "mx-auto w-full max-w-[1200px] px-8 py-12",
  },
  variants: {
    // A v1 Drawer is open; reserve its width instead of letting it overlap.
    hasDrawer: {
      true: { content: "pr-105" },
    },
  },
});

export const topBar = tv({
  slots: {
    bar: "flex h-14 shrink-0 items-center gap-3 border-b border-sand-6 bg-sand-1 px-4",
    brand: "flex shrink-0 items-center",
    logo: "h-5 w-12 text-sand-12",
    separator: "text-sand-8",
    trailing: "ml-auto flex items-center gap-2",
  },
});

// Rounded pill naming the current organization, opening the switcher.
export const organizationSwitcher = tv({
  slots: {
    trigger: [
      "flex h-8 min-w-0 items-center gap-1 rounded-3 px-2",
      "cursor-pointer outline-none transition-colors select-none",
      "hover:bg-sand-3 data-popup-open:bg-sand-3",
      "focus-visible:ring-2 focus-visible:ring-sand-8",
    ],
    popup: "w-72 p-0",
    search: "p-2",
    // Bounded so a viewer in many organizations gets a scrolling list rather
    // than a menu taller than the viewport.
    list: "max-h-96 overflow-y-auto px-1 pb-1",
  },
});

// Rounded pill that opens the authenticated user menu.
export const viewerMembershipMenuTrigger = tv({
  base: [
    "flex h-8 items-center gap-2 rounded-full py-1 pr-2.5 pl-1",
    "cursor-pointer outline-none transition-colors select-none",
    "hover:bg-sand-3 data-popup-open:bg-sand-3",
    "focus-visible:ring-2 focus-visible:ring-sand-8",
  ],
});

// The product rail, which widens on hover to name its icons.
//
// It expands as an overlay rather than in flow: `root` holds the collapsed
// width in the flex row and the rail itself is absolute inside it, so opening
// it cannot shove the panel and the page sideways every time the pointer
// brushes past. `overflow-hidden` is what hides the labels while it is narrow;
// it clips vertically too, which is fine at nine products but would want
// `overflow-y-auto` if the rail ever outgrew a short viewport.
export const navRail = tv({
  slots: {
    root: "relative w-14 shrink-0",
    rail: [
      "group absolute inset-y-0 left-0 z-2 flex w-14 flex-col gap-1 overflow-hidden",
      "border-r border-sand-6 bg-sand-1 p-2",
      "transition-[width,box-shadow] duration-150 ease-out",
    ],
    item: [
      "flex h-10 w-full items-center rounded-3 text-sand-11 outline-none transition-colors",
      "hover:bg-sand-4 hover:text-sand-12",
      "focus-visible:ring-2 focus-visible:ring-sand-8",
    ],
    // A 40px box inside 8px of padding puts the icon's centre on the collapsed
    // rail's centre line, so widening the rail does not shift it.
    icon: "flex size-10 shrink-0 items-center justify-center",
    label: [
      "truncate pr-3 opacity-0 transition-opacity duration-150 ease-out",
      "group-hover:opacity-100 group-focus-within:opacity-100",
    ],
  },
  variants: {
    active: {
      true: { item: "bg-sand-5 text-sand-12" },
    },
    // Off for the skeleton, which would otherwise open onto nothing.
    expandable: {
      true: {
        rail: "hover:w-56 hover:shadow-4 focus-within:w-56 focus-within:shadow-4",
      },
    },
  },
  defaultVariants: {
    expandable: true,
  },
});

export const navPanel = tv({
  slots: {
    panel: "flex w-56 shrink-0 flex-col gap-1 border-r border-sand-6 bg-sand-1 px-3 py-4",
    title: "px-2 pb-2",
    list: "flex flex-col gap-0.5",
    // Nav entries read as a list, so they fill the column and align left
    // rather than sitting as centred pills like a row of buttons would.
    item: "w-full justify-start",
  },
});
