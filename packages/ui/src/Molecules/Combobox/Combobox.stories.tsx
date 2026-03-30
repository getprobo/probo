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

import { times } from "@probo/helpers";
import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";

import { Combobox, ComboboxItem } from "./Combobox";

export default {
  title: "Atoms/Combobox",
  component: Combobox,
  argTypes: {},
} satisfies Meta<typeof Combobox>;

type Story = StoryObj<typeof Combobox>;

export const Default: Story = {
  render: function Render() {
    const [items, setItems] = useState(["a", "b", "c"] as string[]);
    const onSearch = (query: string) => {
      setItems(times(10, i => `${query} ${i}`));
    };
    return (
      <>
        <Combobox onSearch={onSearch}>
          {items.map(item => (
            <ComboboxItem key={item}>{item}</ComboboxItem>
          ))}
        </Combobox>
      </>
    );
  },
};
