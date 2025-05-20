import { Skeleton } from "./Skeleton";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Skeleton",
    component: Skeleton,
    argTypes: {},
} satisfies Meta<typeof Skeleton>;

type Story = StoryObj<typeof Skeleton>;

export const Default: Story = {
    args: {
        className: "w-50 h-4",
    },
};
