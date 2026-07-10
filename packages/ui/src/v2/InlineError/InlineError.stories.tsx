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

import { InlineError } from "./InlineError";

const message = "Unable to load content";

export default {
  title: "v2/InlineError",
  component: InlineError,
  args: {
    layout: "vertical",
    message,
    onRetry: () => {},
  },
} satisfies Meta<typeof InlineError>;

type Story = StoryObj<typeof InlineError>;

export const Playground: Story = {
  render: args => (
    <div className="w-96">
      <InlineError {...args} />
    </div>
  ),
};

export const Vertical: Story = {
  render: () => (
    <div className="w-96">
      <InlineError layout="vertical" message={message} onRetry={() => {}} />
    </div>
  ),
};

export const Horizontal: Story = {
  render: () => (
    <div className="w-96">
      <InlineError layout="horizontal" message={message} onRetry={() => {}} />
    </div>
  ),
};

export const WithoutRetry: Story = {
  render: () => (
    <div className="w-96">
      <InlineError layout="vertical" message={message} />
    </div>
  ),
};
