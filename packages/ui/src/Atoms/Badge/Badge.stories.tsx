// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { Badge } from "./Badge";

export default {
  title: "Atoms/Badge",
  component: Badge,
  argTypes: {},
} satisfies Meta<typeof Badge>;

type Story = StoryObj<typeof Badge>;

const sizes = ["sm", "md"] as const;
const variants = [
  "success",
  "warning",
  "danger",
  "info",
  "neutral",
  "outline",
  "highlight",
] as const;

export const Default: Story = {
  render: () => (
    <div className="space-y-4">
      {sizes.map(size => (
        <div key={size} className="space-y-1">
          <div className="text-xs text-txt-secondary">
            Size :
            {" "}
            {size}
          </div>
          <div className="flex gap-2">
            {variants.map(variant => (
              <Badge key={variant} variant={variant} size={size}>
                {variant}
              </Badge>
            ))}
          </div>
        </div>
      ))}
    </div>
  ),
};
