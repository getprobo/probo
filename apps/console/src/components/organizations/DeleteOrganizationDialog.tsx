// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
