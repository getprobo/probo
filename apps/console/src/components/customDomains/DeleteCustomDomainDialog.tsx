import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  IconTrashCan,
} from "@probo/ui";
import { useState, useEffect, type ReactNode } from "react";
import { sprintf } from "@probo/helpers";

interface DeleteCustomDomainDialogProps {
  children: ReactNode;
  domainName: string;
  onConfirm: () => Promise<void>;
}

export function DeleteCustomDomainDialog({
  children,
  domainName,
  onConfirm,
}: DeleteCustomDomainDialogProps) {
  const { __ } = useTranslate();
  const [inputValue, setInputValue] = useState("");
  const [isDeleting, setIsDeleting] = useState(false);
  const dialogRef = useDialogRef();

  const isConfirmDisabled = inputValue !== domainName || isDeleting;

  const handleConfirm = async () => {
    if (inputValue === domainName && !isDeleting) {
      setIsDeleting(true);
      try {
        await onConfirm();
        dialogRef.current?.close();
      } finally {
        setIsDeleting(false);
      }
    }
  };

  useEffect(() => {
    if (!isDeleting) {
      setInputValue("");
    }
  }, [isDeleting]);

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
              "This will permanently delete the custom domain %s and all its configuration."
            ),
            domainName
          )}
        </p>

        <p className="text-red-600 text-sm font-medium">
          {__("This action cannot be undone.")}
        </p>

        <Field
          label={sprintf(
            __('To confirm deletion, type "%s" below:'),
            domainName
          )}
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder={domainName}
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
          {isDeleting ? __("Deleting...") : __("Delete Domain")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
