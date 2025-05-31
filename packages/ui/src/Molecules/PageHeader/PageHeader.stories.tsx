import { Button } from "../../Atoms/Button/Button.tsx";
import { PageHeader } from "./PageHeader";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Molecules/PageHeader",
    component: PageHeader,
    argTypes: {},
} satisfies Meta<typeof PageHeader>;

type Story = StoryObj<typeof PageHeader>;

export const Default: Story = {
    args: {
        title: "Page Header",
        description: "This is a page header",
        children: <Button variant="primary">Button</Button>,
    },
};
