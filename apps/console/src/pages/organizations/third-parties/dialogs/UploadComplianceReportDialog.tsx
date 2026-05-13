// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatDatetime, todayAsDateInput } from "@probo/helpers";
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

const uploadComplianceReportMutation = graphql`
  mutation UploadComplianceReportDialogMutation(
    $input: UploadThirdPartyComplianceReportInput!
    $connections: [ID!]!
  ) {
    uploadThirdPartyComplianceReport(input: $input) {
      thirdPartyComplianceReportEdge @appendEdge(connections: $connections) {
        node {
          id
          reportName
          reportDate
          validUntil
          file {
            fileName
            mimeType
            size
            downloadUrl
          }
          canDelete: permission(action: "core:thirdParty-compliance-report:delete")
        }
      }
    }
  }
`;

const schema = z.object({
  reportDate: z.string().min(1, "Report date is required"),
  validUntil: z.string().optional(),
});

type Props = {
  children: React.ReactNode;
  thirdPartyId: string;
  connectionId: string;
  onSuccess?: () => void;
};

export function UploadComplianceReportDialog({
  children,
  thirdPartyId,
  connectionId,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const ref = useDialogRef();

  const {
    register,
    handleSubmit,
    formState: { isSubmitting },
    reset,
  } = useFormWithSchema(schema, {
    defaultValues: {
      reportDate: todayAsDateInput(),
      validUntil: "",
    },
  });

  const [mutate] = useMutationWithToasts(uploadComplianceReportMutation, {
    successMessage: __("Compliance report uploaded successfully"),
    errorMessage: __("Failed to upload compliance report"),
  });

  const handleDrop = (files: File[]) => {
    if (files.length > 0) {
      setUploadedFile(files[0]);
    }
  };

  const onSubmit = async (data: z.infer<typeof schema>) => {
    if (!uploadedFile) {
      return;
    }

    await mutate({
      variables: {
        connections: [connectionId],
        input: {
          thirdPartyId,
          reportName: uploadedFile.name,
          reportDate: `${data.reportDate}T00:00:00Z`,
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
      title={__("Upload Compliance Report")}
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

          <div className="grid grid-cols-2 gap-4">
            <Field label={__("Report date")} required>
              <Input {...register("reportDate")} type="date" required />
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
