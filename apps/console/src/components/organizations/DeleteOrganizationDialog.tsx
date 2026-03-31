// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
