import type { Meta, StoryObj } from "@storybook/react";

import { Select } from "../../Atoms/Select/Select";

import { SentitivityOptions } from "./SentitivityOptions";

export default {
  title: "Molecules/Select/SentitivitySelect",
  component: SentitivityOptions,
  argTypes: {},
} satisfies Meta<typeof SentitivityOptions>;

type Story = StoryObj<typeof SentitivityOptions>;

export const Default: Story = {
  render() {
    return (
      <Select placeholder="Select sensitivity">
        <SentitivityOptions />
      </Select>
    );
  },
};
