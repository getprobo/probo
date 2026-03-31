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

import type { FC, PropsWithChildren } from "react";
import { Link } from "react-router";

import { Avatar } from "../../Atoms/Avatar/Avatar";
import { Button } from "../../Atoms/Button/Button";
import {
  Dropdown,
  DropdownItem,
  DropdownSeparator,
} from "../../Atoms/Dropdown/Dropdown";
import { IconChevronDown } from "../../Atoms/Icons";
import type { IconProps } from "../../Atoms/Icons/type";

type Props = PropsWithChildren<{ fullName: string; email: string }>;

export function UserDropdown({ fullName, children, email }: Props) {
  return (
    <Dropdown
      className="w-60"
      toggle={(
        <Button variant="tertiary">
          <Avatar name={fullName} />
          <span>{fullName}</span>
          <IconChevronDown size={16} />
        </Button>
      )}
    >
      <div className="flex gap-2 items-center">
        <Avatar name={fullName} size="l" />
        <div>
          <p className="text-sm font-medium text-txt-primary">
            {fullName}
          </p>
          <p className="text-xxs text-txt-tertiary">{email}</p>
        </div>
      </div>
      <DropdownSeparator />
      {children}
    </Dropdown>
  );
}

export function UserDropdownItem({
  to,
  icon: IconComponent,
  label,
  variant = "tertiary",
  onClick,
}: {
  to: string;
  icon: FC<IconProps>;
  label: string;
  variant?: "tertiary" | "danger";
  onClick?: React.MouseEventHandler<HTMLAnchorElement>;
}) {
  return (
    <DropdownItem asChild variant={variant}>
      <Link to={to} onClick={onClick}>
        {IconComponent && <IconComponent size={16} />}
        {label}
      </Link>
    </DropdownItem>
  );
}
