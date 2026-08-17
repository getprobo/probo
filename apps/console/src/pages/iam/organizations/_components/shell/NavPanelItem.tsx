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
  // Extra route prefixes that belong to this section but are not nested under
  // `to` in the route tree, such as sibling detail pages.
  alsoActiveFor?: readonly string[];
}

function withoutTrailingSlash(path: string) {
  return path.length > 1 && path.endsWith("/") ? path.slice(0, -1) : path;
}

export function NavPanelItem({ label, to, exact, alsoActiveFor }: NavPanelItemProps) {
  const { pathname } = useLocation();
  const path = withoutTrailingSlash(pathname);
  const active = [to, ...alsoActiveFor ?? []]
    .map(withoutTrailingSlash)
    .some(target => exact
      ? path === target
      : path === target || path.startsWith(`${target}/`));

  const slots = navPanel();

  return (
    <ButtonLink
      to={to}
      variant={active ? "soft" : "ghost"}
      color={active ? "gold" : "neutral"}
      size={2}
      active={active}
      aria-current={active ? "page" : undefined}
      className={slots.item()}
    >
      {label}
    </ButtonLink>
  );
}
