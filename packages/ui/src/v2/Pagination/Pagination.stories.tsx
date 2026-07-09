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
