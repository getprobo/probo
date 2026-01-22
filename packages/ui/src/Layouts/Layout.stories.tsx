import type { Meta, StoryObj } from "@storybook/react";

import { Drawer, Layout } from "./Layout";

const meta = {
  title: "Layouts/Base",
  component: Layout,
  parameters: {
    layout: "full",
  },
} satisfies Meta<typeof Layout>;

export default meta;
type Story = StoryObj<typeof Layout>;

export const Default: Story = {
  args: {
    children: (
      <p>
        Lorem ipsum dolor sit amet consectetur, adipisicing elit.
        Tempora tempore odit at ipsa dignissimos cumque deleniti, illum
        rerum sunt laudantium molestias ducimus vero maiores eligendi
        necessitatibus nemo animi. Ipsam, fugit.
      </p>
    ),
  },
};

export const WithSidebar: Story = {
  args: {
    children: (
      <>
        <p>
          Lorem ipsum dolor sit amet consectetur, adipisicing elit.
          Tempora tempore odit at ipsa dignissimos cumque deleniti,
          illum rerum sunt laudantium molestias ducimus vero maiores
          eligendi necessitatibus nemo animi. Ipsam, fugit.
        </p>
        <Drawer>
          Lorem ipsum dolor sit amet consectetur adipisicing elit.
          Nostrum reiciendis eveniet possimus illo rerum labore nisi
          voluptatibus sequi consectetur corrupti, a, sunt dicta ut ad
          pariatur. Beatae optio libero pariatur!
        </Drawer>
      </>
    ),
  },
};
