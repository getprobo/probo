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

import { tv, type VariantProps } from "tailwind-variants/lite";

// Ghost Card has no chrome; the frame owns border + fill so tone can swap hue
// without fighting Card's sand-a3 (tv/lite has no merge).
export const tonedCard = tv({
  slots: {
    frame: "flex h-full flex-col overflow-hidden rounded-4 border bg-sand-1",
    header: "relative flex w-full items-center justify-between gap-3 overflow-hidden px-5 py-4",
    wash: "pointer-events-none absolute top-1/2 left-1/2 aspect-square w-full -translate-x-1/2 -translate-y-1/2 opacity-10 blur-[14px]",
    fade: "pointer-events-none absolute inset-0 bg-linear-to-b from-sand-1/0 to-sand-1",
    icon: "relative z-1 flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-2",
    lead: "relative z-1 min-w-0 flex-1 truncate",
    control: "relative z-1 shrink-0",
    body: "flex flex-1 flex-col gap-2 px-5 pb-4",
  },
  variants: {
    tone: {
      sand: {
        frame: "border-sand-a3",
        wash: "hidden",
        fade: "hidden",
        icon: "bg-sand-1 text-sand-9",
      },
      green: {
        frame: "border-green-6",
        wash: "bg-green-9",
        icon: "bg-transparent text-green-11",
      },
      sky: {
        frame: "border-sky-6",
        wash: "bg-sky-9",
        icon: "bg-transparent text-sky-11",
      },
      amber: {
        frame: "border-amber-6",
        wash: "bg-amber-9",
        icon: "bg-transparent text-amber-11",
      },
      red: {
        frame: "border-red-6",
        wash: "bg-red-9",
        icon: "bg-transparent text-red-11",
      },
    },
  },
  defaultVariants: {
    tone: "sand",
  },
});

export type TonedCardTone = NonNullable<VariantProps<typeof tonedCard>["tone"]>;
