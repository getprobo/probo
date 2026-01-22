import type { Meta, StoryObj } from "@storybook/react";

import { Spinner } from "./Spinner";

export default {
  title: "Atoms/Spinner",
  component: Spinner,
  argTypes: {},
} satisfies Meta<typeof Spinner>;

type Story = StoryObj<typeof Spinner>;

export const Default: Story = {};
