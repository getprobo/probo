import { MeasureImplementation } from "./MeasureImplementation";
import type { Meta, StoryObj } from "@storybook/react";
import { times } from "@probo/helpers";

export default {
    title: "Molecules/MeasureImplementation",
    component: MeasureImplementation,
    argTypes: {},
} satisfies Meta<typeof MeasureImplementation>;

type Story = StoryObj<typeof MeasureImplementation>;

export const Default: Story = {
    args: {
        measures: [
            ...times(15, () => ({
                state: "IMPLEMENTED",
            })),
            ...times(10, () => ({
                state: "IN_PROGRESS",
            })),
            ...times(5, () => ({
                state: "NOT_APPLICABLE",
            })),
            ...times(5, () => ({
                state: "NOT_STARTED",
            })),
        ] as any,
    },
};
