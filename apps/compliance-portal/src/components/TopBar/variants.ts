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

// Trust Center top navigation bar. Slots are shared by the live TopBar and its
// skeleton so the loading placeholder is structurally identical. Desktop-first:
// unprefixed classes are the desktop layout; max-md: collapses into the burger.
export const topBar = tv({
  slots: {
    bar: "flex h-14 items-center bg-sand-1 px-8 max-md:px-4",
    inner: "mx-auto flex w-full max-w-5xl items-center gap-10 max-md:gap-3",
    brand: "flex min-w-0 items-center gap-2",
    brandName: "truncate",
    tagline: "max-md:hidden",
    logo: "shrink-0",
    spacer: "h-px flex-1",
    nav: "flex items-center gap-1 max-md:hidden",
    // Wrapper (not the IconButton) so kit `inline-flex` cannot override `hidden`.
    menuTrigger: "hidden max-md:block",
  },
});

// Rounded pill that opens the authenticated user menu.
export const topBarUserMenuTrigger = tv({
  base: [
    "flex h-8 items-center gap-2 rounded-full py-1 pr-2.5 pl-1",
    "cursor-pointer outline-none transition-colors select-none",
    "hover:bg-sand-3 data-[popup-open]:bg-sand-3",
    "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
  ],
});

// Right-edge mobile navigation drawer (Base UI Drawer).
export const topBarMobileNav = tv({
  slots: {
    backdrop: [
      "fixed inset-0 z-50 bg-sand-12/40",
      "transition-opacity duration-200",
      "data-starting-style:opacity-0 data-ending-style:opacity-0",
    ],
    viewport: "fixed inset-0 z-50 flex justify-end",
    popup: [
      "relative flex h-full w-[min(20rem,100%)] flex-col bg-sand-1 shadow-6 outline-none",
      "[transform:translateX(var(--drawer-swipe-movement-x,0px))]",
      "transition-transform duration-200",
      "data-starting-style:translate-x-full data-ending-style:translate-x-full",
    ],
    content: "flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-4",
    header: "flex items-center justify-between gap-3",
    title: "text-4 font-medium text-sand-12",
    nav: "flex flex-col gap-1",
    actions: "mt-auto flex flex-col gap-2 border-t border-sand-a3 pt-4",
  },
});
