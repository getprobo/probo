// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
