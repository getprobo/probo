import { AuthLayout } from "./AuthLayout";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Layouts/AuthLayout",
    component: AuthLayout,
    argTypes: {},
    parameters: {
        layout: "full",
    },
} satisfies Meta<typeof AuthLayout>;

type Story = StoryObj<typeof AuthLayout>;

export const Default: Story = {
    args: {
        children: "Text goes here",
    },
};
