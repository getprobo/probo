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
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { graphql, useMutation } from "react-relay";

import type { ThreatActionsDeleteMutation } from "#/__generated__/core/ThreatActionsDeleteMutation.graphql";
import type { ThreatActionsUpdateMutation } from "#/__generated__/core/ThreatActionsUpdateMutation.graphql";

const updateThreatMutation = graphql`
  mutation ThreatActionsUpdateMutation($input: UpdateRiskAssessmentThreatInput!) {
    updateRiskAssessmentThreat(input: $input) {
      riskAssessmentThreat { id processId name category }
    }
  }
`;

const deleteThreatMutation = graphql`
  mutation ThreatActionsDeleteMutation(
    $input: DeleteRiskAssessmentThreatInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessmentThreat(input: $input) {
      deletedRiskAssessmentThreatId @deleteEdge(connections: $connections)
    }
  }
`;

export function ThreatActions(props: {
  threat: { id: string; name: string; category: string };
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateThreat] = useMutation<ThreatActionsUpdateMutation>(updateThreatMutation);
  const [deleteThreat] = useMutation<ThreatActionsDeleteMutation>(deleteThreatMutation);
  const { register, handleSubmit } = useForm({
    values: { name: props.threat.name, category: props.threat.category },
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
              deleteThreat({
                variables: {
                  input: { riskAssessmentThreatId: props.threat.id },
                  connections: [props.connectionId],
                },
              });
            },
            { message: __("Delete this threat?") },
          )}
        >
          {__("Delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[__("Threats"), __("Edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateThreat({
            variables: { input: { id: props.threat.id, name: d.name, category: d.category } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" />
            <Field
              label={__("Category")}
              {...register("category", { required: __("This field is required") })}
              type="text"
              placeholder={__("e.g. Confidentiality")}
            />
          </DialogContent>
          <DialogFooter><Button type="submit">{__("Save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
