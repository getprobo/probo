import type { Meta, StoryObj } from "@storybook/react";

import { Select } from "../../Atoms/Select/Select.tsx";

import { ImpactOptions } from "./ImpactOptions.tsx";

export default {
  title: "Molecules/Select/ImpactSelect",
  component: ImpactOptions,
  argTypes: {},
} satisfies Meta<typeof ImpactOptions>;

type Story = StoryObj<typeof ImpactOptions>;

export const Default: Story = {
  render() {
    return (
      <Select placeholder="Select impact">
        <ImpactOptions />
      </Select>
    );
  },
};
