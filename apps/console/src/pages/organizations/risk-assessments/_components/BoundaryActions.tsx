// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
