import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { graphql } from "react-relay";

import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const deleteBusinessAssociateAgreementMutation = graphql`
  mutation DeleteBusinessAssociateAgreementDialogMutation(
    $input: DeleteVendorBusinessAssociateAgreementInput!
  ) {
    deleteVendorBusinessAssociateAgreement(input: $input) {
      deletedVendorId
    }
  }
`;

type Props = {
  children: React.ReactNode;
  vendorId: string;
  fileName: string;
  onSuccess?: () => void;
};

export function DeleteBusinessAssociateAgreementDialog({
  children,
  vendorId,
  fileName,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const ref = useDialogRef();

  const [mutate, isDeleting] = useMutationWithToasts(deleteBusinessAssociateAgreementMutation, {
    successMessage: __("Business Associate Agreement deleted successfully"),
    errorMessage: __("Failed to delete Business Associate Agreement"),
  });

  const handleDelete = async () => {
    await mutate({
      variables: {
        input: {
          vendorId,
        },
      },
    });

    onSuccess?.();
    ref.current?.close();
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={__("Delete Business Associate Agreement")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {sprintf(
            __("Are you sure you want to delete the Business Associate Agreement \"%s\"?"),
            fileName,
          )}
        </p>
        <p className="text-txt-secondary mt-2">
          {__("This action cannot be undone.")}
        </p>
      </DialogContent>

      <DialogFooter>
        <Button
          variant="danger"
          onClick={() => void handleDelete()}
          disabled={isDeleting}
          icon={isDeleting ? Spinner : undefined}
        >
          {__("Delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
