import type { Meta, StoryObj } from "@storybook/react";

import { MeasureBadge } from "./MeasureBadge";

export default {
  title: "Molecules/Badges/MeasureBadge",
  component: MeasureBadge,
  argTypes: {},
} satisfies Meta<typeof MeasureBadge>;

type Story = StoryObj<typeof MeasureBadge>;

export const Default: Story = {};
