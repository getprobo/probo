import { Breadcrumb } from "./Breadcrumb";
import type { Meta, StoryObj } from "@storybook/react";

export default {
  title: "Atoms/Breadcrumb",
  component: Breadcrumb,
  argTypes: {},
} satisfies Meta<typeof Breadcrumb>;

type Story = StoryObj<typeof Breadcrumb>;

export const Default: Story = {
  args: {
    items: [
      {
        label: "Home",
        to: "/",
      },
      {
        label: "Library",
        to: "/library",
      },
      {
        label: "Data",
        to: "/data",
      },
    ],
  },
};
