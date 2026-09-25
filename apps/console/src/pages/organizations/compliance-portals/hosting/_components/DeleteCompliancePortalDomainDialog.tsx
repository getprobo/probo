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
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeleteCompliancePortalDomainDialog_customDomain$key } from "#/__generated__/core/DeleteCompliancePortalDomainDialog_customDomain.graphql";
import type { DeleteCompliancePortalDomainDialogMutation } from "#/__generated__/core/DeleteCompliancePortalDomainDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { deleteDomainDialog } from "../variants";

const deleteCustomDomainMutation = graphql`
  mutation DeleteCompliancePortalDomainDialogMutation($input: DeleteCustomDomainInput!) {
    deleteCustomDomain(input: $input) {
      deletedCustomDomainId
    }
  }
`;

const fragment = graphql`
  fragment DeleteCompliancePortalDomainDialog_customDomain on CustomDomain {
    id
    domain
  }
`;

interface DeleteCompliancePortalDomainDialogProps {
  customDomainKey: DeleteCompliancePortalDomainDialog_customDomain$key;
  children: ReactElement;
}

export function DeleteCompliancePortalDomainDialog({
  customDomainKey,
  children,
}: DeleteCompliancePortalDomainDialogProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { form, fields } = deleteDomainDialog();
  const domain = useFragment(fragment, customDomainKey);
  const [open, setOpen] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const [deleteCustomDomain, isDeleting]
    = useMutation<DeleteCompliancePortalDomainDialogMutation>(
      deleteCustomDomainMutation,
      {
        successMessage: t("deleteDomainDialog.messages.deleted"),
        errorToast: t("deleteDomainDialog.errors.delete"),
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
    if (confirmation !== domain.domain) {
      setErrors({ confirmation: t("deleteDomainDialog.confirmationMismatch") });
      return;
    }

    void deleteCustomDomain({
      variables: {
        input: { customDomainId: domain.id },
      },
      updater: (store) => {
        store.delete(domain.id);
      },
    }).then(
      () => {
        handleOpenChange(false);
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
            <DialogTitle>{t("deleteDomainDialog.title")}</DialogTitle>
            <DialogDescription>
              {t("deleteDomainDialog.description", { domain: domain.domain })}
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <div className={fields()}>
              <Text size={2} color="red" weight="medium">
                {t("deleteDomainDialog.warning")}
              </Text>
              <Field
                label={t("deleteDomainDialog.confirmationLabel", { domain: domain.domain })}
                error={errors.confirmation}
              >
                <TextField
                  name="confirmation"
                  required
                  value={confirmation}
                  placeholder={domain.domain}
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
                  {t("deleteDomainDialog.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="submit"
              variant="solid"
              color="red"
              iconStart={<TrashIcon />}
              loading={isDeleting}
              disabled={confirmation !== domain.domain}
            >
              {t("deleteDomainDialog.actions.delete")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}
