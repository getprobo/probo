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

import { formatDatetime, formatError, todayAsDateInput } from "@probo/helpers";
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
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { graphql, useMutation } from "react-relay";
import { z } from "zod";

import type { UploadComplianceReportDialogMutation } from "#/__generated__/core/UploadComplianceReportDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

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
    reset,
  } = useFormWithSchema(schema, {
    defaultValues: {
      reportDate: todayAsDateInput(),
      validUntil: "",
    },
  });

  const { toast } = useToast();
  const [uploadComplianceReport, isUploading]
    = useMutation<UploadComplianceReportDialogMutation>(
      uploadComplianceReportMutation,
    );

  const handleDrop = (files: File[]) => {
    if (files.length > 0) {
      setUploadedFile(files[0]);
    }
  };

  const onSubmit = (data: z.infer<typeof schema>) => {
    if (!uploadedFile) {
      return;
    }

    uploadComplianceReport({
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
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to upload compliance report"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Compliance report uploaded successfully"),
          variant: "success",
        });
        reset();
        setUploadedFile(null);
        onSuccess?.();
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to upload compliance report"),
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
            disabled={isUploading || !uploadedFile}
            icon={isUploading ? Spinner : undefined}
          >
            {__("Upload")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
