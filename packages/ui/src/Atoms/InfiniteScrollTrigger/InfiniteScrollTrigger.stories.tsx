import { InfiniteScrollTrigger } from "./InfiniteScrollTrigger";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/InfiniteScrollTrigger",
    component: InfiniteScrollTrigger,
    argTypes: {},
} satisfies Meta<typeof InfiniteScrollTrigger>;

type Story = StoryObj<typeof InfiniteScrollTrigger>;

export const Default: Story = {
    args: {
        onView: () => {},
    },
};
