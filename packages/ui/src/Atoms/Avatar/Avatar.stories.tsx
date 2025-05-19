import { Avatar } from "./Avatar";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Avatar",
    component: Avatar,
    argTypes: {},
} satisfies Meta<typeof Avatar>;

type Story = StoryObj<typeof Avatar>;

export const Default: Story = {
    args: {
        fullName: "John Doe",
    },
};
