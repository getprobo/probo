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
