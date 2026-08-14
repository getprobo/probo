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

import type { Icon, IconWeight } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Link } from "react-router";

import { navRail } from "./variants";

type NavRailItemBase = {
  icon: Icon;
  label: string;
  active?: boolean;
  weight?: IconWeight;
};

type NavRailItemAction
  = { to: string; href?: never; onClick?: never }
    | { href: string; to?: never; onClick?: never }
    | { onClick: () => void; to?: never; href?: never };

export type NavRailItemProps = NavRailItemBase & NavRailItemAction;

export function NavRailItem({
  icon: IconComponent,
  label,
  active = false,
  weight = "duotone",
  ...rest
}: NavRailItemProps) {
  const slots = navRail({ active });
  const body = (
    <>
      <span className={slots.icon()}>
        <IconComponent size={20} weight={weight} aria-hidden />
      </span>
      <Text size={2} className={slots.label()}>
        {label}
      </Text>
    </>
  );

  if ("onClick" in rest) {
    return (
      <button type="button" className={slots.item()} onClick={rest.onClick}>
        {body}
      </button>
    );
  }

  if ("href" in rest) {
    return <a href={rest.href} className={slots.item()}>{body}</a>;
  }

  return <Link to={rest.to} className={slots.item()}>{body}</Link>;
}
