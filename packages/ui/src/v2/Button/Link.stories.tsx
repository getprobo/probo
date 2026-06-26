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
import { MemoryRouter } from "react-router";

import { Link } from "./Link";

const navItems = ["Documents", "Subprocessors", "Updates", "Requests"] as const;

export default {
  title: "v2/Link",
  component: Link,
  args: {
    children: "Documents",
    to: "#",
    size: 2,
    variant: "ghost",
    color: "neutral",
    highContrast: false,
    active: false,
  },
  decorators: [
    Story => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
} satisfies Meta<typeof Link>;

type Story = StoryObj<typeof Link>;

export const Playground: Story = {};

// Mirrors the Trust Center top-bar nav: ghost links with a persistent active
// pill on the selected item.
export const Nav: Story = {
  render: () => (
    <div className="flex items-center gap-1">
      {navItems.map((item, index) => (
        <Link
          key={item}
          to="#"
          variant="ghost"
          color="neutral"
          size={2}
          active={index === 0}
        >
          {item}
        </Link>
      ))}
    </div>
  ),
};
