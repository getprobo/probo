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

import type { Icon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Link } from "react-router";

import { navRail } from "./variants";

export interface NavRailItemProps {
  icon: Icon;
  label: string;
  to: string;
  active: boolean;
}

/**
 * One product in the rail: an icon, and a label that fades in once the rail
 * opens.
 *
 * The label is always in the DOM rather than swapped in on hover, so it names
 * the link for assistive technology even while it is visually hidden. That is
 * why there is no `aria-label` here.
 */
export function NavRailItem({ icon: IconComponent, label, to, active }: NavRailItemProps) {
  const slots = navRail({ active });

  return (
    <Link to={to} className={slots.item()}>
      <span className={slots.icon()}>
        <IconComponent size={20} aria-hidden />
      </span>
      <Text size={2} className={slots.label()}>
        {label}
      </Text>
    </Link>
  );
}
