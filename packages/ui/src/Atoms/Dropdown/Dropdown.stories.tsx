import type { Meta, StoryObj } from "@storybook/react";

import { Button } from "../Button/Button";
import { IconCircleQuestionmark1 } from "../Icons";

import { Dropdown, DropdownItem, DropdownSeparator } from "./Dropdown";

export default {
  title: "Atoms/Dropdown",
  component: Dropdown,
  argTypes: {},
  subcomponents: {
    DropdownItem,
    DropdownSeparator,
  },
} satisfies Meta<typeof Dropdown>;

type Story = StoryObj<typeof Dropdown>;

export const Default: Story = {
  render: () => {
    return (
      <Dropdown toggle={<Button variant="tertiary">Dropdown</Button>}>
        <DropdownItem icon={IconCircleQuestionmark1}>
          Item 1
        </DropdownItem>
        <DropdownSeparator />
        <DropdownItem icon={IconCircleQuestionmark1}>
          Item 2
        </DropdownItem>
        <DropdownItem icon={IconCircleQuestionmark1}>
          Item 3
        </DropdownItem>
      </Dropdown>
    );
  },
};
