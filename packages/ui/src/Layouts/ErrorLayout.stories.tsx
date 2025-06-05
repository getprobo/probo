import { ErrorLayout } from "./ErrorLayout";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Layouts/ErrorLayout",
    component: ErrorLayout,
    argTypes: {},
} satisfies Meta<typeof ErrorLayout>;

type Story = StoryObj<typeof ErrorLayout>;

export const Default: Story = {
    args: {
        title: "Something went wrong",
        description: "An unexpected error occurred. Please try again later.",
    },
};
