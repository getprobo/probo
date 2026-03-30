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

import { Button } from "../Button/Button";
import { IconCircleQuestionmark1 } from "../Icons";

import { Dropdown, DropdownItem, DropdownSeparator } from "./Dropdown";

export default {
  title: "Atoms/Dropdown",
  component: Dropdown,
  argTypes: {},
  subcomponents: {
    DropdownItem,
    DropdownSeparator,
  },
} satisfies Meta<typeof Dropdown>;

type Story = StoryObj<typeof Dropdown>;

export const Default: Story = {
  render: () => {
    return (
      <Dropdown toggle={<Button variant="tertiary">Dropdown</Button>}>
        <DropdownItem icon={IconCircleQuestionmark1}>
          Item 1
        </DropdownItem>
        <DropdownSeparator />
        <DropdownItem icon={IconCircleQuestionmark1}>
          Item 2
        </DropdownItem>
        <DropdownItem icon={IconCircleQuestionmark1}>
          Item 3
        </DropdownItem>
      </Dropdown>
    );
  },
};
