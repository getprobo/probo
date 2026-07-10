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

import type { BoundaryActionsDeleteMutation } from "#/__generated__/core/BoundaryActionsDeleteMutation.graphql";
import type { BoundaryActionsUpdateMutation } from "#/__generated__/core/BoundaryActionsUpdateMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const updateBoundaryMutation = graphql`
  mutation BoundaryActionsUpdateMutation($input: UpdateRiskAssessmentBoundaryInput!) {
    updateRiskAssessmentBoundary(input: $input) {
      riskAssessmentBoundary { id name parentBoundaryId }
    }
  }
`;

const deleteBoundaryMutation = graphql`
  mutation BoundaryActionsDeleteMutation(
    $input: DeleteRiskAssessmentBoundaryInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessmentBoundary(input: $input) {
      deletedRiskAssessmentBoundaryId @deleteEdge(connections: $connections)
    }
  }
`;

export function BoundaryActions(props: {
  boundary: { id: string; name: string; parentBoundaryId: string | null };
  boundaries: { id: string; name: string }[];
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateBoundary] = useMutation<BoundaryActionsUpdateMutation>(updateBoundaryMutation);
  const [deleteBoundary] = useMutation<BoundaryActionsDeleteMutation>(deleteBoundaryMutation);
  const { register, control, handleSubmit } = useForm({
    values: {
      name: props.boundary.name,
      parentBoundaryId: props.boundary.parentBoundaryId ?? "none",
    },
  });
  const parentOptions = props.boundaries.filter(b => b.id !== props.boundary.id);
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
              deleteBoundary({
                variables: {
                  input: { riskAssessmentBoundaryId: props.boundary.id },
                  connections: [props.connectionId],
                },
              });
            },
            { message: __("Delete this boundary? Nodes and nested boundaries inside it will be moved to the top level.") },
          )}
        >
          {__("Delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[__("Boundaries"), __("Edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateBoundary({
            variables: { input: { id: props.boundary.id, name: d.name, parentBoundaryId: d.parentBoundaryId === "none" ? null : d.parentBoundaryId } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" />
            <ControlledField label={__("Parent boundary")} name="parentBoundaryId" control={control} type="select">
              <Option value="none">{__("None (top level)")}</Option>
              {parentOptions.map(b => (
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
