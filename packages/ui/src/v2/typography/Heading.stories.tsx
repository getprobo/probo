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

import { Heading } from "./Heading";
import { HeadingSkeleton } from "./HeadingSkeleton";

const sample = "The quick brown fox";

const sizes = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const;
const weights = ["light", "regular", "medium", "bold"] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;

export default {
  title: "v2/Typography/Heading",
  component: Heading,
  args: {
    children: sample,
    level: 1,
    size: 6,
    weight: "bold",
    color: "neutral",
    highContrast: false,
  },
} satisfies Meta<typeof Heading>;

type Story = StoryObj<typeof Heading>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {sizes.map(size => (
        <Heading key={size} size={size}>
          {size}
          {" — "}
          {sample}
        </Heading>
      ))}
    </div>
  ),
};

export const Weights: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {weights.map(weight => (
        <Heading key={weight} weight={weight}>
          {weight}
          {" — "}
          {sample}
        </Heading>
      ))}
    </div>
  ),
};

export const Align: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <Heading size={4} align="left">Left-aligned</Heading>
      <Heading size={4} align="center">Center-aligned</Heading>
      <Heading size={4} align="right">Right-aligned</Heading>
    </div>
  ),
};

export const Colors: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {colors.map(color => (
        <Heading key={color} size={4} color={color}>
          {color}
          {" — "}
          {sample}
        </Heading>
      ))}
    </div>
  ),
};

export const HighContrast: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      <Heading size={4}>Low contrast (step 11)</Heading>
      <Heading size={4} highContrast>High contrast (step 12)</Heading>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <HeadingSkeleton size={8} className="w-2/3" />
      <HeadingSkeleton size={6} className="w-1/2" />
    </div>
  ),
};
