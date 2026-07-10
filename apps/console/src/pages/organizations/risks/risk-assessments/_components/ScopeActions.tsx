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
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { graphql, useMutation } from "react-relay";

import type { ScopeActionsDeleteMutation } from "#/__generated__/core/ScopeActionsDeleteMutation.graphql";
import type { ScopeActionsUpdateMutation } from "#/__generated__/core/ScopeActionsUpdateMutation.graphql";

const updateScopeMutation = graphql`
  mutation ScopeActionsUpdateMutation(
    $input: UpdateRiskAssessmentScopeInput!
  ) {
    updateRiskAssessmentScope(input: $input) {
      riskAssessmentScope { id name }
    }
  }
`;

const deleteScopeMutation = graphql`
  mutation ScopeActionsDeleteMutation(
    $input: DeleteRiskAssessmentScopeInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessmentScope(input: $input) {
      deletedRiskAssessmentScopeId @deleteEdge(connections: $connections)
    }
  }
`;

export function ScopeActions(props: {
  scope: { id: string; name: string };
  connectionId: string;
}) {
  const { __ } = useTranslate();
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
            input: { riskAssessmentScopeId: props.scope.id },
            connections: [props.connectionId],
          },
        });
      },
      { message: __("Delete this scope and all its nodes, processes, and threats?") },
    );
  };

  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {__("Edit")}
        </DropdownItem>
        <DropdownItem icon={IconTrashCan} variant="danger" onSelect={onDelete}>
          {__("Delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog
        className="max-w-lg"
        ref={dialogRef}
        title={<Breadcrumb items={[__("Scopes"), __("Edit Scope")]} />}
      >
        <form onSubmit={e => void handleSubmit(onEdit)(e)}>
          <DialogContent padded className="space-y-4">
            <Field
              label={__("Name")}
              {...register("name", { required: __("This field is required") })}
              type="text"
              error={formState.errors.name?.message}
            />
          </DialogContent>
          <DialogFooter>
            <Button type="submit">{__("Save")}</Button>
          </DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
