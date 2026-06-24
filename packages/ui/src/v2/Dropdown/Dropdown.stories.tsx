// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useState } from "react";

import { Button } from "../Button/Button";

import { Dropdown } from "./Dropdown";
import { DropdownCheckboxItem } from "./DropdownCheckboxItem";
import { DropdownGroup } from "./DropdownGroup";
import { DropdownGroupLabel } from "./DropdownGroupLabel";
import { DropdownItem } from "./DropdownItem";
import { DropdownPopup } from "./DropdownPopup";
import { DropdownRadioGroup } from "./DropdownRadioGroup";
import { DropdownRadioItem } from "./DropdownRadioItem";
import { DropdownSeparator } from "./DropdownSeparator";
import { DropdownSubmenu } from "./DropdownSubmenu";
import { DropdownSubmenuTrigger } from "./DropdownSubmenuTrigger";
import { DropdownTrigger } from "./DropdownTrigger";

export default {
  title: "v2/Dropdown",
  component: Dropdown,
};

export function Default() {
  return (
    <Dropdown>
      <DropdownTrigger render={<Button variant="soft" color="neutral">Open menu</Button>} />
      <DropdownPopup>
        <DropdownItem shortcut="⌘ E">Edit</DropdownItem>
        <DropdownItem shortcut="⌘ D">Duplicate</DropdownItem>
        <DropdownSeparator />
        <DropdownSubmenu>
          <DropdownSubmenuTrigger>More</DropdownSubmenuTrigger>
          <DropdownPopup side="right" sideOffset={4}>
            <DropdownItem>Move</DropdownItem>
            <DropdownItem>Archive</DropdownItem>
          </DropdownPopup>
        </DropdownSubmenu>
        <DropdownSeparator />
        <DropdownItem color="error" shortcut="⌘ ⌫">Delete</DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}

export function Soft() {
  return (
    <Dropdown>
      <DropdownTrigger render={<Button variant="soft" color="neutral">Soft highlight</Button>} />
      <DropdownPopup variant="soft">
        <DropdownItem shortcut="⌘ E">Edit</DropdownItem>
        <DropdownItem shortcut="⌘ D">Duplicate</DropdownItem>
        <DropdownSeparator />
        <DropdownItem color="error">Delete</DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}

export function Size1() {
  return (
    <Dropdown>
      <DropdownTrigger render={<Button size={1} variant="soft" color="neutral">Small</Button>} />
      <DropdownPopup size={1}>
        <DropdownItem shortcut="⌘ E">Edit</DropdownItem>
        <DropdownItem shortcut="⌘ D">Duplicate</DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}

export function WithGroupLabel() {
  return (
    <Dropdown>
      <DropdownTrigger render={<Button variant="soft" color="neutral">Account</Button>} />
      <DropdownPopup>
        <DropdownGroup>
          <DropdownGroupLabel>Account</DropdownGroupLabel>
          <DropdownItem>Profile</DropdownItem>
          <DropdownItem>Settings</DropdownItem>
        </DropdownGroup>
      </DropdownPopup>
    </Dropdown>
  );
}

export function CheckboxAndRadio() {
  const [showGrid, setShowGrid] = useState(true);
  const [view, setView] = useState("comfortable");

  return (
    <Dropdown>
      <DropdownTrigger render={<Button variant="soft" color="neutral">View</Button>} />
      <DropdownPopup>
        <DropdownCheckboxItem checked={showGrid} onCheckedChange={setShowGrid}>
          Show grid
        </DropdownCheckboxItem>
        <DropdownSeparator />
        <DropdownRadioGroup value={view} onValueChange={setView}>
          <DropdownRadioItem value="comfortable">Comfortable</DropdownRadioItem>
          <DropdownRadioItem value="compact">Compact</DropdownRadioItem>
        </DropdownRadioGroup>
      </DropdownPopup>
    </Dropdown>
  );
}
