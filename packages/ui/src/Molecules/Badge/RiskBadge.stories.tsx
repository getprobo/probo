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
            <RiskBadge score={10} />
            <RiskBadge score={8} />
            <RiskBadge score={0} />
        </div>
    ),
};
