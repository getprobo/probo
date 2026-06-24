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
import {
  type RefCallback,
  type TextareaHTMLAttributes,
  useCallback,
  useLayoutEffect,
  useRef,
} from "react";

import { input } from "../Input/Input";

type Props = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  variant?: "bordered" | "ghost" | "title";
  autogrow?: boolean;
  ref?: RefCallback<HTMLTextAreaElement>;
};

export function Textarea(props: Props) {
  const ref = useRef<HTMLTextAreaElement>(null);
  const { autogrow, ref: propsRef, ...restProps } = props;

  const adjustHeight = useCallback(() => {
    if (!autogrow || !ref.current) return;
    ref.current.style.height = "inherit";
    const paddingY = 2;
    ref.current.style.height = `${ref.current.scrollHeight + paddingY * 2}px`;
  }, [autogrow, ref]);

  useLayoutEffect(() => {
    adjustHeight();
  }, [adjustHeight]);

  return (
    <textarea
      {...restProps}
      ref={(node) => {
        ref.current = node;
        propsRef?.(node);
      }}
      onInput={(e) => {
        adjustHeight();
        props.onInput?.(e);
      }}
      className={input({
        ...props,
        className: clsx("min-h-20", props.className),
      })}
    />
  );
}
