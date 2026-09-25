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

import { Code } from "./Code";
import { CodeSkeleton } from "./CodeSkeleton";

const sample = "npm run dev";

const sizes = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const;
const variants = ["solid", "soft", "outline", "ghost"] as const;
const weights = ["regular", "bold"] as const;

export default {
  title: "v2/Typography/Code",
  component: Code,
  args: {
    children: sample,
    size: 2,
    variant: "soft",
    weight: "regular",
    highContrast: false,
  },
} satisfies Meta<typeof Code>;

type Story = StoryObj<typeof Code>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-col items-start gap-3">
      {sizes.map(size => (
        <Code key={size} size={size}>
          {size}
          {" — "}
          {sample}
        </Code>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {variants.map(variant => (
        <Code key={variant} variant={variant}>{variant}</Code>
      ))}
    </div>
  ),
};

export const Weights: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {weights.map(weight => (
        <Code key={weight} weight={weight}>{weight}</Code>
      ))}
    </div>
  ),
};

export const HighContrast: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {variants.map(variant => (
        <Code key={variant} variant={variant} highContrast>
          {variant}
        </Code>
      ))}
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      <CodeSkeleton size={2} className="w-24" />
      <CodeSkeleton size={4} className="w-32" />
    </div>
  ),
};
