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

// Tabs (Radix "Tabs" over Base UI's Tabs). An underlined tab bar whose active
// item is tracked by a sliding indicator positioned from Base UI's
// `--active-tab-*` CSS variables.

export const tabsList = tv({
  base: "relative flex items-center gap-1 border-b border-sand-a3",
});

export const tabsTab = tv({
  base: [
    "relative flex h-full cursor-pointer items-center justify-center gap-2 px-2 py-4 text-2 text-sand-a11",
    "select-none outline-none transition-colors",
    "hover:text-sand-12",
    "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    "data-active:font-medium data-active:text-sand-12",
    "data-disabled:pointer-events-none data-disabled:opacity-50",
  ],
});

export const tabsIndicator = tv({
  base: [
    "absolute bottom-0 left-0 h-[2px] w-(--active-tab-width) rounded-1 bg-sand-12",
    "translate-x-(--active-tab-left) transition-all duration-200 ease-out",
  ],
});

export const tabsSkeleton = tv({
  slots: {
    root: "flex items-center gap-4 border-b border-sand-a3",
    item: "my-4 h-5 animate-pulse rounded-1 bg-sand-3",
  },
});
