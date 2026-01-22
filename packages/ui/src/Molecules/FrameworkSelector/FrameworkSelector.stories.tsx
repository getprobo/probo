import type { Meta, StoryObj } from "@storybook/react";

import { FrameworkSelector } from "./FrameworkSelector";

export default {
  title: "Molecules/FrameworkSelector",
  component: FrameworkSelector,
  argTypes: {},
} satisfies Meta<typeof FrameworkSelector>;

type Story = StoryObj<typeof FrameworkSelector>;

export const Default: Story = {};
