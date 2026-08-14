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

// Product-panel picker. Two-line items (name + detail) do not fit the kit
// DropdownItem's fixed height, so the list item owns these slots and renders
// a Menu.Item directly.
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
    itemCheck: "size-4 shrink-0",
    trigger: [
      "flex h-9 w-full min-w-0 items-center gap-2 rounded-2 px-3",
      "cursor-pointer outline-none transition-colors select-none",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    value: "flex min-w-0 flex-1 items-center gap-2 text-left",
    valueCaret: "size-4 shrink-0",
    // Selected siblings use a gold fill that otherwise kisses this outlined
    // trigger. Collapse when this is the first or last child so a sole
    // switcher does not drop away from the product title or leave a hole
    // at the foot of the column.
    root: "mt-2 mb-2 first:mt-0 last:mb-0",
    // Switcher plus a trailing action (open-in-new-tab) on one row.
    row: "mt-2 mb-2 flex items-center gap-1 first:mt-0 last:mb-0",
    rowTrigger: "min-w-0 flex-1",
    // No [&_svg] size: Phosphor `size` on the icon is the source of truth.
    openLink: [
      "flex size-7 shrink-0 items-center justify-center rounded-2",
      "text-sand-11 outline-none transition-colors select-none",
      "hover:bg-sand-3 hover:text-sand-12",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    // h-5 matches text-2 line-height so the trigger does not jump when the
    // selected name replaces this pulse.
    valueSkeletonName: "block h-5 w-28 animate-pulse rounded-1 bg-sand-3",
  },
  variants: {
    // Ring rather than border: Button's base `border-transparent` and the
    // outline `border-sand-7` both emit under lite tv, and the transparent
    // one wins. A ring does not collide with that.
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
  onOpenChange: (open: boolean) => void;
  value: ReactNode;
  children?: ReactNode;
}

/**
 * The dropdown chrome for a product-panel entity switcher.
 *
 * Feature switchers own the Relay queries and pass the trigger value and
 * menu. Gold only while creating: once an entity is selected, the panel
 * links carry the accent and the trigger stays an outline picker.
 */
export function NavPanelSwitcher({ active, onOpenChange, value, children }: NavPanelSwitcherProps) {
  const slots = navPanelSwitcher({ outlined: !active, active });

  return (
    <div className={slots.root()}>
      <Dropdown onOpenChange={onOpenChange}>
        <DropdownTrigger
          render={(
            <button type="button" className={slots.trigger()}>
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

/**
 * The trigger or list-item name. Same type as the closed trigger so a
 * fallback string does not inherit the page font and flash larger.
 */
export function NavPanelSwitcherValue({ children }: NavPanelSwitcherValueProps) {
  const slots = navPanelSwitcher();

  return (
    <Text size={2} weight="medium" color="neutral" highContrast className={slots.itemName()}>
      {children}
    </Text>
  );
}

/**
 * A pulse bar the same height as the trigger name, used while the selected
 * entity query suspends.
 */
export function NavPanelSwitcherValueSkeleton() {
  const slots = navPanelSwitcher();

  return <span className={slots.valueSkeletonName()} aria-hidden />;
}
