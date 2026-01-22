import type { Meta, StoryObj } from "@storybook/react";

import { ControlItem } from "./ControlItem";

export default {
  title: "Atoms/ControlItem",
  component: ControlItem,
  argTypes: {},
} satisfies Meta<typeof ControlItem>;

type Story = StoryObj<typeof ControlItem>;

export const Default: Story = {
  args: {
    id: "CC1.1",
    description:
            "The entity obtains privacy commitments from vendors and other third parties who have access to personal information to meet the entity’s objectives related to privacy. The entity assesses those parties’ compliance on a periodic and as-needed basis and takes corrective action, if necessary.",
  },
  render: args => (
    <div className="p-4 space-y-2" style={{ width: "240px" }}>
      <ControlItem {...args} />
      <ControlItem {...args} active />
      <ControlItem {...args} />
      <ControlItem {...args} />
      <ControlItem {...args} />
    </div>
  ),
};
