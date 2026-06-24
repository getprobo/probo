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

import { ArrowSquareOutIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";

import { Anchor } from "./Anchor";

const sizes = [1, 2, 3, 4] as const;
const variants = ["classic", "solid", "soft", "surface", "outline", "ghost"] as const;

export default {
  title: "v2/Anchor",
  component: Anchor,
  args: {
    children: "Documentation",
    href: "#",
    size: 2,
    variant: "solid",
    color: "gold",
    highContrast: false,
    active: false,
  },
} satisfies Meta<typeof Anchor>;

type Story = StoryObj<typeof Anchor>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {sizes.map(size => (
        <Anchor key={size} href="#" size={size}>
          Documentation
        </Anchor>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {variants.map(variant => (
        <Anchor key={variant} href="#" variant={variant}>
          {variant}
        </Anchor>
      ))}
    </div>
  ),
};

export const WithIcon: Story = {
  render: () => (
    <Anchor href="#" variant="soft" color="neutral" iconEnd={<ArrowSquareOutIcon />}>
      External link
    </Anchor>
  ),
};
