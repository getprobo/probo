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
import { forwardRef, useImperativeHandle } from "react";
import { z } from "zod";
import { formatDatetime, toDateInput } from "@probo/helpers";

import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { updatePeopleMutation } from "/hooks/graph/PeopleGraph";

const schema = z.object({
  contractEndDate: z.string().optional(),
});

export type SetEndOfContractDialogRef = {
  open: () => void;
  close: () => void;
};

type Props = {
  peopleId: string;
  currentContractEndDate?: string | null;
};

export const SetEndOfContractDialog = forwardRef<SetEndOfContractDialogRef, Props>(function SetEndOfContractDialog(
  {
    peopleId,
    currentContractEndDate,
  },
  ref,
) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  useImperativeHandle(ref, () => ({
    open: () => dialogRef.current?.open(),
    close: () => dialogRef.current?.close(),
  }));

  const {
    register,
    handleSubmit,
    formState: { isSubmitting },
    reset,
  } = useFormWithSchema(schema, {
    defaultValues: {
      contractEndDate: toDateInput(currentContractEndDate),
    },
  });

  const [mutate] = useMutationWithToasts(updatePeopleMutation, {
    successMessage: __("End of contract updated successfully"),
    errorMessage: __("Failed to update end of contract"),
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    await mutate({
      variables: {
        input: {
          id: peopleId,
          contractEndDate: formatDatetime(data.contractEndDate),
        },
      },
    });

    dialogRef.current?.close();
  };

  const handleClose = () => {
    reset();
  };

  return (
    <Dialog
      title={__("Set End of Contract")}
      ref={dialogRef}
      className="max-w-lg"
      onClose={handleClose}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field label={__("End of contract")}>
            <Input {...register("contractEndDate")} type="date" />
          </Field>
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
});
