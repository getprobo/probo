import type { Meta, StoryObj } from "@storybook/react";

import { Avatar } from "./Avatar";

export default {
  title: "Atoms/Avatar",
  component: Avatar,
  argTypes: {},
} satisfies Meta<typeof Avatar>;

type Story = StoryObj<typeof Avatar>;

export const Default: Story = {
  args: {
    name: "John Doe",
  },
};
