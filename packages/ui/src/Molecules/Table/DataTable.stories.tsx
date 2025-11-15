import type { Meta, StoryObj } from "@storybook/react";
import { type FC, Fragment, useState } from "react";
import { CellHead, DataTable, Row } from "../../Atoms/DataTable/DataTable.tsx";
import { EditableRow } from "./EditableRow.tsx";
import { TextCell } from "./TextCell.tsx";
import { fn } from "@storybook/test";
import { SelectCell } from "./SelectCell.tsx";
import { Badge } from "../../Atoms/Badge/Badge.tsx";

type Component = FC<{ onUpdate: (key: string, value: unknown) => void }>;

export default {
    title: "Atoms/DataTable/Cells",
    component: Fragment as Component,
    argTypes: {},
    args: { onUpdate: fn() },
} satisfies Meta<Component>;

type Story = StoryObj<Component>;

export const Default: Story = {
    render: ({ onUpdate }) => {
        const [state, setState] = useState({
            name: "John",
            status: "delivered",
            statuses: ["delivered", "pending"],
        });
        const updateField = (key: string, value: unknown) => {
            onUpdate(key, value);
            setState({
                ...state,
                [key]: value,
            });
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
