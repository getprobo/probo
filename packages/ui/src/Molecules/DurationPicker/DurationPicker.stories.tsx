import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";

import { DurationPicker } from "./DurationPicker";

export default {
  title: "Molecules/DurationPicker",
  component: DurationPicker,
  argTypes: {},
} satisfies Meta<typeof DurationPicker>;

type Story = StoryObj<typeof DurationPicker>;

export const Default: Story = {
  render: function Render() {
    const [value, setValue] = useState<string | null>("PT1H");
    return <DurationPicker value={value} onValueChange={setValue} />;
  },
};
