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
