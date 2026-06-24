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

import { useCardContext } from "./context";
import { cardInset } from "./variants";

export type CardInsetProps = ComponentProps<"div"> & {
  // Which edges the content bleeds to. Defaults to all edges.
  side?: "all" | "x" | "y" | "top" | "bottom";
};

// Content that bleeds past the Card's padding to its edges (Radix "Inset"),
// e.g. a cover image. Negates the parent Card's resolved padding (read from
// context) so it always lines up, even when padding is decoupled from size. With
// `padding="none"` there is nothing to negate, so this is a no-op wrapper.
export function CardInset(props: CardInsetProps) {
  const { side = "all", className, ...rest } = props;
  const padding = useCardContext();

  return <div className={cardInset({ padding, side, className })} {...rest} />;
}
