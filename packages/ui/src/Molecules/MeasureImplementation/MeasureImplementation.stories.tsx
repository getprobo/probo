import { times } from "@probo/helpers";
import type { Meta, StoryObj } from "@storybook/react";

import { MeasureImplementation } from "./MeasureImplementation";

export default {
  title: "Molecules/MeasureImplementation",
  component: MeasureImplementation,
  argTypes: {},
} satisfies Meta<typeof MeasureImplementation>;

type Story = StoryObj<typeof MeasureImplementation>;

type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";

export const Default: Story = {
  args: {
    measures: [
      ...times(15, () => ({
        state: "IMPLEMENTED" as MeasureState,
      })),
      ...times(10, () => ({
        state: "IN_PROGRESS" as MeasureState,
      })),
      ...times(5, () => ({
        state: "NOT_APPLICABLE" as MeasureState,
      })),
      ...times(5, () => ({
        state: "NOT_STARTED" as MeasureState,
      })),
    ],
  },
};
