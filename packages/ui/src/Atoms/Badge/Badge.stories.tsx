// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
