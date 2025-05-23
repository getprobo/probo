import { Select } from "../Select/Select";
import { PropertyRow } from "./PropertyRow";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/PropertyRow",
    component: PropertyRow,
    argTypes: {},
} satisfies Meta<typeof PropertyRow>;

type Story = StoryObj<typeof PropertyRow>;

export const Default: Story = {
    args: {
        id: "test",
        label: "Test",
        children: <Select variant="editor" placeholder="Select an otion" />,
    },
};

export const Disabled: Story = {
    args: {
        ...Default.args,
        disabled: true,
    },
};

export const WithError: Story = {
    args: {
        ...Default.args,
        error: "This is an error",
    },
};
