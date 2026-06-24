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

import { type HTMLAttributes, useEffect, useRef } from "react";
import { Link } from "react-router";
import { tv } from "tailwind-variants";

type Props = {
  active?: boolean;
  id: string;
  description?: string;
  to: string;
} & HTMLAttributes<HTMLAnchorElement>;

const classNames = tv({
  slots: {
    wrapper: "block p-4 space-y-[6px] rounded-xl cursor-pointer text-start",
    id: "px-[6px] py-[2px] text-base font-medium border border-border-low rounded-lg w-max",
    description: "text-sm text-txt-tertiary line-clamp-3",
  },
  variants: {
    active: {
      true: {
        wrapper: "bg-tertiary-pressed",
        id: "bg-active",
      },
      false: {
        wrapper: "hover:bg-tertiary-hover",
        id: "bg-highlight",
      },
    },
  },
  defaultVariants: {
    active: false,
  },
});

export function ControlItem({ active, id, description, to, ...props }: Props) {
  const {
    wrapper,
    id: idCls,
    description: descriptionCls,
  } = classNames({
    active,
  });

  const ref = useRef<HTMLAnchorElement>(null);

  // Make the active element scroll into view when selected
  useEffect(() => {
    if (ref.current && active) {
      ref.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [active]);

  return (
    <Link className={wrapper()} to={to} {...props} ref={ref}>
      <div className={idCls()}>{id}</div>
      <div className={descriptionCls()}>{description}</div>
    </Link>
  );
}
