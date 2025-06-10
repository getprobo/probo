import { TaskStateIcon } from "./TaskStateIcon";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/TaskStateIcon",
    component: TaskStateIcon,
    argTypes: {},
} satisfies Meta<typeof TaskStateIcon>;

type Story = StoryObj<typeof TaskStateIcon>;

export const Default: Story = {
    args: {
        state: "TODO",
    },
};

export const Done: Story = {
    args: {
        state: "DONE",
    },
};
