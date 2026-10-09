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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AccessReviewSourceDeleteDialog_source$key } from "#/__generated__/core/AccessReviewSourceDeleteDialog_source.graphql";
import type { AccessReviewSourceDeleteDialogMutation } from "#/__generated__/core/AccessReviewSourceDeleteDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const accessReviewSourceDeleteDialogFragment = graphql`
  fragment AccessReviewSourceDeleteDialog_source on AccessReviewSource {
    id
    name
    canDelete: permission(action: "access-review:source:delete")
  }
`;

const deleteAccessReviewSourceMutation = graphql`
  mutation AccessReviewSourceDeleteDialogMutation(
    $input: DeleteAccessReviewSourceInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewSource(input: $input) {
      deletedAccessReviewSourceId @deleteEdge(connections: $connections)
    }
  }
`;

interface AccessReviewSourceDeleteDialogProps {
  sourceKey: AccessReviewSourceDeleteDialog_source$key;
  connectionId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AccessReviewSourceDeleteDialog({
  sourceKey,
  connectionId,
  open,
  onOpenChange,
}: AccessReviewSourceDeleteDialogProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const source = useFragment(accessReviewSourceDeleteDialogFragment, sourceKey);
  const [deleteAccessReviewSource, isDeleting]
    = useMutation<AccessReviewSourceDeleteDialogMutation>(
      deleteAccessReviewSourceMutation,
      { errorToast: t("accessReviewSourceRow.errors.delete") },
    );

  if (!source.canDelete) {
    return null;
  }

  function handleDelete() {
    void deleteAccessReviewSource({
      variables: {
        input: { accessReviewSourceId: source.id },
        connections: [connectionId],
      },
    }).then(
      () => {
        toast.add({
          title: t("accessReviewSourceRow.messages.deleted"),
          type: "success",
        });
        onOpenChange(false);
      },
      () => {
        onOpenChange(false);
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("accessReviewSourceRow.actions.delete")}</DialogTitle>
          <DialogDescription>
            {t("accessReviewSourceRow.deleteConfirmation", { name: source.name })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("accessReviewSourceRow.actions.cancel")}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color="red"
            iconStart={<TrashIcon />}
            loading={isDeleting}
            onClick={handleDelete}
          >
            {t("accessReviewSourceRow.actions.delete")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
