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

import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { useLocation } from "react-router";

import { navPanel } from "./variants";

export interface NavPanelItemProps {
  label: string;
  to: string;
  exact?: boolean;
}

/**
 * One entry of the active product.
 *
 * Active state covers descendants, so a detail page keeps its list highlighted
 * (`/governance/frameworks/:id` still marks Frameworks). `exact` opts out when
 * a sibling switcher owns those descendant URLs.
 */
export function NavPanelItem({ label, to, exact }: NavPanelItemProps) {
  const { pathname } = useLocation();
  const active = exact
    ? pathname === to
    : pathname === to || pathname.startsWith(`${to}/`);

  const slots = navPanel();

  return (
    <ButtonLink
      to={to}
      // Gold only once selected. Colouring every entry would spend the accent
      // on the list rather than on the one entry worth pointing at.
      //
      // Soft rather than ghost while selected: ghost-active lands on gold-3,
      // the same weight as the neutral hover, so selection would not read as
      // the stronger state. Soft-active lands a step above it.
      variant={active ? "soft" : "ghost"}
      color={active ? "gold" : "neutral"}
      size={2}
      active={active}
      className={slots.item()}
    >
      {label}
    </ButtonLink>
  );
}
