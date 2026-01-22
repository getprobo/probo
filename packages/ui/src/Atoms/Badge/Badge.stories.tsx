import type { Meta, StoryObj } from "@storybook/react";

import { Badge } from "./Badge";

export default {
  title: "Atoms/Badge",
  component: Badge,
  argTypes: {},
} satisfies Meta<typeof Badge>;

type Story = StoryObj<typeof Badge>;

const sizes = ["sm", "md"] as const;
const variants = [
  "success",
  "warning",
  "danger",
  "info",
  "neutral",
  "outline",
  "highlight",
] as const;

export const Default: Story = {
  render: () => (
    <div className="space-y-4">
      {sizes.map(size => (
        <div key={size} className="space-y-1">
          <div className="text-xs text-txt-secondary">
            Size :
            {" "}
            {size}
          </div>
          <div className="flex gap-2">
            {variants.map(variant => (
              <Badge key={variant} variant={variant} size={size}>
                {variant}
              </Badge>
            ))}
          </div>
        </div>
      ))}
    </div>
  ),
};
