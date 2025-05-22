import { SeverityBadge } from "./SeverityBadge.tsx";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Badges/SeverityBadge",
    component: SeverityBadge,
    argTypes: {},
} satisfies Meta<typeof SeverityBadge>;

type Story = StoryObj<typeof SeverityBadge>;

export const Default: Story = {
    render: () => (
        <div className="space-y-2">
            <SeverityBadge severity={75} />
            <SeverityBadge severity={50} />
            <SeverityBadge severity={25} />
            <SeverityBadge severity={0} />
        </div>
    ),
};
