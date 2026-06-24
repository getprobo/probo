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
import type { VariantProps } from "tailwind-variants/lite";

import { heading } from "./variants";

type HeadingLevel = 1 | 2 | 3 | 4 | 5 | 6;

export type HeadingProps = ComponentProps<"h1"> & VariantProps<typeof heading> & {
  // The rendered heading element (h1–h6), driving document outline. Decoupled
  // from `size`, which controls the visual scale. Defaults to h1.
  level?: HeadingLevel;
};

// Foundational heading primitive (Radix "Heading"). Renders the semantic
// h1–h6 chosen by `level`; tune the visual scale with `size`. See
// contrib/claude/ui.md (typography).
export function Heading(props: HeadingProps) {
  const { level = 1, size, weight, align, color, highContrast, className, ...rest } = props;
  const Tag = `h${level}` as const;

  return (
    <Tag
      className={heading({ size, weight, align, color, highContrast, className })}
      {...rest}
    />
  );
}
