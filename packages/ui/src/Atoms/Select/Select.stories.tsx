import { Select, Option } from "./Select";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Form/Select",
    component: Select,
    argTypes: {},
} satisfies Meta<typeof Select>;

type Story = StoryObj<typeof Select>;

export const Default: Story = {
    args: {
        placeholder: "Select an option",
        children: (
            <>
                <Option value="apple">Apple</Option>
                <Option value="banana">Banana</Option>
                <Option value="blueberry">Blueberry</Option>
                <Option value="grapes">Grapes</Option>
                <Option value="pineapple">Pineapple</Option>
            </>
        ),
    },
};

export const NoChildren: Story = {
    args: {
        placeholder: "Select an option",
    },
};
