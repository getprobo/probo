import type { Meta, StoryObj } from "@storybook/react";

import { TabBadge, TabLink, Tabs } from "./Tabs";

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
      <TabLink to="/demo">Tab 2</TabLink>
      <TabLink to="/demo2">
        Tab 3
        {" "}
        <TabBadge>3</TabBadge>
      </TabLink>
    </Tabs>
  ),
};
