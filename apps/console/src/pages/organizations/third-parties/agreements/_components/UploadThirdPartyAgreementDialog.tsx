// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
  Option,
  Select,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";

import type { UploadThirdPartyAgreementDialogMutation } from "#/__generated__/core/UploadThirdPartyAgreementDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { z } from "#/lib/zod";

import { THIRD_PARTY_AGREEMENT_TYPES } from "./agreementTypes";

const uploadAgreementMutation = graphql`
  mutation UploadThirdPartyAgreementDialogMutation(
    $input: UploadThirdPartyAgreementInput!
    $connections: [ID!]!
  ) {
    uploadThirdPartyAgreement(input: $input) {
      thirdPartyAgreementEdge @prependEdge(connections: $connections) {
        node {
          ...ThirdPartyAgreementListItem_agreement
        }
      }
    }
  }
`;

type Props = {
  children: React.ReactNode;
  thirdPartyId: string;
  connectionId: string;
};

export function UploadThirdPartyAgreementDialog({
  children,
  thirdPartyId,
  connectionId,
}: Props) {
  const { t } = useTranslation();
  const schema = z.object({
    type: z.enum(THIRD_PARTY_AGREEMENT_TYPES),
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
    control,
  } = useFormWithSchema(schema, {
    defaultValues: {
      type: "OTHER",
      validFrom: "",
      validUntil: "",
    },
  });

  const toast = useToast();
  const [uploadAgreement, isUploading]
    = useMutation<UploadThirdPartyAgreementDialogMutation>(uploadAgreementMutation);

  const handleDrop = (files: File[]) => {
    if (files.length > 0) {
      setUploadedFile(files[0]);
    }
  };

  const onSubmit = (data: z.infer<typeof schema>) => {
    if (!uploadedFile) {
      return;
    }

    uploadAgreement({
      variables: {
        input: {
          thirdPartyId,
          type: data.type,
          validity: {
            start: formatDatetime(data.validFrom),
            end: formatDatetime(data.validUntil),
          },
          file: null,
        },
        connections: [connectionId],
      },
      uploadables: {
        "input.file": uploadedFile,
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast.add({
            title: t("uploadThirdPartyAgreementDialog.messages.error"),
            description: formatError(
              t("uploadThirdPartyAgreementDialog.errors.upload"),
              errors,
            ),
            type: "error",
          });
          return;
        }
        toast.add({
          title: t("uploadThirdPartyAgreementDialog.messages.success"),
          description: t("uploadThirdPartyAgreementDialog.messages.uploaded"),
          type: "success",
        });
        reset();
        setUploadedFile(null);
        ref.current?.close();
      },
      onError(error) {
        toast.add({
          title: t("uploadThirdPartyAgreementDialog.messages.error"),
          description: formatError(
            t("uploadThirdPartyAgreementDialog.errors.upload"),
            error,
          ),
          type: "error",
        });
      },
    });
  };

  return (
    <Dialog
      title={t("uploadThirdPartyAgreementDialog.title")}
      ref={ref}
      trigger={children}
      className="max-w-lg"
      onClose={() => {
        reset();
        setUploadedFile(null);
      }}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Dropzone
            description={t("uploadThirdPartyAgreementDialog.fileHelp")}
            isUploading={isUploading}
            onDrop={handleDrop}
            accept={{
              "application/pdf": [".pdf"],
            }}
            maxSize={30}
          />

          {uploadedFile && (
            <div className="p-3 bg-tertiary-subtle rounded-lg">
              <p className="text-sm font-medium">
                {t("uploadThirdPartyAgreementDialog.selectedFile")}
              </p>
              <p className="text-sm text-txt-secondary">{uploadedFile.name}</p>
            </div>
          )}

          <Controller
            control={control}
            name="type"
            render={({ field }) => (
              <Field
                label={t("uploadThirdPartyAgreementDialog.fields.type")}
                required
                error={errors.type?.message}
              >
                <Select
                  variant="editor"
                  value={field.value}
                  onValueChange={field.onChange}
                  className="w-full"
                >
                  {THIRD_PARTY_AGREEMENT_TYPES.map(type => (
                    <Option key={type} value={type}>
                      {t(`thirdPartyAgreementsPage.types.${type}`)}
                    </Option>
                  ))}
                </Select>
              </Field>
            )}
          />

          <div className="grid grid-cols-2 gap-4">
            <Field label={t("uploadThirdPartyAgreementDialog.fields.validFrom")}>
              <Input {...register("validFrom")} type="date" />
            </Field>
            <Field label={t("uploadThirdPartyAgreementDialog.fields.validUntil")}>
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
            {t("uploadThirdPartyAgreementDialog.actions.upload")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function formatDatetime(dateString?: string) {
  if (!dateString) {
    return null;
  }

  return `${dateString}T00:00:00Z`;
}
