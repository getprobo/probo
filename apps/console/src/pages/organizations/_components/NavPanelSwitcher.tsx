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

import { CaretDownIcon } from "@phosphor-icons/react";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { tv } from "tailwind-variants/lite";

export const navPanelSwitcher = tv({
  slots: {
    popup: "w-72 p-0",
    list: "max-h-96 overflow-y-auto px-1 py-1",
    empty: "block px-3 py-2",
    item: [
      "group flex w-full cursor-pointer items-center gap-2 rounded-2 px-3 py-2 outline-none select-none",
      "text-sand-12 data-disabled:pointer-events-none data-disabled:opacity-50",
      "data-highlighted:bg-gold-9 data-highlighted:text-white",
    ],
    itemBody: "flex min-w-0 flex-1 flex-col gap-0.5",
    itemName: "truncate group-data-highlighted:text-white",
    itemOrigin: "truncate group-data-highlighted:text-white",
    itemCheck: "size-4 shrink-0 text-sand-12 group-data-highlighted:text-white",
    more: "w-full justify-start",
    trigger: [
      "flex h-9 w-full min-w-0 items-center gap-2 rounded-2 px-3",
      "cursor-pointer outline-none transition-colors select-none",
      "text-sand-12",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    triggerLabel: "sr-only",
    value: "flex min-w-0 flex-1 items-center gap-2 text-left",
    valueCaret: "size-4 shrink-0 text-current",
    root: "mt-2 mb-2 first:mt-0 last:mb-0",
    row: "mt-2 mb-2 flex items-center gap-1 first:mt-0 last:mb-0",
    rowTrigger: "min-w-0 flex-1",
    openLink: [
      "flex size-7 shrink-0 items-center justify-center rounded-2",
      "text-sand-11 outline-none transition-colors select-none",
      "hover:bg-sand-3 hover:text-sand-12",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    valueSkeletonName: "block h-5 w-28 animate-pulse rounded-1 bg-sand-3",
  },
  variants: {
    outlined: {
      true: { trigger: "ring-1 ring-inset ring-sand-7 hover:bg-sand-3" },
    },
    active: {
      true: { trigger: "bg-gold-4 text-gold-12" },
    },
  },
});

export interface NavPanelSwitcherProps {
  active: boolean;
  // Names the trigger for assistive tech; the selected value alone is a
  // skeleton while it loads, which would leave the button unnamed.
  label: string;
  onOpenChange: (open: boolean) => void;
  value: ReactNode;
  children?: ReactNode;
}

export function NavPanelSwitcher({
  active,
  label,
  onOpenChange,
  value,
  children,
}: NavPanelSwitcherProps) {
  const slots = navPanelSwitcher({ outlined: !active, active });

  return (
    <div className={slots.root()}>
      <Dropdown onOpenChange={onOpenChange}>
        <DropdownTrigger
          render={(
            <button type="button" className={slots.trigger()}>
              <span className={slots.triggerLabel()}>{label}</span>
              <span className={slots.value()}>{value}</span>
              <CaretDownIcon className={slots.valueCaret()} />
            </button>
          )}
        />
        <DropdownPopup side="bottom" align="start" className={slots.popup()}>
          {children}
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}

export interface NavPanelSwitcherValueProps {
  children: ReactNode;
}

export function NavPanelSwitcherValue({ children }: NavPanelSwitcherValueProps) {
  const slots = navPanelSwitcher();

  return (
    <Text size={2} weight="medium" color="current" className={slots.itemName()}>
      {children}
    </Text>
  );
}

export function NavPanelSwitcherValueSkeleton() {
  const slots = navPanelSwitcher();

  return <span className={slots.valueSkeletonName()} aria-hidden />;
}
