import { Badge } from "./Badge";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Badge",
    component: Badge,
    argTypes: {},
} satisfies Meta<typeof Badge>;

type Story = StoryObj<typeof Badge>;

export const Default: Story = {
    render: () => (
        <>
            {(
                [
                    "success",
                    "warning",
                    "danger",
                    "info",
                    "neutral",
                    "outline",
                ] as const
            ).map((variant) => (
                <Badge key={variant} variant={variant}>
                    {variant}
                </Badge>
            ))}
        </>
    ),
};
