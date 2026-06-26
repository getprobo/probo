// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
