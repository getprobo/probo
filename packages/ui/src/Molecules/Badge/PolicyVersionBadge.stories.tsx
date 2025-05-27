import { PolicyVersionBadge } from "./PolicyVersionBadge";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Badges/PolicyVersionBadge",
    component: PolicyVersionBadge,
    argTypes: {},
} satisfies Meta<typeof PolicyVersionBadge>;

type Story = StoryObj<typeof PolicyVersionBadge>;

export const Default: Story = {
    render() {
        return (
            <div className="flex gap-2">
                <PolicyVersionBadge state="DRAFT" />
                <PolicyVersionBadge state="PUBLISHED" />
            </div>
        );
    },
};
