import { Table } from "./Table";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Table",
    component: Table,
    argTypes: {},
} satisfies Meta<typeof Table>;

type Story = StoryObj<typeof Table>;

export const Default: Story = {};
