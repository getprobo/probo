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

import { Menu } from "@base-ui/react/menu";
import { CheckIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { Link } from "react-router";

import { navPanelSwitcher } from "./NavPanelSwitcher";

export interface NavPanelSwitcherListItemProps {
  to: string;
  name: string;
  detail?: string;
  leading?: ReactNode;
  selected: boolean;
}

/**
 * One entity in a product-panel switcher.
 *
 * Name and optional detail stack because a single truncated line cannot tell
 * two similar entities apart. The check marks the entity the URL is on.
 */
export function NavPanelSwitcherListItem({
  to,
  name,
  detail,
  leading,
  selected,
}: NavPanelSwitcherListItemProps) {
  const slots = navPanelSwitcher();

  return (
    <Menu.Item
      className={slots.item()}
      render={<Link to={to} />}
    >
      {leading}
      <span className={slots.itemBody()}>
        <Text size={2} weight="medium" color="neutral" highContrast className={slots.itemName()}>
          {name}
        </Text>
        {detail != null && detail !== "" && (
          <Text size={1} color="faint" className={slots.itemOrigin()}>
            {detail}
          </Text>
        )}
      </span>
      {selected && <CheckIcon className={slots.itemCheck()} />}
    </Menu.Item>
  );
}
