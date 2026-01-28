import { getTrustCenterVisibilityOptions } from "@probo/helpers";
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
  "application/csv": [".csv"],
  "application/json": [".json"],
  "application/msword": [".doc"],
  "application/pdf": [".pdf"],
  "application/vnd.ms-excel": [".xls"],
  "application/vnd.ms-powerpoint": [".ppt"],
  "application/vnd.oasis.opendocument.presentation": [".odp"],
  "application/vnd.oasis.opendocument.spreadsheet": [".ods"],
  "application/vnd.oasis.opendocument.text": [".odt"],
  "application/vnd.openxmlformats-officedocument.presentationml.presentation": [
    ".pptx",
  ],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
    ".xlsx",
  ],
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [
    ".docx",
  ],
  "application/yaml": [".yaml", ".yml"],
  "image/gif": [".gif"],
  "image/jpeg": [".jpeg", ".jpg"],
  "image/png": [".png"],
  "image/svg+xml": [".svg"],
  "image/webp": [".webp"],
  "text/csv": [".csv"],
  "text/json": [".json"],
  "text/markdown": [".md"],
  "text/plain": [".txt"],
  "text/uri-list; charset=utf-8": [".uri"],
  "text/uri-list": [".uri"],
  "text/x-log": [".log"],
  "text/yaml": [".yaml", ".yml"],
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
