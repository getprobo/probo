import type { Meta, StoryObj } from "@storybook/react";

import { InfiniteScrollTrigger } from "./InfiniteScrollTrigger";

export default {
  title: "Atoms/InfiniteScrollTrigger",
  component: InfiniteScrollTrigger,
  argTypes: {},
} satisfies Meta<typeof InfiniteScrollTrigger>;

type Story = StoryObj<typeof InfiniteScrollTrigger>;

export const Default: Story = {
  args: {
    onView: () => {},
  },
};
