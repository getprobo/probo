import { Label } from "./Label";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Form/Label",
    component: Label,
    argTypes: {},
} satisfies Meta<typeof Label>;

type Story = StoryObj<typeof Label>;

export const Default: Story = {};
