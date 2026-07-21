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

import { formatError } from "@probo/helpers";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";

import type { DeleteBusinessAssociateAgreementDialogMutation } from "#/__generated__/core/DeleteBusinessAssociateAgreementDialogMutation.graphql";

const deleteBusinessAssociateAgreementMutation = graphql`
  mutation DeleteBusinessAssociateAgreementDialogMutation(
    $input: DeleteThirdPartyBusinessAssociateAgreementInput!
  ) {
    deleteThirdPartyBusinessAssociateAgreement(input: $input) {
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

export function DeleteBusinessAssociateAgreementDialog({
  children,
  thirdPartyId,
  fileName,
  onSuccess,
}: Props) {
  const { t } = useTranslation();
  const ref = useDialogRef();

  const { toast } = useToast();
  const [deleteAgreement, isDeleting]
    = useMutation<DeleteBusinessAssociateAgreementDialogMutation>(
      deleteBusinessAssociateAgreementMutation,
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
            title: t("deleteBusinessAssociateAgreementDialog.messages.error"),
            description: formatError(
              t("deleteBusinessAssociateAgreementDialog.errors.delete"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("deleteBusinessAssociateAgreementDialog.messages.success"),
          description: t("deleteBusinessAssociateAgreementDialog.messages.deleted"),
          variant: "success",
        });
        onSuccess?.();
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: t("deleteBusinessAssociateAgreementDialog.messages.error"),
          description: formatError(
            t("deleteBusinessAssociateAgreementDialog.errors.delete"),
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
      title={t("deleteBusinessAssociateAgreementDialog.title")}
      className="max-w-md"
    >
      <DialogContent padded>
        <p className="text-txt-secondary">
          {t("deleteBusinessAssociateAgreementDialog.confirmation", { fileName })}
        </p>
        <p className="text-txt-secondary mt-2">
          {t("deleteBusinessAssociateAgreementDialog.cannotUndo")}
        </p>
      </DialogContent>

      <DialogFooter>
        <Button
          variant="danger"
          onClick={handleDelete}
          disabled={isDeleting}
          icon={isDeleting ? Spinner : undefined}
        >
          {t("deleteBusinessAssociateAgreementDialog.actions.delete")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
