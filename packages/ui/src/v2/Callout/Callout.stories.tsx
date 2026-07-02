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
