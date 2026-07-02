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

import type { Meta, StoryObj } from "@storybook/react";
import { type FC, Fragment, useState } from "react";
import { fn } from "storybook/test";

import { Badge } from "../../Atoms/Badge/Badge";
import { CellHead, DataTable, Row } from "../../Atoms/DataTable/DataTable";

import { EditableRow } from "./EditableRow";
import { SelectCell } from "./SelectCell";
import { TextCell } from "./TextCell";

type Component = FC<{ onUpdate: (key: string, value: unknown) => void }>;

export default {
  title: "Atoms/DataTable/Cells",
  component: Fragment as Component,
  argTypes: {},

  args: { onUpdate: fn() as (key: string, value: unknown) => void },
} satisfies Meta<Component>;

type Story = StoryObj<Component>;

export const Default: Story = {
  render: function Render({ onUpdate }) {
    const [state, setState] = useState({
      name: "John",
      status: "delivered",
      statuses: ["delivered", "pending"],
    });
    const updateField = (key: string, value: unknown) => {
      onUpdate(key, value);
      setState(prevState => ({
        ...prevState,
        [key]: value,
      }));
    };
    return (
      <DataTable columns={["1fr", "1fr", "1fr"]}>
        <Row>
          <CellHead>Nom</CellHead>
          <CellHead>Status</CellHead>
          <CellHead>Statuses</CellHead>
        </Row>
        <EditableRow onUpdate={updateField}>
          <TextCell required name="name" defaultValue={state.name} />
          <SelectCell
            items={["delivered", "pending"]}
            itemRenderer={({ item }) => <Badge>{item}</Badge>}
            name="status"
            defaultValue={state.status}
          />
          <SelectCell
            multiple
            items={["delivered", "pending"]}
            itemRenderer={({ item, onRemove }) => (
              <Badge onClick={() => onRemove?.(item)}>
                {item}
              </Badge>
            )}
            name="statuses"
            defaultValue={state.statuses}
          />
        </EditableRow>
      </DataTable>
    );
  },
};
