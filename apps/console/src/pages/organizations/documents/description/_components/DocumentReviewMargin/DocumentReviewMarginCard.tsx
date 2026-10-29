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

import { type ReactNode } from "react";
import { tv } from "tailwind-variants/lite";

const reviewCard = tv({
  base: "flex flex-col gap-2 rounded-3 border bg-sand-1 p-3",
  variants: {
    active: {
      true: "border-sand-8 shadow-2",
      false: "border-sand-6",
    },
  },
});

export interface DocumentReviewMarginCardProps {
  active: boolean;
  children: ReactNode;
  onActivate: () => void;
}

export function DocumentReviewMarginCard({
  active,
  children,
  onActivate,
}: DocumentReviewMarginCardProps) {
  return (
    <article
      className={reviewCard({ active })}
      onMouseDown={(event) => {
        const target = event.target;
        if (target instanceof HTMLElement && target.closest("textarea, input, button, a")) return;
        event.preventDefault();
      }}
      onClick={onActivate}
    >
      {children}
    </article>
  );
}
