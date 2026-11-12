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
  base: "group/review-card relative rounded-3 border bg-sand-1 p-3",
  variants: {
    active: {
      true: "border-sand-8 shadow-2",
      false: "border-sand-6",
    },
  },
});

export interface DocumentReviewMarginCardProps {
  active: boolean;
  label: string;
  children: ReactNode;
  onActivate?: () => void;
}

export function DocumentReviewMarginCard({
  active,
  label,
  children,
  onActivate,
}: DocumentReviewMarginCardProps) {
  return (
    <article className={reviewCard({ active })}>
      {onActivate
        ? (
            <button
              type="button"
              aria-label={label}
              aria-pressed={active}
              className="absolute inset-0 z-0 rounded-3 outline-none focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1"
              onMouseDown={event => event.preventDefault()}
              onClick={onActivate}
            />
          )
        : null}
      <div className="pointer-events-none relative z-10 flex flex-col gap-2 [&_a]:pointer-events-auto [&_button]:pointer-events-auto [&_input]:pointer-events-auto [&_label]:pointer-events-auto [&_textarea]:pointer-events-auto [&_[contenteditable=true]]:pointer-events-auto">
        {children}
      </div>
    </article>
  );
}
