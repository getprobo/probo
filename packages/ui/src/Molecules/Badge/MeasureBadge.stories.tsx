import { MeasureBadge } from "./MeasureBadge";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Badges/MeasureBadge",
    component: MeasureBadge,
    argTypes: {},
} satisfies Meta<typeof MeasureBadge>;

type Story = StoryObj<typeof MeasureBadge>;

export const Default: Story = {};
