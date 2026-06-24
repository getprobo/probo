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

import { ArrowRightIcon, BookmarkSimpleIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";

import { Button } from "./Button";
import { ButtonSkeleton } from "./ButtonSkeleton";

const sizes = [1, 2, 3, 4] as const;
const variants = ["classic", "solid", "soft", "surface", "outline", "ghost"] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;

export default {
  title: "v2/Button",
  component: Button,
  args: {
    children: "Edit profile",
    size: 2,
    variant: "solid",
    color: "gold",
    highContrast: false,
    loading: false,
    disabled: false,
  },
} satisfies Meta<typeof Button>;

type Story = StoryObj<typeof Button>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {sizes.map(size => (
        <Button key={size} size={size}>
          Edit profile
        </Button>
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {variants.map(variant => (
        <Button key={variant} variant={variant}>
          {variant}
        </Button>
      ))}
    </div>
  ),
};

export const Colors: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {variants.map(variant => (
        <div key={variant} className="flex items-center gap-3">
          {colors.map(color => (
            <Button key={color} variant={variant} color={color}>
              {color}
            </Button>
          ))}
        </div>
      ))}
    </div>
  ),
};

export const HighContrast: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {colors.map(color => (
        <Button key={color} color={color} highContrast>
          {color}
        </Button>
      ))}
    </div>
  ),
};

export const WithIcons: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button iconStart={<BookmarkSimpleIcon />}>Bookmark</Button>
      <Button variant="soft" color="neutral" iconEnd={<ArrowRightIcon />}>Next</Button>
    </div>
  ),
};

export const Loading: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button loading>Saving</Button>
      <Button variant="soft" color="neutral" loading>Saving</Button>
    </div>
  ),
};

export const Disabled: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button disabled>Solid</Button>
      <Button variant="soft" disabled>Soft</Button>
      <Button variant="outline" color="neutral" disabled>Outline</Button>
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <ButtonSkeleton size={2} />
      <ButtonSkeleton size={3} />
    </div>
  ),
};
