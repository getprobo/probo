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
} from "@probo/ui";
import { graphql } from "react-relay";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateDataPrivacyAgreementMutation = graphql`
  mutation EditDataPrivacyAgreementDialogMutation(
    $input: UpdateThirdPartyDataPrivacyAgreementInput!
  ) {
    updateThirdPartyDataPrivacyAgreement(input: $input) {
      thirdPartyDataPrivacyAgreement {
        id
        fileUrl
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

export function EditDataPrivacyAgreementDialog({
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
    formState: { isSubmitting },
    reset,
  } = useFormWithSchema(schema, {
    defaultValues: {
      validFrom: formatDateForForm(agreement.validFrom),
      validUntil: formatDateForForm(agreement.validUntil),
    },
  });

  const [mutate] = useMutationWithToasts(updateDataPrivacyAgreementMutation, {
    successMessage: __("Data Privacy Agreement updated successfully"),
    errorMessage: __("Failed to update Data Privacy Agreement"),
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const formatDatetime = (dateString?: string) => {
      if (!dateString) return null;
      return `${dateString}T00:00:00Z`;
    };

    await mutate({
      variables: {
        input: {
          thirdPartyId,
          validFrom: formatDatetime(data.validFrom),
          validUntil: formatDatetime(data.validUntil),
        },
      },
    });

    onSuccess?.();
    ref.current?.close();
  };

  const handleClose = () => {
    reset();
  };

  return (
    <Dialog
      title={__("Edit Data Privacy Agreement")}
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
            disabled={isSubmitting}
            icon={isSubmitting ? Spinner : undefined}
          >
            {__("Update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
