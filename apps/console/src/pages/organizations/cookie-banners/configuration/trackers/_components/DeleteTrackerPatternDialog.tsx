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
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { DeleteTrackerPatternDialogMutation } from "#/__generated__/core/DeleteTrackerPatternDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const deletePatternMutation = graphql`
  mutation DeleteTrackerPatternDialogMutation(
    $input: DeleteTrackerPatternInput!
  ) {
    deleteTrackerPattern(input: $input) {
      deletedTrackerPatternId
      cookieBanner {
        id
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface DeleteTrackerPatternDialogProps {
  trackerPatternId: string;
  displayName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRemoved: () => void;
}

export function DeleteTrackerPatternDialog({
  trackerPatternId,
  displayName,
  open,
  onOpenChange,
  onRemoved,
}: DeleteTrackerPatternDialogProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [deletePattern, isDeleting] = useMutation<DeleteTrackerPatternDialogMutation>(
    deletePatternMutation,
    {
      successMessage: t("trackerPatternRow.messages.cookieDeleted"),
      errorToast: t("trackerPatternRow.errors.deleteCookie"),
    },
  );

  function handleDelete() {
    void deletePattern({
      variables: {
        input: { trackerPatternId },
      },
    }).then(
      () => {
        onOpenChange(false);
        onRemoved();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("trackerPatternRow.dialogs.deleteTitle")}</DialogTitle>
          <DialogDescription>
            {t("trackerPatternRow.deleteConfirmation", { name: displayName })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("trackerPatternRowEdit.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="red"
            loading={isDeleting}
            iconStart={<TrashIcon />}
            onClick={handleDelete}
          >
            {t("trackerPatternRow.actions.delete")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
