import { PriorityLevel } from "./PriorityLevel.tsx";
import type { Meta, StoryObj } from "@storybook/react";

export default {
  title: "Atoms/PriorityLevel",
  component: PriorityLevel,
  argTypes: {},
} satisfies Meta<typeof PriorityLevel>;

type Story = StoryObj<typeof PriorityLevel>;

export const Default: Story = {
  args: {
    level: 1,
  },
};
