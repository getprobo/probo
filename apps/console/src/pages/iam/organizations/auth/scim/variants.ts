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

export const scimPage = tv({
  slots: {
    root: "flex flex-col gap-6",
    header: "flex items-start justify-between gap-4",
    intro: "flex min-w-0 flex-col gap-2",
    grid: "grid grid-cols-1 gap-3 md:grid-cols-3",
  },
});

export const scimSetupCard = tv({
  slots: {
    frame: "relative flex h-full flex-col gap-6 overflow-hidden px-6 pt-10 pb-6",
    wash: [
      "pointer-events-none absolute inset-x-0 top-0 z-0 h-3/5",
      "bg-[radial-gradient(ellipse_85%_55%_at_50%_0%,color-mix(in_srgb,var(--color-lime-9)_72%,transparent)_0%,color-mix(in_srgb,var(--color-lime-9)_28%,transparent)_35%,transparent_62%)]",
      "mask-[linear-gradient(to_bottom,black_0%,black_40%,transparent_100%)]",
    ],
    header: "relative z-1 flex items-start gap-4",
    icon: "size-8 shrink-0 text-sand-a9 [&_svg]:size-8",
    copy: "flex min-w-0 flex-1 flex-col justify-center gap-1",
    description: "text-sand-a9",
    body: "relative z-1 mt-auto flex flex-col items-stretch",
  },
});
