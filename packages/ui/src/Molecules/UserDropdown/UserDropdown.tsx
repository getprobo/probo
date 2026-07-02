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
