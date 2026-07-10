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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
  compliancePageId: string;
}>;

export function DeleteCompliancePageDomainDialog(props: DeleteCompliancePageDomainDialogProps) {
  const { children, domain, customDomainId } = props;

  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [inputValue, setInputValue] = useState("");

  const [deleteCustomDomain, isDeleting]
    = useMutation<DeleteCompliancePageDomainDialogMutation>(
      deleteCustomDomainMutation,
      {
        successMessage: __("Domain deleted successfully"),
        errorToast: __("Failed to delete domain"),
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
      title={__("Delete Custom Domain")}
    >
      <DialogContent padded className="space-y-4">
        <p className="text-txt-secondary text-sm">
          {sprintf(
            __(
              "This will permanently delete the custom domain %s and all its configuration.",
            ),
            domain,
          )}
        </p>

        <p className="text-red-600 text-sm font-medium">
          {__("This action cannot be undone.")}
        </p>

        <Field
          label={sprintf(__("To confirm deletion, type \"%s\" below:"), domain)}
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
          {isDeleting ? __("Deleting...") : __("Delete Domain")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
