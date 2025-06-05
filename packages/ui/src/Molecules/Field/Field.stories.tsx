import { Field } from "./Field.tsx";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/Field",
    component: Field,
    argTypes: {},
} satisfies Meta<typeof Field>;

type Story = StoryObj<typeof Field>;

export const Default: Story = {
    args: {
        label: "Label",
        name: "id",
        help: "e.g. This is a hint",
    },
};

export const Invalid: Story = {
    args: {
        ...Default.args,
        error: "This is an error",
    },
};
