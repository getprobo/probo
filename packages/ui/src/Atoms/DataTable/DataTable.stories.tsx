import type { Meta, StoryObj } from "@storybook/react";
import { DataTable, CellHead, Cell } from "./DataTable.tsx";

export default {
    title: "Atoms/DataTable",
    component: DataTable,
    argTypes: {},
} satisfies Meta<typeof DataTable>;

type Story = StoryObj<typeof DataTable>;

export const Default: Story = {
    render: () => {
        return (
            <DataTable columns={3}>
                <CellHead>Header 1</CellHead>
                <CellHead>Header 2</CellHead>
                <CellHead>Header 3</CellHead>
                <Cell>Row 1, Cell 1</Cell>
                <Cell>Row 1, Cell 2</Cell>
                <Cell>Row 1, Cell 3</Cell>
                <Cell>Row 2, Cell 1</Cell>
                <Cell>Row 2, Cell 2</Cell>
                <Cell>Row 2, Cell 3</Cell>
            </DataTable>
        );
    },
};
