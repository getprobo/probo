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

import { Text } from "./Text";
import { TextSkeleton } from "./TextSkeleton";

const sample = "The quick brown fox jumps over the lazy dog.";

const sizes = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const;
const weights = ["light", "regular", "medium", "bold"] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;

export default {
  title: "v2/Typography/Text",
  component: Text,
  args: {
    children: sample,
    size: 3,
    weight: "regular",
    color: "neutral",
    highContrast: false,
  },
} satisfies Meta<typeof Text>;

type Story = StoryObj<typeof Text>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {sizes.map(size => (
        <Text key={size} className="block" size={size}>
          {size}
          {" — "}
          {sample}
        </Text>
      ))}
    </div>
  ),
};

export const Weights: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {weights.map(weight => (
        <Text key={weight} className="block" weight={weight}>
          {weight}
          {" — "}
          {sample}
        </Text>
      ))}
    </div>
  ),
};

export const Align: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <Text className="block" align="left">Left-aligned</Text>
      <Text className="block" align="center">Center-aligned</Text>
      <Text className="block" align="right">Right-aligned</Text>
    </div>
  ),
};

export const Colors: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {colors.map(color => (
        <Text key={color} className="block" color={color}>
          {color}
          {" — "}
          {sample}
        </Text>
      ))}
    </div>
  ),
};

export const HighContrast: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      <Text className="block">Low contrast (step 11)</Text>
      <Text className="block" highContrast>High contrast (step 12)</Text>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex w-96 flex-col gap-3">
      <TextSkeleton size={6} className="w-1/2" />
      <TextSkeleton size={3} className="w-full" />
      <TextSkeleton size={3} className="w-2/3" />
    </div>
  ),
};
