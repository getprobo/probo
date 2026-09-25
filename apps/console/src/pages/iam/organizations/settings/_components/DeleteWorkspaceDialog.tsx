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

import { Form } from "@base-ui/react/form";
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
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactElement, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { DeleteWorkspaceDialogMutation } from "#/__generated__/iam/DeleteWorkspaceDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { deleteWorkspaceDialog } from "../variants";

const SETTINGS_NS = "iam/organizations/settings";

const deleteOrganizationMutation = graphql`
  mutation DeleteWorkspaceDialogMutation($input: DeleteOrganizationInput!) {
    deleteOrganization(input: $input) {
      deletedOrganizationId
    }
  }
`;

interface DeleteWorkspaceDialogProps {
  organizationId: string;
  organizationName: string;
  children: ReactElement;
}

export function DeleteWorkspaceDialog({
  organizationId,
  organizationName,
  children,
}: DeleteWorkspaceDialogProps) {
  const { t } = useTranslation(SETTINGS_NS);
  const { form, fields } = deleteWorkspaceDialog();
  const [open, setOpen] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const [deleteOrganization, isDeleting] = useMutation<DeleteWorkspaceDialogMutation>(
    deleteOrganizationMutation,
    {
      successMessage: t("deleteWorkspaceDialog.messages.deleted"),
      errorToast: t("deleteWorkspaceDialog.errors.delete"),
    },
  );

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setConfirmation("");
      setErrors({});
    }
  }

  function handleSubmit() {
    if (confirmation !== organizationName) {
      setErrors({ confirmation: t("deleteWorkspaceDialog.confirmationMismatch") });
      return;
    }

    void deleteOrganization({
      variables: {
        input: { organizationId },
      },
    }).then(
      () => {
        window.location.replace(new URL("/", window.location.origin));
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t("deleteWorkspaceDialog.title")}</DialogTitle>
            <DialogDescription>
              {t("deleteWorkspaceDialog.description", { organizationName })}
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <div className={fields()}>
              <Text size={2} color="red" weight="medium">
                {t("deleteWorkspaceDialog.warning")}
              </Text>
              <Field
                label={t("deleteWorkspaceDialog.confirmationLabel", { organizationName })}
                error={errors.confirmation}
              >
                <TextField
                  name="confirmation"
                  required
                  value={confirmation}
                  placeholder={organizationName}
                  disabled={isDeleting}
                  autoComplete="off"
                  autoFocus
                  onValueChange={(value) => {
                    setConfirmation(value);
                    setErrors({});
                  }}
                />
              </Field>
            </div>
          </DialogBody>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral">
                  {t("deleteWorkspaceDialog.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="submit"
              variant="solid"
              color="red"
              iconStart={<TrashIcon />}
              loading={isDeleting}
              disabled={confirmation !== organizationName}
            >
              {t("deleteWorkspaceDialog.actions.delete")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}
