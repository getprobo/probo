import type { Meta, StoryObj } from "@storybook/react";

import { Option, Select } from "./Select";

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

export const Dashed: Story = {
  args: {
    ...Default.args,
    variant: "dashed",
  },
};

export const Invalid: Story = {
  args: {
    ...Default.args,
    invalid: true,
  },
};

export const Loading: Story = {
  args: {
    ...Default.args,
    loading: true,
  },
};

export const Disabled: Story = {
  args: {
    ...Default.args,
    disabled: true,
  },
};
