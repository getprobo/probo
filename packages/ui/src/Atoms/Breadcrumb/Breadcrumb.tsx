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
import { Fragment } from "react/jsx-runtime";
import { Link } from "react-router";

import { IconChevronRight } from "../Icons";

type Props = {
  items: (ItemProps | string)[];
};

type ItemProps = {
  label: string;
  to?: string;
  active?: boolean;
};

export function Breadcrumb({ items }: Props) {
  return (
    <div className="flex items-center gap-[6px] text-txt-tertiary">
      {items.map((item, k) => (
        <Fragment key={k}>
          {k > 0 && <IconChevronRight size={12} />}
          <BreadcrumbItem
            {...(typeof item === "string" ? { label: item } : item)}
            active={k === items.length - 1}
          />
        </Fragment>
      ))}
    </div>
  );
}

function BreadcrumbItem({ label, to, active }: ItemProps) {
  const className = clsx(
    "text-sm px-1 rounded-sm h-5",
    active && "text-txt-primary font-medium",
  );
  if (to) {
    return (
      <Link
        to={to}
        className={clsx(
          className,
          "hover:bg-tertiary-hover active:bg-tertiary-pressed",
        )}
      >
        {label}
      </Link>
    );
  }
  return <span className={className}>{label}</span>;
}
