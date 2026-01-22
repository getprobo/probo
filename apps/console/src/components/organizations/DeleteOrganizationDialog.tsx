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
import { useState } from "react";

type DeleteOrganizationDialogProps = {
  children: React.ReactNode;
  organizationName: string;
  onConfirm: () => void;
  isDeleting?: boolean;
};

export function DeleteOrganizationDialog({
  children,
  organizationName,
  onConfirm,
  isDeleting = false,
}: DeleteOrganizationDialogProps) {
  const { __ } = useTranslate();
  const [inputValue, setInputValue] = useState("");
  const dialogRef = useDialogRef();
  const isConfirmDisabled = inputValue !== organizationName || isDeleting;

  const handleConfirm = () => {
    if (inputValue === organizationName) {
      onConfirm();
      setInputValue("");
    }
  };

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={children}
      title={__("Delete Organization")}
    >
      <DialogContent padded className="space-y-4">
        <p className="text-txt-secondary text-sm">
          {sprintf(
            __("This will permanently delete the organization %s and all its data."),
            organizationName,
          )}
        </p>

        <p className="text-red-600 text-sm font-medium">
          {__("This action cannot be undone.")}
        </p>

        <Field
          label={sprintf(
            __("To confirm deletion, type \"%s\" below:"),
            organizationName,
          )}
          type="text"
          value={inputValue}
          onChange={e => setInputValue(e.target.value)}
          placeholder={organizationName}
          disabled={isDeleting}
          autoComplete="off"
          autoFocus
        />
      </DialogContent>
      <DialogFooter>
        <Button
          variant="danger"
          icon={IconTrashCan}
          onClick={handleConfirm}
          disabled={isConfirmDisabled}
        >
          {isDeleting ? __("Deleting...") : __("Delete Organization")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
