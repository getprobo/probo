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
  Dropzone,
  Field,
  Input,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";
import { z } from "zod";

import type { UploadDataPrivacyAgreementDialogMutation } from "#/__generated__/core/UploadDataPrivacyAgreementDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const uploadDataPrivacyAgreementMutation = graphql`
  mutation UploadDataPrivacyAgreementDialogMutation(
    $input: UploadThirdPartyDataPrivacyAgreementInput!
  ) {
    uploadThirdPartyDataPrivacyAgreement(input: $input) {
      thirdPartyDataPrivacyAgreement {
        id
        file {
          fileName
          downloadUrl
        }
        validFrom
        validUntil
        createdAt
      }
    }
  }
`;

type Props = {
  children: React.ReactNode;
  thirdPartyId: string;
  onSuccess?: () => void;
};

export function UploadDataPrivacyAgreementDialog({
  children,
  thirdPartyId,
  onSuccess,
}: Props) {
  const { t } = useTranslation();
  const schema = z.object({
    fileName: z.string().min(1, t("uploadDataPrivacyAgreementDialog.validation.fileNameRequired")),
    validFrom: z.string().optional(),
    validUntil: z.string().optional(),
  });
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const ref = useDialogRef();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useFormWithSchema(schema, {
    defaultValues: {
      fileName: "",
      validFrom: "",
      validUntil: "",
    },
  });

  const { toast } = useToast();
  const [uploadAgreement, isUploading]
    = useMutation<UploadDataPrivacyAgreementDialogMutation>(
      uploadDataPrivacyAgreementMutation,
    );

  const handleDrop = (files: File[]) => {
    if (files.length > 0) {
      const file = files[0];
      setUploadedFile(file);
      setValue("fileName", file.name);
    }
  };

  const onSubmit = (data: z.infer<typeof schema>) => {
    if (!uploadedFile) {
      return;
    }

    const formatDatetime = (dateString?: string) => {
      if (!dateString) return null;
      return `${dateString}T00:00:00Z`;
    };

    uploadAgreement({
      variables: {
        input: {
          thirdPartyId,
          fileName: data.fileName,
          validFrom: formatDatetime(data.validFrom),
          validUntil: formatDatetime(data.validUntil),
          file: null,
        },
      },
      uploadables: {
        "input.file": uploadedFile,
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: t("uploadDataPrivacyAgreementDialog.messages.error"),
            description: formatError(
              t("uploadDataPrivacyAgreementDialog.errors.upload"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("uploadDataPrivacyAgreementDialog.messages.success"),
          description: t("uploadDataPrivacyAgreementDialog.messages.uploaded"),
          variant: "success",
        });
        reset();
        setUploadedFile(null);
        onSuccess?.();
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: t("uploadDataPrivacyAgreementDialog.messages.error"),
          description: formatError(
            t("uploadDataPrivacyAgreementDialog.errors.upload"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleClose = () => {
    reset();
    setUploadedFile(null);
  };

  return (
    <Dialog
      title={t("uploadDataPrivacyAgreementDialog.title")}
      ref={ref}
      trigger={children}
      className="max-w-lg"
      onClose={handleClose}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Dropzone
            description={t("uploadDataPrivacyAgreementDialog.fileHelp")}
            isUploading={isUploading}
            onDrop={handleDrop}
            accept={{
              "application/pdf": [".pdf"],
            }}
            maxSize={10}
          />

          {uploadedFile && (
            <div className="p-3 bg-tertiary-subtle rounded-lg">
              <p className="text-sm font-medium">
                {t("uploadDataPrivacyAgreementDialog.selectedFile")}
              </p>
              <p className="text-sm text-txt-secondary">{uploadedFile.name}</p>
            </div>
          )}

          <Field
            {...register("fileName")}
            label={t("uploadDataPrivacyAgreementDialog.fields.fileName")}
            type="text"
            required
            error={errors.fileName?.message}
            placeholder={t("uploadDataPrivacyAgreementDialog.placeholders.fileName")}
          />

          <div className="grid grid-cols-2 gap-4">
            <Field label={t("uploadDataPrivacyAgreementDialog.fields.validFrom")}>
              <Input {...register("validFrom")} type="date" />
            </Field>
            <Field label={t("uploadDataPrivacyAgreementDialog.fields.validUntil")}>
              <Input {...register("validUntil")} type="date" />
            </Field>
          </div>
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            disabled={isUploading || !uploadedFile}
            icon={isUploading ? Spinner : undefined}
          >
            {t("uploadDataPrivacyAgreementDialog.actions.upload")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
