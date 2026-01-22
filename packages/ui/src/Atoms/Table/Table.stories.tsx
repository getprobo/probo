import type { Meta, StoryObj } from "@storybook/react";

import { Table, Tbody, Td, Th, Thead, Tr, TrButton } from "./Table";

export default {
  title: "Atoms/Table",
  component: Table,
  argTypes: {},
} satisfies Meta<typeof Table>;

type Story = StoryObj<typeof Table>;

export const Default: Story = {
  render: () => {
    return (
      <Table>
        <Thead>
          <Tr>
            <Th>Header 1</Th>
            <Th>Header 2</Th>
            <Th>Header 3</Th>
          </Tr>
        </Thead>
        <Tbody>
          <Tr>
            <Td>Row 1, Cell 1</Td>
            <Td>Row 1, Cell 2</Td>
            <Td>Row 1, Cell 3</Td>
          </Tr>
          <Tr>
            <Td>Row 2, Cell 1</Td>
            <Td>Row 2, Cell 2</Td>
            <Td>Row 2, Cell 3</Td>
          </Tr>
          <TrButton onClick={() => {}} colspan={3}>
            Add row
          </TrButton>
        </Tbody>
      </Table>
    );
  },
};
