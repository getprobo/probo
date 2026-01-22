import type { Meta, StoryObj } from "@storybook/react";

import { Checkbox } from "./Checkbox";

export default {
  title: "Atoms/Checkbox",
  component: Checkbox,
  argTypes: {},
} satisfies Meta<typeof Checkbox>;

type Story = StoryObj<typeof Checkbox>;

export const Default: Story = {};
