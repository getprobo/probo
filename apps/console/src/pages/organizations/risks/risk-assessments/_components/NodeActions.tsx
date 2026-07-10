// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Field,
  IconPencil,
  IconTrashCan,
  Option,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { graphql, useMutation } from "react-relay";

import type { NodeActionsDeleteMutation } from "#/__generated__/core/NodeActionsDeleteMutation.graphql";
import type { NodeActionsUpdateMutation } from "#/__generated__/core/NodeActionsUpdateMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const updateNodeMutation = graphql`
  mutation NodeActionsUpdateMutation($input: UpdateRiskAssessmentNodeInput!) {
    updateRiskAssessmentNode(input: $input) {
      riskAssessmentNode { id nodeType name boundaryId }
    }
  }
`;

const deleteNodeMutation = graphql`
  mutation NodeActionsDeleteMutation(
    $input: DeleteRiskAssessmentNodeInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessmentNode(input: $input) {
      deletedRiskAssessmentNodeId @deleteEdge(connections: $connections)
    }
  }
`;

export function NodeActions(props: {
  node: { id: string; name: string; nodeType: string; boundaryId: string | null };
  boundaries: { id: string; name: string }[];
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateNode] = useMutation<NodeActionsUpdateMutation>(updateNodeMutation);
  const [deleteNode] = useMutation<NodeActionsDeleteMutation>(deleteNodeMutation);
  const { register, control, handleSubmit } = useForm({
    values: {
      name: props.node.name,
      nodeType: props.node.nodeType,
      boundaryId: props.node.boundaryId ?? "none",
    },
  });
  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {__("Edit")}
        </DropdownItem>
        <DropdownItem
          icon={IconTrashCan}
          variant="danger"
          onSelect={() => confirm(
            () => {
              deleteNode({
                variables: {
                  input: { riskAssessmentNodeId: props.node.id },
                  connections: [props.connectionId],
                },
              });
            },
            { message: __("Delete this node?") },
          )}
        >
          {__("Delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[__("Nodes"), __("Edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateNode({
            variables: { input: { id: props.node.id, name: d.name, nodeType: d.nodeType as "ENTITY" | "ASSET" | "DATA", boundaryId: d.boundaryId === "none" ? null : d.boundaryId } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <ControlledField label={__("Type")} name="nodeType" control={control} type="select">
              <Option value="ENTITY">{__("Entity")}</Option>
              <Option value="ASSET">{__("Asset")}</Option>
              <Option value="DATA">{__("Data")}</Option>
            </ControlledField>
            <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" />
            <ControlledField label={__("Boundary")} name="boundaryId" control={control} type="select">
              <Option value="none">{__("None")}</Option>
              {props.boundaries.map(b => (
                <Option key={b.id} value={b.id}>{b.name}</Option>
              ))}
            </ControlledField>
          </DialogContent>
          <DialogFooter><Button type="submit">{__("Save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
