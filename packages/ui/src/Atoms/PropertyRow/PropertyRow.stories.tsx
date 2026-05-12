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

import { Select } from "../Select/Select";

import { PropertyRow } from "./PropertyRow";

export default {
  title: "Atoms/PropertyRow",
  component: PropertyRow,
  argTypes: {},
} satisfies Meta<typeof PropertyRow>;

type Story = StoryObj<typeof PropertyRow>;

export const Default: Story = {
  args: {
    id: "test",
    label: "Test",
    children: <Select variant="editor" placeholder="Select an otion" />,
  },
};

export const WithError: Story = {
  args: {
    ...Default.args,
    error: "This is an error",
  },
};
