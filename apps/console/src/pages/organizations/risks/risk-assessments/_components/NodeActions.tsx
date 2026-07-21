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
  const { t } = useTranslation();
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
          {t("riskAssessmentNodeActions.actions.edit")}
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
            { message: t("riskAssessmentNodeActions.deleteConfirmation") },
          )}
        >
          {t("riskAssessmentNodeActions.actions.delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[t("riskAssessmentNodeActions.breadcrumb.nodes"), t("riskAssessmentNodeActions.actions.edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateNode({
            variables: { input: { id: props.node.id, name: d.name, nodeType: d.nodeType as "ENTITY" | "ASSET" | "DATA", boundaryId: d.boundaryId === "none" ? null : d.boundaryId } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <ControlledField label={t("riskAssessmentNodeActions.fields.type")} name="nodeType" control={control} type="select">
              <Option value="ENTITY">{t("riskAssessmentNodeActions.types.entity")}</Option>
              <Option value="ASSET">{t("riskAssessmentNodeActions.types.asset")}</Option>
              <Option value="DATA">{t("riskAssessmentNodeActions.types.data")}</Option>
            </ControlledField>
            <Field label={t("riskAssessmentNodeActions.fields.name")} {...register("name", { required: t("riskAssessmentNodeActions.validation.nameRequired") })} type="text" />
            <ControlledField label={t("riskAssessmentNodeActions.fields.boundary")} name="boundaryId" control={control} type="select">
              <Option value="none">{t("riskAssessmentNodeActions.none")}</Option>
              {props.boundaries.map(b => (
                <Option key={b.id} value={b.id}>{b.name}</Option>
              ))}
            </ControlledField>
          </DialogContent>
          <DialogFooter><Button type="submit">{t("riskAssessmentNodeActions.actions.save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
