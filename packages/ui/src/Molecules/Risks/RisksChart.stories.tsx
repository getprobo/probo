import { randomInt, times } from "@probo/helpers";
import type { Meta, StoryObj } from "@storybook/react";

import { RisksChart } from "./RisksChart";

export default {
  title: "Molecules/Risks/RisksChart",
  component: RisksChart,
  argTypes: {},
} satisfies Meta<typeof RisksChart>;

type Story = StoryObj<typeof RisksChart>;

export const Default: Story = {
  args: {
    type: "inherent",
    organizationId: "1",
    risks: times(20, i => ({
      id: i.toString(),
      name: `Risk ${i}`,
      inherentLikelihood: randomInt(1, 5),
      inherentImpact: randomInt(1, 5),
      residualLikelihood: randomInt(1, 5),
      residualImpact: randomInt(1, 5),
    })),
  },
  render: args => (
    <div style={{ maxWidth: 630 }}>
      <RisksChart {...args} />
    </div>
  ),
};
