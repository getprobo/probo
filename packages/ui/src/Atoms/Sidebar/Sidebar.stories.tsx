import type { Meta, StoryObj } from "@storybook/react";

import { Sidebar } from "./Sidebar";
import { SidebarItem } from "./SidebarItem";
import { IconBank } from "../Icons/IconBank.tsx";

const meta: Meta<typeof Sidebar> = {
    component: Sidebar,
    subcomponents: { SidebarItem },
};
export default meta;

type Story = StoryObj<typeof Sidebar>;

export const Default: Story = {
    render: (args) => (
        <Sidebar {...args}>
            <SidebarItem label="Dashboard" icon={IconBank} />
        </Sidebar>
    ),
};
