import type { Meta, StoryObj } from "@storybook/react";

import { CenteredLayout } from "./CenteredLayout";

export default {
  title: "Layouts/CenteredLayout",
  component: CenteredLayout,
  argTypes: {},
} satisfies Meta<typeof CenteredLayout>;

type Story = StoryObj<typeof CenteredLayout>;

export const Default: Story = {
  args: {
    children: "lorem ipsum",
  },
};
