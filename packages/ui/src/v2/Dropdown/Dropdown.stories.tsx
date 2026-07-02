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
