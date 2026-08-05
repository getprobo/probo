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

import type { ScopeActionsDeleteMutation } from "#/__generated__/core/ScopeActionsDeleteMutation.graphql";
import type { ScopeActionsUpdateMutation } from "#/__generated__/core/ScopeActionsUpdateMutation.graphql";

const updateScopeMutation = graphql`
  mutation ScopeActionsUpdateMutation(
    $input: UpdateRiskAnalysisScopeInput!
  ) {
    updateRiskAnalysisScope(input: $input) {
      riskAnalysisScope { id name }
    }
  }
`;

const deleteScopeMutation = graphql`
  mutation ScopeActionsDeleteMutation(
    $input: DeleteRiskAnalysisScopeInput!
    $connections: [ID!]!
  ) {
    deleteRiskAnalysisScope(input: $input) {
      deletedRiskAnalysisScopeId @deleteEdge(connections: $connections)
    }
  }
`;

export function ScopeActions(props: {
  scope: { id: string; name: string };
  connectionId: string;
}) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateScope] = useMutation<ScopeActionsUpdateMutation>(updateScopeMutation);
  const [deleteScope] = useMutation<ScopeActionsDeleteMutation>(deleteScopeMutation);
  const { register, handleSubmit, formState } = useForm({
    values: {
      name: props.scope.name,
    },
  });

  const onEdit = (data: { name: string }) => {
    updateScope({
      variables: {
        input: {
          id: props.scope.id,
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
        deleteScope({
          variables: {
            input: { riskAnalysisScopeId: props.scope.id },
            connections: [props.connectionId],
          },
        });
      },
      { message: t("riskAnalysisScopeActions.deleteConfirmation") },
    );
  };

  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {t("riskAnalysisScopeActions.actions.edit")}
        </DropdownItem>
        <DropdownItem icon={IconTrashCan} variant="danger" onSelect={onDelete}>
          {t("riskAnalysisScopeActions.actions.delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog
        className="max-w-lg"
        ref={dialogRef}
        title={<Breadcrumb items={[t("riskAnalysisScopeActions.breadcrumb.scopes"), t("riskAnalysisScopeActions.breadcrumb.editScope")]} />}
      >
        <form onSubmit={e => void handleSubmit(onEdit)(e)}>
          <DialogContent padded className="space-y-4">
            <Field
              label={t("riskAnalysisScopeActions.fields.name")}
              {...register("name", { required: t("riskAnalysisScopeActions.validation.nameRequired") })}
              type="text"
              error={formState.errors.name?.message}
            />
          </DialogContent>
          <DialogFooter>
            <Button type="submit">{t("riskAnalysisScopeActions.actions.save")}</Button>
          </DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
