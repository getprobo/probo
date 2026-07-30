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

// Full-bleed portal banner under the TopBar: message row + actions, stacking on
// small screens. Content column matches TopBar / page layout (px-8 + centered
// max-w-5xl). Color tints the band and leading icon (sky = info, amber = warning).
export const banner = tv({
  slots: {
    root: "w-full px-8 py-2.5 max-md:px-4 max-md:py-3",
    inner:
      "mx-auto flex w-full max-w-5xl items-center gap-3 max-md:flex-col max-md:items-stretch max-md:gap-3",
    content: "flex min-w-0 flex-1 items-start gap-2",
    icon: "mt-0.5 shrink-0 [&_svg]:size-4",
    message: "min-w-0 flex-1",
    dismissMobile: "md:hidden",
    // Full-width CTAs on small screens; leave the trailing dismiss control alone.
    actions:
      "flex shrink-0 items-center gap-2 max-md:flex-col max-md:items-stretch [&>*:not(:last-child)]:max-md:w-full",
    dismissDesktop: "max-md:hidden",
  },
  variants: {
    color: {
      sky: {
        root: "bg-sky-3",
        icon: "text-sky-11",
      },
      amber: {
        root: "bg-amber-3",
        icon: "text-amber-11",
      },
    },
  },
});
