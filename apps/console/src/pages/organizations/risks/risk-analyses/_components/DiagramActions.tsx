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

import type { DiagramActionsDeleteMutation } from "#/__generated__/core/DiagramActionsDeleteMutation.graphql";
import type { DiagramActionsUpdateMutation } from "#/__generated__/core/DiagramActionsUpdateMutation.graphql";

const updateDiagramMutation = graphql`
  mutation DiagramActionsUpdateMutation(
    $input: UpdateRiskAnalysisDiagramInput!
  ) {
    updateRiskAnalysisDiagram(input: $input) {
      riskAnalysisDiagram { id name }
    }
  }
`;

const deleteDiagramMutation = graphql`
  mutation DiagramActionsDeleteMutation(
    $input: DeleteRiskAnalysisDiagramInput!
    $connections: [ID!]!
  ) {
    deleteRiskAnalysisDiagram(input: $input) {
      deletedRiskAnalysisDiagramId @deleteEdge(connections: $connections)
    }
  }
`;

export function DiagramActions(props: {
  diagram: { id: string; name: string };
  connectionId: string;
}) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateDiagram] = useMutation<DiagramActionsUpdateMutation>(updateDiagramMutation);
  const [deleteDiagram] = useMutation<DiagramActionsDeleteMutation>(deleteDiagramMutation);
  const { register, handleSubmit, formState } = useForm({
    values: {
      name: props.diagram.name,
    },
  });

  const onEdit = (data: { name: string }) => {
    updateDiagram({
      variables: {
        input: {
          id: props.diagram.id,
          name: data.name,
        },
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
    });
  };

  const onDelete = () => {
    confirm(
      () => {
        deleteDiagram({
          variables: {
            input: { riskAnalysisDiagramId: props.diagram.id },
            connections: [props.connectionId],
          },
        });
      },
      { message: t("riskAnalysisDiagramActions.deleteConfirmation") },
    );
  };

  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {t("riskAnalysisDiagramActions.actions.edit")}
        </DropdownItem>
        <DropdownItem icon={IconTrashCan} variant="danger" onSelect={onDelete}>
          {t("riskAnalysisDiagramActions.actions.delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog
        className="max-w-lg"
        ref={dialogRef}
        title={<Breadcrumb items={[t("riskAnalysisDiagramActions.breadcrumb.diagrams"), t("riskAnalysisDiagramActions.breadcrumb.editDiagram")]} />}
      >
        <form onSubmit={e => void handleSubmit(onEdit)(e)}>
          <DialogContent padded className="space-y-4">
            <Field
              label={t("riskAnalysisDiagramActions.fields.name")}
              {...register("name", { required: t("riskAnalysisDiagramActions.validation.nameRequired") })}
              type="text"
              error={formState.errors.name?.message}
            />
          </DialogContent>
          <DialogFooter>
            <Button type="submit">{t("riskAnalysisDiagramActions.actions.save")}</Button>
          </DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
