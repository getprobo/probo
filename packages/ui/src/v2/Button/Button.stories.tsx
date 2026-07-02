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
