import { DocumentVersionBadge } from "./DocumentVersionBadge";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Badges/DocumentVersionBadge",
    component: DocumentVersionBadge,
    argTypes: {},
} satisfies Meta<typeof DocumentVersionBadge>;

type Story = StoryObj<typeof DocumentVersionBadge>;

export const Default: Story = {
    render() {
        return (
            <div className="flex gap-2">
                <DocumentVersionBadge state="DRAFT" />
                <DocumentVersionBadge state="PUBLISHED" />
            </div>
        );
    },
};
