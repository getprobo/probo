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
  getCompliancePageVisibilityOptions,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Dialog, DialogContent, DialogFooter, type DialogRef, Dropzone, Field, Option, Spinner } from "@probo/ui";
import { useCallback, useState } from "react";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageFileDialog_createMutation } from "#/__generated__/core/NewCompliancePageFileDialog_createMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const acceptedFileTypes = {
  ...acceptDocument,
  ...acceptSpreadsheet,
  ...acceptPresentation,
  ...acceptText,
  ...acceptImage,
  ...acceptData,
};

const createCompliancePageFileMutation = graphql`
  mutation NewCompliancePageFileDialog_createMutation(
    $input: CreateTrustCenterFileInput!
    $connections: [ID!]!
  ) {
    createTrustCenterFile(input: $input) {
      trustCenterFileEdge @prependEdge(connections: $connections) {
        node {
          ...CompliancePageFileListItem_fileFragment
        }
      }
    }
  }
`;

export function NewCompliancePageFileDialog(props: {
  connectionId: DataID;
  ref: DialogRef;
}) {
  const { connectionId, ref } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const [uploadedFile, setUploadedFile] = useState<File | null>(null);

  const createSchema = z.object({
    name: z.string().min(1, __("Name is required")),
    category: z.string().min(1, __("Category is required")),
    compliancePageVisibility: z.enum(["NONE", "PRIVATE", "PUBLIC"]),
  });
  const createForm = useFormWithSchema(createSchema, {
    defaultValues: { name: "", category: "", compliancePageVisibility: "NONE" },
  });

  const handleFileUpload = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        const file = acceptedFiles[0];

        if (!Object.keys(acceptedFileTypes).includes(file.type)) {
          createForm.setError("root", {
            type: "manual",
            message: __("File type is not allowed"),
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
    [createForm, __],
  );

  const [createFile, isCreating] = useMutation<NewCompliancePageFileDialog_createMutation>(
    createCompliancePageFileMutation, {
      successMessage: "File uploaded successfully",
      errorToast: "Failed to upload file",
    },
  );
  const handleCreate = async (data: z.infer<typeof createSchema>) => {
    if (!uploadedFile) {
      return;
    }

    await createFile({
      variables: {
        input: {
          organizationId,
          name: data.name,
          category: data.category,
          trustCenterVisibility: data.compliancePageVisibility,
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
    <Dialog ref={ref} title={__("Add File")}>
      <form onSubmit={e => void createForm.handleSubmit(handleCreate)(e)}>
        <DialogContent padded className="space-y-4">
          <Dropzone
            description={__("Upload file (max 10MB)")}
            isUploading={isCreating}
            onDrop={handleFileUpload}
            maxSize={10}
            accept={acceptedFileTypes}
          />
          {uploadedFile && (
            <div className="text-sm text-txt-secondary">
              {__("Selected file")}
              :
              {uploadedFile.name}
            </div>
          )}
          {createForm.formState.errors.root && (
            <p className="text-sm text-txt-danger">
              {createForm.formState.errors.root.message}
            </p>
          )}
          <Field
            label={__("Name")}
            type="text"
            {...createForm.register("name")}
            error={createForm.formState.errors.name?.message}
          />
          <Field
            label={__("Category")}
            type="text"
            {...createForm.register("category")}
            error={createForm.formState.errors.category?.message}
          />
          <Field
            label={__("Visibility")}
            type="select"
            value={createForm.watch("compliancePageVisibility")}
            onValueChange={value =>
              createForm.setValue(
                "compliancePageVisibility",
                value as "NONE" | "PRIVATE" | "PUBLIC",
              )}
            error={createForm.formState.errors.compliancePageVisibility?.message}
          >
            {getCompliancePageVisibilityOptions(__).map(option => (
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
            {__("Add File")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
