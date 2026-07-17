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

import type { ComponentProps } from "react";

import { textArea } from "./variants";

export type TextareaProps
  = & Omit<ComponentProps<"textarea">, "className">
    & {
      // Applied to the bordered container (the top-level element).
      className?: string;
      // Surface treatment (defaults to "surface").
      variant?: "classic" | "surface" | "soft";
    };

// Multi-line text input on a bordered surface mirroring TextField. The container
// is the top-level node; all native textarea props spread onto the inner control.
export function Textarea(props: TextareaProps) {
  const { className, variant, rows = 4, ...textareaProps } = props;
  const { root, textarea } = textArea({ variant });

  return (
    <div className={root({ className })}>
      <textarea className={textarea()} rows={rows} {...textareaProps} />
    </div>
  );
}
