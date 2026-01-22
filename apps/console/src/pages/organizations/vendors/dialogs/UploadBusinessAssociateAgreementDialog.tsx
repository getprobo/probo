import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Dropzone,
  Field,
  Input,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { useState } from "react";
import { graphql } from "react-relay";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const uploadBusinessAssociateAgreementMutation = graphql`
  mutation UploadBusinessAssociateAgreementDialogMutation(
    $input: UploadVendorBusinessAssociateAgreementInput!
  ) {
    uploadVendorBusinessAssociateAgreement(input: $input) {
      vendorBusinessAssociateAgreement {
        id
        fileName
        fileUrl
        validFrom
        validUntil
        createdAt
      }
    }
  }
`;

const schema = z.object({
  fileName: z.string().min(1, "File name is required"),
  validFrom: z.string().optional(),
  validUntil: z.string().optional(),
});

type Props = {
  children: React.ReactNode;
  vendorId: string;
  onSuccess?: () => void;
};

export function UploadBusinessAssociateAgreementDialog({
  children,
  vendorId,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const ref = useDialogRef();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
    setValue,
  } = useFormWithSchema(schema, {
    defaultValues: {
      fileName: "",
      validFrom: "",
      validUntil: "",
    },
  });

  const [mutate] = useMutationWithToasts(uploadBusinessAssociateAgreementMutation, {
    successMessage: __("Business Associate Agreement uploaded successfully"),
    errorMessage: __("Failed to upload Business Associate Agreement"),
  });

  const handleDrop = (files: File[]) => {
    if (files.length > 0) {
      const file = files[0];
      setUploadedFile(file);
      setValue("fileName", file.name);
    }
  };

  const onSubmit = async (data: z.infer<typeof schema>) => {
    if (!uploadedFile) {
      return;
    }

    const formatDatetime = (dateString?: string) => {
      if (!dateString) return null;
      return `${dateString}T00:00:00Z`;
    };

    await mutate({
      variables: {
        input: {
          vendorId,
          fileName: data.fileName,
          validFrom: formatDatetime(data.validFrom),
          validUntil: formatDatetime(data.validUntil),
          file: null,
        },
      },
      uploadables: {
        "input.file": uploadedFile,
      },
    });

    reset();
    setUploadedFile(null);
    onSuccess?.();
    ref.current?.close();
  };

  const handleClose = () => {
    reset();
    setUploadedFile(null);
  };

  return (
    <Dialog
      title={__("Upload Business Associate Agreement")}
      ref={ref}
      trigger={children}
      className="max-w-lg"
      onClose={handleClose}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Dropzone
            description={__("Only PDF files up to 10MB are allowed")}
            isUploading={isSubmitting}
            onDrop={handleDrop}
            accept={{
              "application/pdf": [".pdf"],
            }}
            maxSize={10}
          />

          {uploadedFile && (
            <div className="p-3 bg-tertiary-subtle rounded-lg">
              <p className="text-sm font-medium">
                {__("Selected file")}
                :
              </p>
              <p className="text-sm text-txt-secondary">{uploadedFile.name}</p>
            </div>
          )}

          <Field
            {...register("fileName")}
            label={__("File name")}
            type="text"
            required
            error={errors.fileName?.message}
            placeholder={__("Business Associate Agreement")}
          />

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
            disabled={isSubmitting || !uploadedFile}
            icon={isSubmitting ? Spinner : undefined}
          >
            {__("Upload")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
