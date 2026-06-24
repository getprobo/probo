// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
