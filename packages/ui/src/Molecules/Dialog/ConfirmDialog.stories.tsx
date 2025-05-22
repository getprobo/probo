import { Button } from "../../Atoms/Button/Button";
import { ConfirmDialog } from "./ConfirmDialog";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/ConfirmDialog",
    component: ConfirmDialog,
    argTypes: {},
} satisfies Meta<typeof ConfirmDialog>;

type Story = StoryObj<typeof ConfirmDialog>;

export const Default: Story = {
    args: {
        message:
            'This will permanently delete the risk "Demo". This action cannot be undone.',
    },
    render() {
        return (
            <ConfirmDialog
                message="Are you sure you want to delete this risk?"
                onConfirm={() => {}}
            >
                <Button variant="danger">Delete this</Button>
            </ConfirmDialog>
        );
    },
};
