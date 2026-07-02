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

import { GearIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";

import { IconButton } from "./IconButton";
import { IconButtonSkeleton } from "./IconButtonSkeleton";

const sizes = [1, 2, 3, 4] as const;
const variants = ["classic", "solid", "soft", "surface", "outline", "ghost"] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;

export default {
  title: "v2/IconButton",
  component: IconButton,
  args: {
    "aria-label": "Settings",
    "children": <GearIcon />,
    "size": 2,
    "variant": "solid",
    "color": "gold",
    "highContrast": false,
    "loading": false,
    "disabled": false,
  },
} satisfies Meta<typeof IconButton>;

type Story = StoryObj<typeof IconButton>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {sizes.map(size => (
        <IconButton key={size} size={size} aria-label="Settings">
          <GearIcon />
        </IconButton>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {variants.map(variant => (
        <IconButton key={variant} variant={variant} aria-label="Settings">
          <GearIcon />
        </IconButton>
      ))}
    </div>
  ),
};

export const Colors: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {variants.map(variant => (
        <div key={variant} className="flex items-center gap-3">
          {colors.map(color => (
            <IconButton key={color} variant={variant} color={color} aria-label="Settings">
              <GearIcon />
            </IconButton>
          ))}
        </div>
      ))}
    </div>
  ),
};

export const HighContrast: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {colors.map(color => (
        <IconButton key={color} color={color} highContrast aria-label="Settings">
          <GearIcon />
        </IconButton>
      ))}
    </div>
  ),
};

export const States: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <IconButton loading aria-label="Loading"><GearIcon /></IconButton>
      <IconButton disabled aria-label="Settings"><GearIcon /></IconButton>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <IconButtonSkeleton size={2} />
      <IconButtonSkeleton size={3} />
    </div>
  ),
};
