import { FrameworkLogo } from "./FrameworkLogo";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Badges/FrameworkLogo",
    component: FrameworkLogo,
    argTypes: {},
} satisfies Meta<typeof FrameworkLogo>;

type Story = StoryObj<typeof FrameworkLogo>;

export const Default: Story = {
    args: {
        name: "SOC 2",
        className: "size-12",
    },
};
