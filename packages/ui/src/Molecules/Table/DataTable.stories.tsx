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
