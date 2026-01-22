import type { Meta, StoryObj } from "@storybook/react";

import { Skeleton } from "./Skeleton";

export default {
  title: "Atoms/Skeleton",
  component: Skeleton,
  argTypes: {},
} satisfies Meta<typeof Skeleton>;

type Story = StoryObj<typeof Skeleton>;

export const Default: Story = {
  args: {
    className: "w-50 h-4",
  },
};
