import { FrameworkSelector } from "./FrameworkSelector";
import type { Meta, StoryObj } from "@storybook/react";

export default {
  title: "Molecules/FrameworkSelector",
  component: FrameworkSelector,
  argTypes: {},
} satisfies Meta<typeof FrameworkSelector>;

type Story = StoryObj<typeof FrameworkSelector>;

export const Default: Story = {};
