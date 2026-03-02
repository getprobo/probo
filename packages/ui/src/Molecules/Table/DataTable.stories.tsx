import type { Meta, StoryObj } from "@storybook/react";
import { type FC, Fragment, useState } from "react";
// eslint-disable-next-line import-x/no-unresolved -- storybook v10 subpath export
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
  // eslint-disable-next-line @typescript-eslint/no-unsafe-call -- fn type unresolvable due to storybook/test subpath
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
