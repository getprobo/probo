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
  // Products stay duotone so they read as a set. Help is a utility, not a
  // product, so it passes regular and keeps the column from looking like
  // another destination.
  weight?: IconWeight;
};

export type NavRailItemProps =
  | (NavRailItemBase & { to: string; href?: never })
  | (NavRailItemBase & { href: string; to?: never });

/**
 * One row in the rail: an icon, and a label that fades in once the rail
 * opens. Product entries pass `to`; Help passes `href` because it is a
 * mailto, not a route.
 *
 * The label is always in the DOM rather than swapped in on hover, so it names
 * the link for assistive technology even while it is visually hidden. That is
 * why there is no `aria-label` here.
 */
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
        {/* Duotone rather than outlined or filled: at 20px in a narrow column
            thin strokes go muddy, but a flat fill loses the interior detail
            that tells the products apart. */}
        <IconComponent size={20} weight={weight} aria-hidden />
      </span>
      <Text size={2} className={slots.label()}>
        {label}
      </Text>
    </>
  );

  if ("href" in rest) {
    return <a href={rest.href} className={slots.item()}>{body}</a>;
  }

  return <Link to={rest.to} className={slots.item()}>{body}</Link>;
}
