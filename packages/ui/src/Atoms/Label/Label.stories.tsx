import type { Meta, StoryObj } from "@storybook/react";

import { Label } from "./Label";

export default {
  title: "Atoms/Form/Label",
  component: Label,
  argTypes: {},
} satisfies Meta<typeof Label>;

type Story = StoryObj<typeof Label>;

export const Default: Story = {};
