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
