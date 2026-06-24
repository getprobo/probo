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

import type { Meta, StoryObj } from "@storybook/react";

import { Card } from "./Card";
import { CardInset } from "./CardInset";
import { CardSkeleton } from "./CardSkeleton";

const sizes = [1, 2, 3, 4, 5] as const;
const variants = ["surface", "classic", "ghost"] as const;

function Sample() {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-2 font-medium text-sand-12">Teodros Girmay</span>
      <span className="text-2 text-sand-11">Engineering</span>
    </div>
  );
}

export default {
  title: "v2/Card",
  component: Card,
  args: {
    size: 1,
    variant: "surface",
    interactive: false,
  },
  render: args => (
    <div className="w-72">
      <Card {...args}><Sample /></Card>
    </div>
  ),
} satisfies Meta<typeof Card>;

type Story = StoryObj<typeof Card>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex w-72 flex-col gap-4">
      {sizes.map(size => (
        <Card key={size} size={size}><Sample /></Card>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex w-72 flex-col gap-4">
      {variants.map(variant => (
        <Card key={variant} variant={variant}><Sample /></Card>
      ))}
    </div>
  ),
};

export const Interactive: Story = {
  render: () => (
    <div className="flex w-72 flex-col gap-4">
      <Card interactive><Sample /></Card>
      <Card variant="ghost" interactive><Sample /></Card>
    </div>
  ),
};

export const Inset: Story = {
  render: () => (
    <div className="w-72">
      <Card size={2}>
        <CardInset side="top">
          <img
            src="https://images.unsplash.com/photo-1517433670267-08bbd4be890f?w=600"
            alt=""
            className="h-32 w-full object-cover"
          />
        </CardInset>
        <Sample />
      </Card>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="w-72">
      <CardSkeleton size={2} />
    </div>
  ),
};
