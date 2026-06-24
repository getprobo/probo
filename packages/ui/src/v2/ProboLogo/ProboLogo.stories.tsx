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

import type { Meta, StoryObj } from "@storybook/react";

import { ProboLogo } from "./ProboLogo";

export default {
  title: "v2/ProboLogo",
  component: ProboLogo,
  args: {
    className: "h-6 w-auto text-sand-12",
  },
  render: args => <ProboLogo {...args} />,
} satisfies Meta<typeof ProboLogo>;

type Story = StoryObj<typeof ProboLogo>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-6 text-sand-11">
      <ProboLogo className="h-4 w-auto" />
      <ProboLogo className="h-6 w-auto" />
      <ProboLogo className="h-10 w-auto" />
    </div>
  ),
};
