// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { clsx } from "clsx";
import type { PropsWithChildren, ReactNode } from "react";
import { Children, cloneElement, isValidElement } from "react";

export type AsChildProps<DefaultElementProps>
  = | ({ asChild?: false } & DefaultElementProps)
    | ({ asChild: true; children: ReactNode } & DefaultElementProps);

export function Slot({
  children,
  ...props
}: PropsWithChildren<{ className?: string | undefined } & Record<string, unknown>>) {
  if (isValidElement<{ className?: string | undefined } & Record<string, unknown>>(children)) {
    return cloneElement(children, {
      ...props,
      className: clsx(props.className, children.props.className),
    });
  }
  if (Children.count(children) > 1) {
    Children.only(null);
  }
  return null;
}
