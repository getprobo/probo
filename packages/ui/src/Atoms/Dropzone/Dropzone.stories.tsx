import { Dropzone } from "./Dropzone";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Dropzone",
    component: Dropzone,
    argTypes: {},
} satisfies Meta<typeof Dropzone>;

type Story = StoryObj<typeof Dropzone>;

export const Default: Story = {};
