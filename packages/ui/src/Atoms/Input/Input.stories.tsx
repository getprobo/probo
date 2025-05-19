import { Input } from "./Input";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Form/Input",
    component: Input,
    argTypes: {},
} satisfies Meta<typeof Input>;

type Story = StoryObj<typeof Input>;

export const Default: Story = {
    args: {
        label: "Label",
        name: "id",
        help: "e.g. This is a hint",
    },
};
