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

import { Pagination } from "./Pagination";
import { PaginationSkeleton } from "./PaginationSkeleton";

const noop = () => {};

export default {
  title: "v2/Pagination",
  component: Pagination,
  args: {
    hasPrevious: true,
    hasNext: true,
    onPrevious: noop,
    onNext: noop,
  },
} satisfies Meta<typeof Pagination>;

type Story = StoryObj<typeof Pagination>;

export const Playground: Story = {};

// Each arrow only shows when its page exists, but its slot stays reserved so a
// lone arrow sits in the same position as when both are present.
export const States: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Pagination hasPrevious hasNext onPrevious={noop} onNext={noop} />
      <Pagination hasPrevious={false} hasNext onPrevious={noop} onNext={noop} />
      <Pagination hasPrevious hasNext={false} onPrevious={noop} onNext={noop} />
    </div>
  ),
};

export const WithLabel: Story = {
  args: {
    label: "Page 2",
  },
};

export const Skeleton: Story = {
  render: () => <PaginationSkeleton />,
};
