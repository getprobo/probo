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

// Employee Portal top bar. Slots are shared by the live TopBar and its
// skeleton so the loading placeholder is structurally identical. Below lg
// (<1024px) the brand shows logo only — name + tagline return at lg+.
// justify-between keeps the user menu right-aligned.
export const topBar = tv({
  slots: {
    bar: "flex h-14 items-center bg-sand-1 px-8 max-md:px-4",
    inner: "flex w-full items-center justify-between gap-6 max-md:gap-3",
    brand: "flex min-w-0 items-center gap-2",
    brandText: "flex min-w-0 flex-col max-lg:hidden",
    brandName: "truncate",
    tagline: "truncate",
    logo: "shrink-0",
  },
});

// User-menu trigger. Radius matches v2 Button size 2 and Avatar `small`.
export const topBarUserMenuTrigger = tv({
  base: [
    "flex h-8 items-center gap-2 rounded-2 py-1 pr-2.5 pl-1",
    "cursor-pointer outline-none transition-colors select-none",
    "hover:bg-sand-3 data-popup-open:bg-sand-3",
    "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
  ],
});
