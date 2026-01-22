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
