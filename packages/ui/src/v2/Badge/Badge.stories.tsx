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

import { CheckCircleIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";

import { Badge } from "./Badge";
import { BadgeSkeleton } from "./BadgeSkeleton";

const sizes = [1, 2, 3] as const;
const variants = ["solid", "soft", "surface", "outline"] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;

export default {
  title: "v2/Badge",
  component: Badge,
  args: {
    children: "Badge",
    size: 1,
    variant: "soft",
    color: "neutral",
    highContrast: false,
  },
} satisfies Meta<typeof Badge>;

type Story = StoryObj<typeof Badge>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {sizes.map(size => (
        <Badge key={size} size={size}>Badge</Badge>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {variants.map(variant => (
        <Badge key={variant} variant={variant}>{variant}</Badge>
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
            <Badge key={color} variant={variant} color={color}>{color}</Badge>
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
        <Badge key={color} color={color} highContrast>{color}</Badge>
      ))}
    </div>
  ),
};

export const WithIcon: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Badge color="green" iconStart={<CheckCircleIcon weight="fill" />}>Approved</Badge>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <BadgeSkeleton size={1} />
      <BadgeSkeleton size={2} />
    </div>
  ),
};
