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
