import type { Meta, StoryObj } from "@storybook/react";

import { SeverityBadge } from "./SeverityBadge";

export default {
  title: "Molecules/Badges/SeverityBadge",
  component: SeverityBadge,
  argTypes: {},
} satisfies Meta<typeof SeverityBadge>;

type Story = StoryObj<typeof SeverityBadge>;

export const Default: Story = {
  render: () => (
    <div className="space-y-2">
      <SeverityBadge score={75} />
      <SeverityBadge score={50} />
      <SeverityBadge score={25} />
      <SeverityBadge score={0} />
    </div>
  ),
};
