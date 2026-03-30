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

import { randomInt, times } from "@probo/helpers";
import type { Meta, StoryObj } from "@storybook/react";

import { RisksChart } from "./RisksChart";

export default {
  title: "Molecules/Risks/RisksChart",
  component: RisksChart,
  argTypes: {},
} satisfies Meta<typeof RisksChart>;

type Story = StoryObj<typeof RisksChart>;

export const Default: Story = {
  args: {
    type: "inherent",
    organizationId: "1",
    risks: times(20, i => ({
      id: i.toString(),
      name: `Risk ${i}`,
      inherentLikelihood: randomInt(1, 5),
      inherentImpact: randomInt(1, 5),
      residualLikelihood: randomInt(1, 5),
      residualImpact: randomInt(1, 5),
    })),
  },
  render: args => (
    <div style={{ maxWidth: 630 }}>
      <RisksChart {...args} />
    </div>
  ),
};
