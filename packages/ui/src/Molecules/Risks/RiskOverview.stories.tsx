import type { Meta, StoryObj } from "@storybook/react";

import { RiskOverview } from "./RiskOverview";

export default {
  title: "Molecules/Risks/RiskOverview",
  component: RiskOverview,
  argTypes: {},
} satisfies Meta<typeof RiskOverview>;

type Story = StoryObj<typeof RiskOverview>;

export const Default: Story = {
  args: {
    type: "inherent",
    risk: {
      inherentLikelihood: 5,
      inherentImpact: 5,
      residualLikelihood: 1,
      residualImpact: 1,
    },
  },
};
