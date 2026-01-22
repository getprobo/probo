import type { Meta, StoryObj } from "@storybook/react";

import { IconPlusLarge } from "../Icons";

import { Button } from "./Button";

export default {
  title: "Atoms/Button",
  component: Button,
  argTypes: {},
} satisfies Meta<typeof Button>;

type Story = StoryObj<typeof Button>;

const variants = [
  "primary",
  "secondary",
  "tertiary",
  "quaternary",
  "danger",
] as const;

export const Default: Story = {
  render() {
    return (
      <div className="space-y-4">
        {variants.map(variant => (
          <div className="flex gap-4" key={variant}>
            <Button
              key={variant}
              variant={variant}
              icon={IconPlusLarge}
              iconAfter={IconPlusLarge}
            >
              Button
            </Button>
            <Button
              key={variant}
              disabled
              variant={variant}
              icon={IconPlusLarge}
              iconAfter={IconPlusLarge}
            >
              Button
            </Button>
            <Button
              key={variant}
              variant={variant}
              icon={IconPlusLarge}
            />
          </div>
        ))}
      </div>
    );
  },
};
