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
  IconTrashCan,
  Spinner,
  useDialogRef,
} from "@probo/ui";

import type { TrustCenterReferenceGraphDeleteMutation } from "#/__generated__/core/TrustCenterReferenceGraphDeleteMutation.graphql";
import { deleteTrustCenterReferenceMutation } from "#/hooks/graph/TrustCenterReferenceGraph";
import { useMutation } from "#/lib/relay/useMutation";

type Props = {
  children: React.ReactNode;
  referenceId: string;
  referenceName: string;
  connectionId: string;
  onSuccess?: () => void;
};

export function DeleteCompliancePageReferenceDialog({
  children,
  referenceId,
  referenceName,
  connectionId,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const ref = useDialogRef();

  const [mutate, isDeleting] = useMutation<TrustCenterReferenceGraphDeleteMutation>(
    deleteTrustCenterReferenceMutation,
    {
      successMessage: __("Reference deleted successfully"),
      errorToast: __("Failed to delete reference"),
    },
  );

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
