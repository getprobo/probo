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

import {
  acceptData,
  acceptDocument,
  acceptImage,
  acceptPresentation,
  acceptSpreadsheet,
  acceptText,
  getTrustCenterVisibilityOptions,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Dialog, DialogContent, DialogFooter, type DialogRef, Dropzone, Field, Option, Spinner } from "@probo/ui";
import { useCallback, useState } from "react";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageFileDialog_createMutation } from "#/__generated__/core/NewCompliancePageFileDialog_createMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

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
    trustCenterVisibility: z.enum(["NONE", "PRIVATE", "PUBLIC"]),
  });
  const createForm = useFormWithSchema(createSchema, {
    defaultValues: { name: "", category: "", trustCenterVisibility: "NONE" },
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

  const [createFile, isCreating] = useMutationWithToasts<NewCompliancePageFileDialog_createMutation>(
    createCompliancePageFileMutation, {
      successMessage: "File uploaded successfully",
      errorMessage: "Failed to upload file",
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
          trustCenterVisibility: data.trustCenterVisibility,
          file: null,
        },
        connections: connectionId ? [connectionId] : [],
      },
      uploadables: {
        "input.file": uploadedFile,
      },
      onSuccess: () => {
        ref.current?.close();
        createForm.reset();
        setUploadedFile(null);
      },
    });
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
            value={createForm.watch("trustCenterVisibility")}
            onValueChange={value =>
              createForm.setValue(
                "trustCenterVisibility",
                value as "NONE" | "PRIVATE" | "PUBLIC",
              )}
            error={createForm.formState.errors.trustCenterVisibility?.message}
          >
            {getTrustCenterVisibilityOptions(__).map(option => (
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
