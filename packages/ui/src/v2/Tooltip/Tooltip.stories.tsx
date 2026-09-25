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

import { BellIcon, GearIcon, HouseIcon } from "@phosphor-icons/react";

import { Button } from "../Button/Button";
import { IconButton } from "../IconButton/IconButton";

import { Tooltip } from "./Tooltip";
import { TooltipPopup } from "./TooltipPopup";
import { TooltipProvider } from "./TooltipProvider";
import { TooltipTrigger } from "./TooltipTrigger";

export default {
  title: "v2/Tooltip",
  component: Tooltip,
};

export function Default() {
  return (
    <Tooltip>
      <TooltipTrigger render={<Button variant="soft" color="neutral">Hover me</Button>} />
      <TooltipPopup>Tooltip label</TooltipPopup>
    </Tooltip>
  );
}

export function Sides() {
  return (
    <div className="flex gap-3">
      {(["top", "right", "bottom", "left"] as const).map(side => (
        <Tooltip key={side}>
          <TooltipTrigger render={<Button variant="soft" color="neutral">{side}</Button>} />
          <TooltipPopup side={side}>{`Opens on the ${side}`}</TooltipPopup>
        </Tooltip>
      ))}
    </div>
  );
}

export function IconRail() {
  const items = [
    { label: "Home", icon: <HouseIcon /> },
    { label: "Notifications", icon: <BellIcon /> },
    { label: "Settings", icon: <GearIcon /> },
  ];

  return (
    <TooltipProvider>
      <div className="flex w-12 flex-col items-center gap-1 rounded-3 bg-sand-2 p-2">
        {items.map(item => (
          <Tooltip key={item.label}>
            <TooltipTrigger
              render={(
                <IconButton variant="ghost" color="neutral" aria-label={item.label}>
                  {item.icon}
                </IconButton>
              )}
            />
            <TooltipPopup side="right">{item.label}</TooltipPopup>
          </Tooltip>
        ))}
      </div>
    </TooltipProvider>
  );
}
