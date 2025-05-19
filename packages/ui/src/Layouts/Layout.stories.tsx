import { Layout } from "./Layout";
import type { Meta, StoryObj } from "@storybook/react";

const meta = {
    title: "Layouts/Base",
    component: Layout,
    parameters: {
        layout: "full",
    },
} satisfies Meta<typeof Layout>;

export default meta;
type Story = StoryObj<typeof Layout>;

export const Default: Story = {};
