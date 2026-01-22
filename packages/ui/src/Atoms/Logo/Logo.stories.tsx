import type { Meta, StoryObj } from "@storybook/react";

import { Logo } from "./Logo";

const meta = {
  title: "Atoms/Logo",
  component: Logo,
} satisfies Meta<typeof Logo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
