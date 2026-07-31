// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { ArrowRightIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";
import { MemoryRouter } from "react-router";

import { Link } from "./Link";

const sizes = [1, 2, 3, 4] as const;

export default {
  title: "v2/Link",
  component: Link,
  args: {
    children: "View all updates",
    to: "#",
    size: 2,
    color: "neutral",
    highContrast: false,
  },
  decorators: [
    Story => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
} satisfies Meta<typeof Link>;

type Story = StoryObj<typeof Link>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-col items-start gap-3">
      {sizes.map(size => (
        <Link key={size} to="#" size={size}>
          View all updates
        </Link>
      ))}
    </div>
  ),
};

export const WithIcon: Story = {
  render: () => (
    <Link to="#" size={2} iconEnd={<ArrowRightIcon />}>
      View all updates
    </Link>
  ),
};

export const Plain: Story = {
  args: {
    underline: false,
  },
};
