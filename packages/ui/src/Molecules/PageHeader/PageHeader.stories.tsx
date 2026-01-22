import type { Meta, StoryObj } from "@storybook/react";

import { Button } from "../../Atoms/Button/Button";

import { PageHeader } from "./PageHeader";

export default {
  title: "Molecules/PageHeader",
  component: PageHeader,
  argTypes: {},
} satisfies Meta<typeof PageHeader>;

type Story = StoryObj<typeof PageHeader>;

export const Default: Story = {
  args: {
    title: "Page Header",
    description: "This is a page header",
    children: <Button variant="primary">Button</Button>,
  },
};
