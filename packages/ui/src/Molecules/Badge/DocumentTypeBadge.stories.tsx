import type { Meta, StoryObj } from "@storybook/react";

import { DocumentTypeBadge } from "./DocumentTypeBadge";

export default {
  title: "Molecules/Badges/DocumentTypeBadge",
  component: DocumentTypeBadge,
  argTypes: {},
} satisfies Meta<typeof DocumentTypeBadge>;

type Story = StoryObj<typeof DocumentTypeBadge>;

export const Default: Story = {
  args: {
    type: "OTHER",
  },
};
