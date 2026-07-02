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
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { graphql, useMutation } from "react-relay";
import { z } from "zod";

import type { EditBusinessAssociateAgreementDialogMutation } from "#/__generated__/core/EditBusinessAssociateAgreementDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const updateBusinessAssociateAgreementMutation = graphql`
  mutation EditBusinessAssociateAgreementDialogMutation(
    $input: UpdateThirdPartyBusinessAssociateAgreementInput!
  ) {
    updateThirdPartyBusinessAssociateAgreement(input: $input) {
      thirdPartyBusinessAssociateAgreement {
        id
        file {
          downloadUrl
        }
        validFrom
        validUntil
        createdAt
      }
    }
  }
`;

const schema = z.object({
  validFrom: z.string().optional(),
  validUntil: z.string().optional(),
});

type Props = {
  children: React.ReactNode;
  thirdPartyId: string;
  agreement: {
    validFrom?: string | null;
    validUntil?: string | null;
  };
  onSuccess?: () => void;
};

export function EditBusinessAssociateAgreementDialog({
  children,
  thirdPartyId,
  agreement,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const ref = useDialogRef();

  const formatDateForForm = (datetime?: string | null) => {
    if (!datetime) return "";
    return datetime.split("T")[0];
  };

  const {
    register,
    handleSubmit,
    reset,
  } = useFormWithSchema(schema, {
    defaultValues: {
      validFrom: formatDateForForm(agreement.validFrom),
      validUntil: formatDateForForm(agreement.validUntil),
    },
  });

  const { toast } = useToast();
  const [updateAgreement, isUpdating]
    = useMutation<EditBusinessAssociateAgreementDialogMutation>(
      updateBusinessAssociateAgreementMutation,
    );

  const onSubmit = (data: z.infer<typeof schema>) => {
    const formatDatetime = (dateString?: string) => {
      if (!dateString) return null;
      return `${dateString}T00:00:00Z`;
    };

    updateAgreement({
      variables: {
        input: {
          thirdPartyId,
          validFrom: formatDatetime(data.validFrom),
          validUntil: formatDatetime(data.validUntil),
        },
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to update Business Associate Agreement"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Business Associate Agreement updated successfully"),
          variant: "success",
        });
        onSuccess?.();
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update Business Associate Agreement"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleClose = () => {
    reset();
  };

  return (
    <Dialog
      title={__("Edit Business Associate Agreement")}
      ref={ref}
      trigger={children}
      className="max-w-lg"
      onClose={handleClose}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Field label={__("Valid from")}>
              <Input {...register("validFrom")} type="date" />
            </Field>
            <Field label={__("Valid until")}>
              <Input {...register("validUntil")} type="date" />
            </Field>
          </div>
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            disabled={isUpdating}
            icon={isUpdating ? Spinner : undefined}
          >
            {__("Update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
