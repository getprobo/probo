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
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Link } from "react-router";

import { navRail } from "./variants";

export interface NavRailItemProps {
  icon: Icon;
  label: string;
  to: string;
  active: boolean;
}

/**
 * One product in the rail: an icon that names itself on hover.
 *
 * The label is the tooltip *and* the accessible name, so the icon never has to
 * carry meaning on its own.
 */
export function NavRailItem({ icon: IconComponent, label, to, active }: NavRailItemProps) {
  const slots = navRail({ active });

  return (
    <Tooltip>
      <TooltipTrigger
        render={(
          <Link to={to} aria-label={label} className={slots.item()}>
            <IconComponent size={20} aria-hidden />
          </Link>
        )}
      />
      <TooltipPopup side="right">{label}</TooltipPopup>
    </Tooltip>
  );
}
