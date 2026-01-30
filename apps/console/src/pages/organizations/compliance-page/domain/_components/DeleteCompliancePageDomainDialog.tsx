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
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const deleteCustomDomainMutation = graphql`
  mutation DeleteCompliancePageDomainDialogMutation($input: DeleteCustomDomainInput!) {
    deleteCustomDomain(input: $input) {
      deletedCustomDomainId
    }
  }
`;

type DeleteCompliancePageDomainDialogProps = PropsWithChildren<{
  domain: string;
}>;

export function DeleteCompliancePageDomainDialog(props: DeleteCompliancePageDomainDialogProps) {
  const { children, domain } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [inputValue, setInputValue] = useState("");

  const [deleteCustomDomain, isDeleting]
    = useMutationWithToasts<DeleteCompliancePageDomainDialogMutation>(
      deleteCustomDomainMutation,
      {
        successMessage: __("Domain deleted successfully"),
        errorMessage: __("Failed to delete domain"),
      },
    );

  const handleDeleteDomain = async () => {
    return deleteCustomDomain({
      variables: {
        input: { organizationId },
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
      updater: (store) => {
        // Update the cache by setting customDomain to null
        const organizationRecord = store.get(organizationId);
        if (organizationRecord) {
          organizationRecord.setValue(null, "customDomain");
        }
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
