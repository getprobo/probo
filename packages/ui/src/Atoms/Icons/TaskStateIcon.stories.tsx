import type { Meta, StoryObj } from "@storybook/react";

import { TaskStateIcon } from "./TaskStateIcon";

export default {
  title: "Atoms/TaskStateIcon",
  component: TaskStateIcon,
  argTypes: {},
} satisfies Meta<typeof TaskStateIcon>;

type Story = StoryObj<typeof TaskStateIcon>;

export const Default: Story = {
  args: {
    state: "TODO",
  },
};

export const Done: Story = {
  args: {
    state: "DONE",
  },
};
