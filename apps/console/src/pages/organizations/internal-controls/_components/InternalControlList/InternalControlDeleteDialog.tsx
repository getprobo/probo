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

import { TrashIcon } from "@phosphor-icons/react";
import { ConfirmDialog } from "@probo/ui/src/v2/Dialog/ConfirmDialog";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { InternalControlDeleteDialogMutation } from "#/__generated__/core/InternalControlDeleteDialogMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const deleteInternalControlMutation = graphql`
  mutation InternalControlDeleteDialogMutation($input: DeleteInternalControlInput!) {
    deleteInternalControl(input: $input) {
      deletedInternalControlId
    }
  }
`;

interface InternalControlDeleteDialogProps {
  open: boolean;
  internalControlId: string;
  name: string;
  onOpenChange: (open: boolean) => void;
  onDeleted: () => void;
}

export function InternalControlDeleteDialog(props: InternalControlDeleteDialogProps) {
  const { open, internalControlId, name, onOpenChange, onDeleted } = props;
  const { t } = useTranslation("organizations/internal-controls");
  const [deleteInternalControl, isDeleting]
    = useMutationWithToasts<InternalControlDeleteDialogMutation>(
      deleteInternalControlMutation,
      {
        successMessage: t("list.messages.deleted"),
        errorMessage: t("list.errors.delete"),
      },
    );

  return (
    <ConfirmDialog
      open={open}
      title={t("list.actions.delete")}
      description={t("list.deleteConfirmation", { name })}
      cancelLabel={t("list.actions.cancel")}
      confirmLabel={t("list.actions.delete")}
      confirmIcon={<TrashIcon />}
      loading={isDeleting}
      onOpenChange={onOpenChange}
      onConfirm={() => {
        void deleteInternalControl({
          variables: { input: { internalControlId } },
          onCompleted(_response, errors) {
            if (!errors) {
              onOpenChange(false);
              onDeleted();
            }
          },
        });
      }}
    />
  );
}
