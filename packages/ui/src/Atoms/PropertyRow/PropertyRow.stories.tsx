import type { Meta, StoryObj } from "@storybook/react";

import { Select } from "../Select/Select";

import { PropertyRow } from "./PropertyRow";

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

export const WithError: Story = {
  args: {
    ...Default.args,
    error: "This is an error",
  },
};
