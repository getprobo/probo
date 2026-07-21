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
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
          {t("riskAssessmentThreatActions.actions.edit")}
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
            { message: t("riskAssessmentThreatActions.deleteConfirmation") },
          )}
        >
          {t("riskAssessmentThreatActions.actions.delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[t("riskAssessmentThreatActions.breadcrumb.threats"), t("riskAssessmentThreatActions.actions.edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateThreat({
            variables: { input: { id: props.threat.id, name: d.name, category: d.category } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <Field label={t("riskAssessmentThreatActions.fields.name")} {...register("name", { required: t("riskAssessmentThreatActions.validation.required") })} type="text" />
            <Field
              label={t("riskAssessmentThreatActions.fields.category")}
              {...register("category", { required: t("riskAssessmentThreatActions.validation.required") })}
              type="text"
              placeholder={t("riskAssessmentThreatActions.placeholders.category")}
            />
          </DialogContent>
          <DialogFooter><Button type="submit">{t("riskAssessmentThreatActions.actions.save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
