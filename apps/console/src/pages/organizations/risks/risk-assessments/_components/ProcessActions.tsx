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
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";

import type { ProcessActionsDeleteMutation } from "#/__generated__/core/ProcessActionsDeleteMutation.graphql";
import type { ProcessActionsUpdateMutation } from "#/__generated__/core/ProcessActionsUpdateMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const updateProcessMutation = graphql`
  mutation ProcessActionsUpdateMutation($input: UpdateRiskAssessmentProcessInput!) {
    updateRiskAssessmentProcess(input: $input) {
      riskAssessmentProcess { id sourceNodeId targetNodeId name }
    }
  }
`;

const deleteProcessMutation = graphql`
  mutation ProcessActionsDeleteMutation(
    $input: DeleteRiskAssessmentProcessInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessmentProcess(input: $input) {
      deletedRiskAssessmentProcessId @deleteEdge(connections: $connections)
    }
  }
`;

export function ProcessActions(props: {
  process: { id: string; name: string; sourceNodeId: string; targetNodeId: string };
  nodes: { id: string; name: string }[];
  connectionId: string;
}) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateProcess] = useMutation<ProcessActionsUpdateMutation>(updateProcessMutation);
  const [deleteProcess] = useMutation<ProcessActionsDeleteMutation>(deleteProcessMutation);
  const { register, control, handleSubmit } = useForm({
    values: {
      name: props.process.name,
      sourceNodeId: props.process.sourceNodeId,
      targetNodeId: props.process.targetNodeId,
    },
  });
  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {t("riskAssessmentProcessActions.actions.edit")}
        </DropdownItem>
        <DropdownItem
          icon={IconTrashCan}
          variant="danger"
          onSelect={() => confirm(
            () => {
              deleteProcess({
                variables: {
                  input: { riskAssessmentProcessId: props.process.id },
                  connections: [props.connectionId],
                },
              });
            },
            { message: t("riskAssessmentProcessActions.deleteConfirmation") },
          )}
        >
          {t("riskAssessmentProcessActions.actions.delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[t("riskAssessmentProcessActions.breadcrumb.processes"), t("riskAssessmentProcessActions.actions.edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateProcess({
            variables: {
              input: {
                id: props.process.id,
                name: d.name,
                sourceNodeId: d.sourceNodeId,
                targetNodeId: d.targetNodeId,
              },
            },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <ControlledField label={t("riskAssessmentProcessActions.fields.source")} name="sourceNodeId" control={control} type="select" placeholder={t("riskAssessmentProcessActions.placeholders.sourceNode")}>
              {props.nodes.map(n => <Option key={n.id} value={n.id}>{n.name}</Option>)}
            </ControlledField>
            <ControlledField label={t("riskAssessmentProcessActions.fields.target")} name="targetNodeId" control={control} type="select" placeholder={t("riskAssessmentProcessActions.placeholders.targetNode")}>
              {props.nodes.map(n => <Option key={n.id} value={n.id}>{n.name}</Option>)}
            </ControlledField>
            <Field label={t("riskAssessmentProcessActions.fields.name")} {...register("name", { required: t("riskAssessmentProcessActions.validation.nameRequired") })} type="text" />
          </DialogContent>
          <DialogFooter><Button type="submit">{t("riskAssessmentProcessActions.actions.save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
