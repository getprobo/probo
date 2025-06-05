import { RiskBadge } from "./RiskBadge";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Badges/RiskBadge",
    component: RiskBadge,
    argTypes: {},
} satisfies Meta<typeof RiskBadge>;

type Story = StoryObj<typeof RiskBadge>;

export const Default: Story = {
    render: () => (
        <div className="space-y-2">
            <RiskBadge level={10} />
            <RiskBadge level={8} />
            <RiskBadge level={0} />
        </div>
    ),
};
