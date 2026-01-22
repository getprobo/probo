import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconTrashCan,
  Spinner,
  useDialogRef,
} from "@probo/ui";

import { deleteTrustCenterReferenceMutation } from "#/hooks/graph/TrustCenterReferenceGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

type Props = {
  children: React.ReactNode;
  referenceId: string;
  referenceName: string;
  connectionId: string;
  onSuccess?: () => void;
};

export function DeleteTrustCenterReferenceDialog({
  children,
  referenceId,
  referenceName,
  connectionId,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const ref = useDialogRef();

  const [mutate, isDeleting] = useMutationWithToasts(deleteTrustCenterReferenceMutation, {
    successMessage: __("Reference deleted successfully"),
    errorMessage: __("Failed to delete reference"),
  });

  const handleDelete = async () => {
    await mutate({
      variables: {
        input: {
          id: referenceId,
        },
        connections: [connectionId],
      },
    });

    onSuccess?.();
    ref.current?.close();
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={__("Delete Reference")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {sprintf(
            __("Are you sure you want to delete the reference \"%s\"?"),
            referenceName,
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
          icon={isDeleting ? Spinner : IconTrashCan}
        >
          {isDeleting ? __("Deleting...") : __("Delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
