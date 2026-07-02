// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import type { Meta, StoryObj } from "@storybook/react";

import { DropdownSeparator } from "../../Atoms/Dropdown/Dropdown";
import {
  IconArrowBoxLeft,
  IconCircleQuestionmark,
  IconSettingsGear2,
} from "../../Atoms/Icons";

import { UserDropdown, UserDropdownItem } from "./UserDropdown";

export default {
  title: "Molecules/UserDropdown",
  component: UserDropdown,
  argTypes: {},
  subcomponents: {
    UserDropdownItem,
  },
} satisfies Meta<typeof UserDropdown>;

type Story = StoryObj<typeof UserDropdown>;

export const Default: Story = {
  args: {
    fullName: "John Doe",
    email: "john.doe@example.com",
    children: (
      <>
        <UserDropdownItem
          to="/settings"
          icon={IconSettingsGear2}
          label="Settings"
        />
        <UserDropdownItem
          to="/profile"
          icon={IconCircleQuestionmark}
          label="Profile"
        />
        <DropdownSeparator />
        <UserDropdownItem
          variant="danger"
          to="/logout"
          icon={IconArrowBoxLeft}
          label="Logout"
        />
      </>
    ),
  },
};
