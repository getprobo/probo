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
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { DeleteWebhookSubscriptionDialogMutation } from "#/__generated__/core/DeleteWebhookSubscriptionDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { deleteWebhookSubscriptionDialog } from "../variants";

const deleteWebhookSubscriptionMutation = graphql`
  mutation DeleteWebhookSubscriptionDialogMutation(
    $input: DeleteWebhookSubscriptionInput!
  ) {
    deleteWebhookSubscription(input: $input) {
      deletedWebhookSubscriptionId
    }
  }
`;

interface DeleteWebhookSubscriptionDialogProps {
  webhookSubscriptionId: string;
  onDeleted: () => void;
  trigger?: "icon" | "button";
}

export function DeleteWebhookSubscriptionDialog({
  webhookSubscriptionId,
  onDeleted,
  trigger = "icon",
}: DeleteWebhookSubscriptionDialogProps) {
  const { t } = useTranslation();
  const { body } = deleteWebhookSubscriptionDialog();
  const [open, setOpen] = useState(false);
  const [deleteWebhook, isDeleting] = useMutation<DeleteWebhookSubscriptionDialogMutation>(
    deleteWebhookSubscriptionMutation,
    {
      successMessage: t("webhooksSettingsPage.messages.deleted"),
      errorToast: t("webhooksSettingsPage.errors.delete"),
    },
  );

  function handleDelete() {
    void deleteWebhook({
      variables: {
        input: {
          webhookSubscriptionId,
        },
      },
    }).then(
      () => {
        setOpen(false);
        onDeleted();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={
          trigger === "button"
            ? (
                <Button variant="surface" color="red" iconStart={<TrashIcon />}>
                  {t("webhooksSettingsPage.deleteWebhook")}
                </Button>
              )
            : (
                <IconButton
                  variant="surface"
                  color="red"
                  size={1}
                  aria-label={t("webhooksSettingsPage.deleteWebhook")}
                >
                  <TrashIcon />
                </IconButton>
              )
        }
      />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("webhooksSettingsPage.dialogs.deleteTitle")}</DialogTitle>
          <DialogDescription>
            {t("webhooksSettingsPage.deleteConfirmation")}
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <div className={body()}>
            <Text size={2} color="neutral">
              {t("webhooksSettingsPage.cannotUndo")}
            </Text>
          </div>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("webhooksSettingsPage.actions.cancel")}
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
            {t("webhooksSettingsPage.actions.delete")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
