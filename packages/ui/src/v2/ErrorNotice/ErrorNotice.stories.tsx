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

import { Card } from "../Card/Card";

import { ErrorNotice } from "./ErrorNotice";

const message = "Google Workspace credentials are invalid. Probo cannot read this vendor with the current credentials.";

const messages = [
  "Reconnect Google Workspace to restore access.",
  message,
] as const;

export default {
  title: "v2/ErrorNotice",
  component: ErrorNotice,
  args: {
    messages: [message],
    copyLabel: "Copy error",
  },
} satisfies Meta<typeof ErrorNotice>;

type Story = StoryObj<typeof ErrorNotice>;

export const OneMessage: Story = {
  render: args => (
    <div className="w-96">
      <ErrorNotice {...args} />
    </div>
  ),
};

export const SeveralMessages: Story = {
  render: () => (
    <div className="w-96">
      <ErrorNotice messages={messages} copyLabel="Copy error" />
    </div>
  ),
};

export const InCard: Story = {
  render: () => (
    <div className="w-96">
      <Card size={2}>
        <ErrorNotice messages={[message]} copyLabel="Copy error" />
      </Card>
    </div>
  ),
};
