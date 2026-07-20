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

// Landing hero content rendered inside the shared HeaderBand: a centered
// title/description section above an optional bottom slot (the contact row).
// Slots are shared by the live Hero and its skeleton.
export const hero = tv({
  slots: {
    content: "flex w-full flex-col gap-10",
    section: "flex w-full flex-col gap-4",
  },
});

// Organization contact block: a horizontal row of icon + label items, with a
// top divider so it reads as the hero's bottom section. Self-contained so the
// divider only appears when there is contact info to show.
export const organizationContactInfo = tv({
  slots: {
    // Only a top divider gap; the band's py-8 provides the bottom spacing.
    root: "flex w-full flex-wrap items-center gap-x-6 gap-y-2 border-t border-sand-a2 pt-4 max-md:flex-col max-md:items-start",
    item: "flex min-w-0 items-center gap-2 text-sand-11 [&_svg]:size-5 [&_svg]:shrink-0 [&>*]:min-w-0 [&>*:not(svg)]:break-words",
    link: "flex min-w-0 items-center gap-2 text-sand-11 hover:underline [&_svg]:size-5 [&_svg]:shrink-0 [&>*]:min-w-0 [&>*:not(svg)]:break-all",
  },
});
