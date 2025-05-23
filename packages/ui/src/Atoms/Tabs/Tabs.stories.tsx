import { Tabs, TabLink } from "./Tabs";
import type { Meta, StoryObj } from "@storybook/react";

export default {
    title: "Atoms/Tabs",
    component: Tabs,
    argTypes: {},
} satisfies Meta<typeof Tabs>;

type Story = StoryObj<typeof Tabs>;

export const Default: Story = {
    render: () => (
        <Tabs>
            <TabLink to="#">Tab 1</TabLink>
            <TabLink to="#">Tab 2</TabLink>
            <TabLink to="#">Tab 3</TabLink>
        </Tabs>
    ),
};
