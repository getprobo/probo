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
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconTrashCan,
  useDialogRef,
} from "@probo/ui";
import { type PropsWithChildren, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { DeleteCompliancePageDomainDialogMutation } from "#/__generated__/core/DeleteCompliancePageDomainDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const deleteCustomDomainMutation = graphql`
  mutation DeleteCompliancePageDomainDialogMutation($input: DeleteCustomDomainInput!) {
    deleteCustomDomain(input: $input) {
      deletedCustomDomainId
    }
  }
`;

type DeleteCompliancePageDomainDialogProps = PropsWithChildren<{
  domain: string;
  customDomainId: string;
}>;

export function DeleteCompliancePageDomainDialog(props: DeleteCompliancePageDomainDialogProps) {
  const { children, domain, customDomainId } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const dialogRef = useDialogRef();
  const [inputValue, setInputValue] = useState("");

  const [deleteCustomDomain, isDeleting]
    = useMutation<DeleteCompliancePageDomainDialogMutation>(
      deleteCustomDomainMutation,
      {
        successMessage: t("deleteDomainDialog.messages.deleted"),
        errorToast: t("deleteDomainDialog.errors.delete"),
      },
    );

  const handleDeleteDomain = async () => {
    return deleteCustomDomain({
      variables: {
        input: { customDomainId },
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
      updater: (store) => {
        store.delete(customDomainId);
      },
    });
  };

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={children}
      title={t("deleteDomainDialog.title")}
    >
      <DialogContent padded className="space-y-4">
        <p className="text-txt-secondary text-sm">
          {t("deleteDomainDialog.description", { domain })}
        </p>

        <p className="text-red-600 text-sm font-medium">
          {t("deleteDomainDialog.warning")}
        </p>

        <Field
          label={t("deleteDomainDialog.confirmationLabel", { domain })}
          type="text"
          value={inputValue}
          onChange={e => setInputValue(e.target.value)}
          placeholder={domain}
          disabled={isDeleting}
          autoComplete="off"
          autoFocus
        />
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          icon={IconTrashCan}
          onClick={() => void handleDeleteDomain()}
          disabled={isDeleting || inputValue !== domain}
        >
          {isDeleting ? t("deleteDomainDialog.actions.deleting") : t("deleteDomainDialog.actions.delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
