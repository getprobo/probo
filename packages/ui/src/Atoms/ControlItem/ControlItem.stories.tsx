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

import { ControlItem } from "./ControlItem";

export default {
  title: "Atoms/ControlItem",
  component: ControlItem,
  argTypes: {},
} satisfies Meta<typeof ControlItem>;

type Story = StoryObj<typeof ControlItem>;

export const Default: Story = {
  args: {
    id: "CC1.1",
    description:
            "The entity obtains privacy commitments from thirdParties and other third parties who have access to personal information to meet the entity’s objectives related to privacy. The entity assesses those parties’ compliance on a periodic and as-needed basis and takes corrective action, if necessary.",
  },
  render: args => (
    <div className="p-4 space-y-2" style={{ width: "240px" }}>
      <ControlItem {...args} />
      <ControlItem {...args} active />
      <ControlItem {...args} />
      <ControlItem {...args} />
      <ControlItem {...args} />
    </div>
  ),
};
