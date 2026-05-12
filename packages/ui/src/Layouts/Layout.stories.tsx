// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
