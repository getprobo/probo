import type { Meta, StoryObj } from "@storybook/react";

import { Card } from "./Card";

export default {
  title: "Atoms/Card",
  component: Card,
  argTypes: {},
} satisfies Meta<typeof Card>;

type Story = StoryObj<typeof Card>;

export const Default: Story = {
  args: {
    padded: true,
    children: "lorem ipsum",
  },
};
