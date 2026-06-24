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
