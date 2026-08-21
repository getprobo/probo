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

import { Spinner } from "./Spinner";
import { SpinnerSkeleton } from "./SpinnerSkeleton";

const sizes = [1, 2, 3] as const;

export default {
  title: "v2/Spinner",
  component: Spinner,
  args: {
    "size": 2,
    "aria-label": "Loading",
  },
} satisfies Meta<typeof Spinner>;

type Story = StoryObj<typeof Spinner>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      {sizes.map(size => (
        <Spinner key={size} size={size} aria-label={`Size ${size}`} />
      ))}
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <SpinnerSkeleton size={1} />
      <SpinnerSkeleton size={2} />
      <SpinnerSkeleton size={3} />
    </div>
  ),
};
