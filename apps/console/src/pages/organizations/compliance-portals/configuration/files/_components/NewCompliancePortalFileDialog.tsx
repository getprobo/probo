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

import {
  acceptData,
  acceptDocument,
  acceptImage,
  acceptPresentation,
  acceptSpreadsheet,
  acceptText,
  getCompliancePortalVisibilityOptions,
} from "@probo/helpers";
import { Badge, Button, Dialog, DialogContent, DialogFooter, type DialogRef, Dropzone, Field, Option, Spinner } from "@probo/ui";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { type DataID, graphql } from "relay-runtime";

import type { NewCompliancePortalFileDialog_createMutation } from "#/__generated__/core/NewCompliancePortalFileDialog_createMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutation } from "#/lib/relay/useMutation";
import { z } from "#/lib/zod";

const acceptedFileTypes = {
  ...acceptDocument,
  ...acceptSpreadsheet,
  ...acceptPresentation,
  ...acceptText,
  ...acceptImage,
  ...acceptData,
};

const createCompliancePortalFileMutation = graphql`
  mutation NewCompliancePortalFileDialog_createMutation(
    $input: CreateCompliancePortalFileInput!
    $connections: [ID!]!
  ) {
    createCompliancePortalFile(input: $input) {
      compliancePortalFileEdge @prependEdge(connections: $connections) {
        node {
          ...CompliancePortalFileListItem_fileFragment
        }
      }
    }
  }
`;

export function NewCompliancePortalFileDialog(props: {
  connectionId: DataID;
  compliancePortalId: string;
  ref: DialogRef;
}) {
  const { connectionId, compliancePortalId, ref } = props;

  const { t } = useTranslation("organizations/compliance-portals");

  const [uploadedFile, setUploadedFile] = useState<File | null>(null);

  const createSchema = z.object({
    name: z.string().min(1, t("newFileDialog.validation.nameRequired")),
    category: z.string().min(1, t("newFileDialog.validation.categoryRequired")),
    compliancePortalVisibility: z.enum(["NONE", "RESTRICTED", "PUBLIC"]),
  });
  const createForm = useFormWithSchema(createSchema, {
    defaultValues: { name: "", category: "", compliancePortalVisibility: "NONE" },
  });

  const handleFileUpload = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        const file = acceptedFiles[0];

        if (!Object.keys(acceptedFileTypes).includes(file.type)) {
          createForm.setError("root", {
            type: "manual",
            message: t("newFileDialog.validation.fileTypeNotAllowed"),
          });
          return;
        }

        setUploadedFile(file);
        createForm.clearErrors("root");
        if (!createForm.getValues().name) {
          createForm.setValue("name", file.name.replace(/\.[^/.]+$/, ""));
        }
      }
    },
    [createForm, t],
  );

  const [createFile, isCreating] = useMutation<NewCompliancePortalFileDialog_createMutation>(
    createCompliancePortalFileMutation, {
      successMessage: t("newFileDialog.messages.created"),
      errorToast: t("newFileDialog.errors.create"),
    },
  );
  const handleCreate = async (data: z.infer<typeof createSchema>) => {
    if (!uploadedFile) {
      return;
    }

    await createFile({
      variables: {
        input: {
          compliancePortalId,
          name: data.name,
          category: data.category,
          compliancePortalVisibility: data.compliancePortalVisibility,
          file: null,
        },
        connections: connectionId ? [connectionId] : [],
      },
      uploadables: {
        "input.file": uploadedFile,
      },
    });

    ref.current?.close();
    createForm.reset();
    setUploadedFile(null);
  };

  return (
    <Dialog ref={ref} title={t("newFileDialog.title")}>
      <form onSubmit={e => void createForm.handleSubmit(handleCreate)(e)}>
        <DialogContent padded className="space-y-4">
          <Dropzone
            description={t("newFileDialog.uploadDescription")}
            isUploading={isCreating}
            onDrop={handleFileUpload}
            maxSize={10}
            accept={acceptedFileTypes}
          />
          {uploadedFile && (
            <div className="text-sm text-txt-secondary">
              {t("newFileDialog.selectedFile", { name: uploadedFile.name })}
            </div>
          )}
          {createForm.formState.errors.root && (
            <p className="text-sm text-txt-danger">
              {createForm.formState.errors.root.message}
            </p>
          )}
          <Field
            label={t("newFileDialog.fields.name")}
            type="text"
            {...createForm.register("name")}
            error={createForm.formState.errors.name?.message}
          />
          <Field
            label={t("newFileDialog.fields.category")}
            type="text"
            {...createForm.register("category")}
            error={createForm.formState.errors.category?.message}
          />
          <Field
            label={t("newFileDialog.fields.visibility")}
            type="select"
            value={createForm.watch("compliancePortalVisibility")}
            onValueChange={value =>
              createForm.setValue(
                "compliancePortalVisibility",
                value as "NONE" | "RESTRICTED" | "PUBLIC",
              )}
            error={createForm.formState.errors.compliancePortalVisibility?.message}
          >
            {getCompliancePortalVisibilityOptions(t).map(option => (
              <Option key={option.value} value={option.value}>
                <div className="flex items-center justify-between w-full">
                  <Badge variant={option.variant}>{option.label}</Badge>
                </div>
              </Option>
            ))}
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={isCreating || !uploadedFile}
          >
            {isCreating && <Spinner />}
            {t("newFileDialog.actions.add")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
