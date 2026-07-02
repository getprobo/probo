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

import { UserIcon } from "@phosphor-icons/react";
import type { Meta, StoryObj } from "@storybook/react";

import { Avatar } from "./Avatar";
import { AvatarSkeleton } from "./AvatarSkeleton";

const image = "https://i.pravatar.cc/240?img=12";

const sizes = [1, 2, 3, 4, 5, 6, 7, 8, 9] as const;
const colors = ["neutral", "gold", "red", "green", "amber", "sky"] as const;
const radii = ["none", "small", "medium", "large", "full"] as const;

export default {
  title: "v2/Avatar",
  component: Avatar,
  args: {
    fallback: "AB",
    size: 3,
    variant: "soft",
    color: "neutral",
    highContrast: false,
    radius: "medium",
  },
} satisfies Meta<typeof Avatar>;

type Story = StoryObj<typeof Avatar>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-wrap items-end gap-3">
      {sizes.map(size => (
        <Avatar key={size} size={size} fallback="AB" />
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      {(["soft", "solid"] as const).map(variant => (
        <div key={variant} className="flex items-center gap-3">
          {colors.map(color => (
            <Avatar key={color} variant={variant} color={color} size={4} fallback="AB" />
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
        <Avatar key={color} variant="solid" color={color} size={4} fallback="AB" highContrast />
      ))}
    </div>
  ),
};

export const Radius: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      {radii.map(radius => (
        <Avatar key={radius} radius={radius} size={5} fallback="AB" />
      ))}
    </div>
  ),
};

export const Fallback: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Avatar size={5} fallback="AB" />
      <Avatar size={5} color="gold" fallback={<UserIcon weight="fill" />} />
      <Avatar size={5} src={image} alt="User" fallback="AB" />
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <AvatarSkeleton size={3} />
      <AvatarSkeleton size={5} />
      <AvatarSkeleton size={6} radius="full" />
    </div>
  ),
};
