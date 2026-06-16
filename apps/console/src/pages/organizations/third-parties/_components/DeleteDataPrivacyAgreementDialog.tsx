// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { formatError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { graphql, useMutation } from "react-relay";

import type { DeleteDataPrivacyAgreementDialogMutation } from "#/__generated__/core/DeleteDataPrivacyAgreementDialogMutation.graphql";

const deleteDataPrivacyAgreementMutation = graphql`
  mutation DeleteDataPrivacyAgreementDialogMutation(
    $input: DeleteThirdPartyDataPrivacyAgreementInput!
  ) {
    deleteThirdPartyDataPrivacyAgreement(input: $input) {
      deletedThirdPartyId
    }
  }
`;

type Props = {
  children: React.ReactNode;
  thirdPartyId: string;
  fileName: string;
  onSuccess?: () => void;
};

export function DeleteDataPrivacyAgreementDialog({
  children,
  thirdPartyId,
  fileName,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const ref = useDialogRef();

  const { toast } = useToast();
  const [deleteAgreement, isDeleting]
    = useMutation<DeleteDataPrivacyAgreementDialogMutation>(
      deleteDataPrivacyAgreementMutation,
    );

  const handleDelete = () => {
    deleteAgreement({
      variables: {
        input: {
          thirdPartyId,
        },
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to delete Data Privacy Agreement"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Data Privacy Agreement deleted successfully"),
          variant: "success",
        });
        onSuccess?.();
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to delete Data Privacy Agreement"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={__("Delete Data Privacy Agreement")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {sprintf(
            __("Are you sure you want to delete the Data Privacy Agreement \"%s\"?"),
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
          onClick={handleDelete}
          disabled={isDeleting}
          icon={isDeleting ? Spinner : undefined}
        >
          {__("Delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
