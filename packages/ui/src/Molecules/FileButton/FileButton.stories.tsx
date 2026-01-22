import type { Meta, StoryObj } from "@storybook/react";

import { FileButton } from "./FileButton";

export default {
  title: "Molecules/FileButton",
  component: FileButton,
  argTypes: {},
} satisfies Meta<typeof FileButton>;

type Story = StoryObj<typeof FileButton>;

export const Default: Story = {
  args: {
    children: "Upload",
  },
};
