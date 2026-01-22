import { Textarea } from "./Textarea";
import type { Meta, StoryObj } from "@storybook/react";

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
