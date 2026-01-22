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

const updateBusinessAssociateAgreementMutation = graphql`
  mutation EditBusinessAssociateAgreementDialogMutation(
    $input: UpdateVendorBusinessAssociateAgreementInput!
  ) {
    updateVendorBusinessAssociateAgreement(input: $input) {
      vendorBusinessAssociateAgreement {
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
  vendorId: string;
  agreement: {
    validFrom?: string | null;
    validUntil?: string | null;
  };
  onSuccess?: () => void;
};

export function EditBusinessAssociateAgreementDialog({
  children,
  vendorId,
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

  const [mutate] = useMutationWithToasts(updateBusinessAssociateAgreementMutation, {
    successMessage: __("Business Associate Agreement updated successfully"),
    errorMessage: __("Failed to update Business Associate Agreement"),
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const formatDatetime = (dateString?: string) => {
      if (!dateString) return null;
      return `${dateString}T00:00:00Z`;
    };

    await mutate({
      variables: {
        input: {
          vendorId,
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
