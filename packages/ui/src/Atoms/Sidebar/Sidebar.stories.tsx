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

import { IconBank } from "../Icons/IconBank";

import { Sidebar } from "./Sidebar";
import { SidebarItem } from "./SidebarItem";

const meta: Meta<typeof Sidebar> = {
  component: Sidebar,
  subcomponents: { SidebarItem },
};
export default meta;

type Story = StoryObj<typeof Sidebar>;

export const Default: Story = {
  render: args => (
    <Sidebar {...args}>
      <SidebarItem label="Dashboard" icon={IconBank} />
    </Sidebar>
  ),
};
