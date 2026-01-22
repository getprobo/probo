import type { Meta, StoryObj } from "@storybook/react";

import { Textarea } from "./Textarea";

export default {
  title: "Atoms/Form/Textarea",
  component: Textarea,
  argTypes: {},
} satisfies Meta<typeof Textarea>;

type Story = StoryObj<typeof Textarea>;

export const Default: Story = {};
export const Autogrow: Story = {
  args: {
    autogrow: true,
  },
};
