import type { Meta, StoryObj } from "@storybook/react";

import { Input } from "./Input";

export default {
  title: "Atoms/Form/Input",
  component: Input,
  argTypes: {},
} satisfies Meta<typeof Input>;

type Story = StoryObj<typeof Input>;

export const Default: Story = {};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};

export const Title: Story = {
  args: {
    variant: "title",
    placeholder: "Policy title",
  },
};
