import { Checkbox } from "./Checkbox";
import type { Meta, StoryObj } from "@storybook/react";

export default {
  title: "Atoms/Checkbox",
  component: Checkbox,
  argTypes: {},
} satisfies Meta<typeof Checkbox>;

type Story = StoryObj<typeof Checkbox>;

export const Default: Story = {};
