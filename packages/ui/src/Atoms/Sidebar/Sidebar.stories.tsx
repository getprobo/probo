import type { Meta, StoryObj } from "@storybook/react";

import { IconBank } from "../Icons/IconBank";

import { Sidebar } from "./Sidebar";
import { SidebarItem } from "./SidebarItem";

const meta: Meta<typeof Sidebar> = {
  component: Sidebar,
  subcomponents: { SidebarItem },
};
export default meta;

type Story = StoryObj<typeof Sidebar>;

export const Default: Story = {
  render: args => (
    <Sidebar {...args}>
      <SidebarItem label="Dashboard" icon={IconBank} />
    </Sidebar>
  ),
};
