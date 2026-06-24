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

import { WarningIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";

import { Callout } from "./Callout";
import { CalloutSkeleton } from "./CalloutSkeleton";

const message = "You will need admin privileges to install and access this application.";

const sizes = [1, 2, 3] as const;
const variants = ["soft", "surface", "outline"] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;

export default {
  title: "v2/Callout",
  component: Callout,
  args: {
    children: message,
    size: 2,
    variant: "soft",
    color: "neutral",
    highContrast: false,
  },
} satisfies Meta<typeof Callout>;

type Story = StoryObj<typeof Callout>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      {sizes.map(size => (
        <Callout key={size} size={size}>{message}</Callout>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      {variants.map(variant => (
        <Callout key={variant} variant={variant}>{message}</Callout>
      ))}
    </div>
  ),
};

export const Colors: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      {colors.map(color => (
        <Callout key={color} color={color}>{message}</Callout>
      ))}
    </div>
  ),
};

export const HighContrast: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <Callout color="sky">{message}</Callout>
      <Callout color="sky" highContrast>{message}</Callout>
    </div>
  ),
};

export const CustomIcon: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <Callout color="amber" icon={<WarningIcon weight="fill" />}>{message}</Callout>
      <Callout color="neutral" icon={null}>{message}</Callout>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <CalloutSkeleton size={2} />
    </div>
  ),
};
