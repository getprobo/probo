import type { Meta, StoryObj } from "@storybook/react";

import { PriorityLevel } from "./PriorityLevel";

export default {
  title: "Atoms/PriorityLevel",
  component: PriorityLevel,
  argTypes: {},
} satisfies Meta<typeof PriorityLevel>;

type Story = StoryObj<typeof PriorityLevel>;

export const Default: Story = {
  args: {
    level: 1,
  },
};
