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
