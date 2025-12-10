import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Spinner,
  useDialogRef,
  Dropzone,
  Option,
  Badge,
  IconPlusLarge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOutletContext } from "react-router";
import { useState, useCallback, use } from "react";
import z from "zod";
import { getTrustCenterVisibilityOptions } from "@probo/helpers";
import {
  useCreateTrustCenterFileMutation,
  useUpdateTrustCenterFileMutation,
  useDeleteTrustCenterFileMutation,
} from "/hooks/graph/TrustCenterFileGraph";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { TrustCenterFilesCard } from "/components/trustCenter/TrustCenterFilesCard";
import type { TrustCenterFilesCardFragment$key } from "/components/trustCenter/__generated__/TrustCenterFilesCardFragment.graphql";
import { PermissionsContext } from "/providers/PermissionsContext";

type ContextType = {
  organization: {
    id: string;
    trustCenterFiles?: {
      __id?: string;
      edges: Array<{
        node: TrustCenterFilesCardFragment$key;
      }>;
    };
  };
};

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
  "application/vnd.openxmlformats-officedocument.presentationml.presentation": [".pptx"],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [".xlsx"],
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [".docx"],
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
}

export default function TrustCenterFilesTab() {
  const { __ } = useTranslate();
  const { organization } = useOutletContext<ContextType>();
  const { isAuthorized } = use(PermissionsContext);
  const createSchema = z.object({
    name: z.string().min(1, __("Name is required")),
    category: z.string().min(1, __("Category is required")),
    trustCenterVisibility: z.enum(["NONE", "PRIVATE", "PUBLIC"]),
  });

  const editSchema = z.object({
    name: z.string().min(1, __("Name is required")),
    category: z.string().min(1, __("Category is required")),
  });

  const [createFile, isCreating] = useCreateTrustCenterFileMutation();
  const [updateFile, isUpdating] = useUpdateTrustCenterFileMutation();
  const [deleteFile, isDeleting] = useDeleteTrustCenterFileMutation();

  const createDialogRef = useDialogRef();
  const editDialogRef = useDialogRef();
  const deleteDialogRef = useDialogRef();

  const [editingFile, setEditingFile] = useState<{ id: string; name: string; category: string } | null>(null);
  const [deletingFileId, setDeletingFileId] = useState<string | null>(null);
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const createForm = useFormWithSchema(createSchema, {
    defaultValues: { name: "", category: "", trustCenterVisibility: "NONE" },
  });

  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: "", category: "" },
  });

  const files = organization.trustCenterFiles?.edges?.map((edge) => edge.node) || [];

  const handleFileUpload = useCallback((acceptedFiles: File[]) => {
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
  }, [createForm, __]);

  const handleCreate = createForm.handleSubmit(async (data) => {
    if (!uploadedFile) {
      return;
    }

    setIsUploading(true);

    const connectionId = organization.trustCenterFiles?.__id;

    try {
      await createFile({
        variables: {
          input: {
            organizationId: organization.id,
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
          createDialogRef.current?.close();
          createForm.reset();
          setUploadedFile(null);
        },
      });
    } finally {
      setIsUploading(false);
    }
  });

  const handleEdit = useCallback((file: { id: string; name: string; category: string }) => {
    setEditingFile(file);
    editForm.reset({ name: file.name, category: file.category });
    editDialogRef.current?.open();
  }, [editDialogRef, editForm]);

  const handleUpdate = editForm.handleSubmit(async (data) => {
    if (!editingFile) {
      return;
    }

    await updateFile({
      variables: {
        input: {
          id: editingFile.id,
          name: data.name,
          category: data.category,
        },
      },
      onSuccess: () => {
        editDialogRef.current?.close();
        setEditingFile(null);
      },
    });
  });

  const handleDeleteClick = useCallback((id: string) => {
    setDeletingFileId(id);
    deleteDialogRef.current?.open();
  }, [deleteDialogRef]);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deletingFileId) {
      return;
    }

    const connectionId = organization.trustCenterFiles?.__id;

    await deleteFile({
      variables: {
        input: { id: deletingFileId },
        connections: connectionId ? [connectionId] : [],
      },
      onSuccess: () => {
        deleteDialogRef.current?.close();
        setDeletingFileId(null);
      },
    });
  }, [deletingFileId, deleteFile, deleteDialogRef, organization.trustCenterFiles?.__id]);

  const handleChangeVisibility = useCallback((params: {
    variables: {
      input: {
        id: string;
        trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC";
      };
    };
  }) => {
    updateFile(params);
  }, [updateFile]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Files")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Upload and manage files for your trust center")}
          </p>
        </div>
        {isAuthorized("Organization", "createTrustCenterFile") && (
          <Button icon={IconPlusLarge} onClick={() => createDialogRef.current?.open()}>
            {__("Add File")}
          </Button>
        )}
      </div>
      {(isUpdating || isDeleting) && (
        <div className="flex items-center justify-center">
          <Spinner />
        </div>
      )}
      <TrustCenterFilesCard
        files={files}
        params={{}}
        disabled={isUpdating || isDeleting}
        onChangeVisibility={handleChangeVisibility}
        onEdit={handleEdit}
        onDelete={handleDeleteClick}
      />

      <Dialog ref={createDialogRef} title={__("Add File")}>
        <form onSubmit={handleCreate}>
          <DialogContent padded className="space-y-4">
            <Dropzone
              description={__("Upload file (max 10MB)")}
              isUploading={isUploading}
              onDrop={handleFileUpload}
              maxSize={10}
              accept={acceptedFileTypes}
            />
            {uploadedFile && (
              <div className="text-sm text-txt-secondary">
                {__("Selected file")}: {uploadedFile.name}
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
              onValueChange={(value) => createForm.setValue("trustCenterVisibility", value as "NONE" | "PRIVATE" | "PUBLIC")}
              error={createForm.formState.errors.trustCenterVisibility?.message}
            >
              {getTrustCenterVisibilityOptions(__).map((option) => (
                <Option key={option.value} value={option.value}>
                  <div className="flex items-center justify-between w-full">
                    <Badge variant={option.variant}>
                      {option.label}
                    </Badge>
                  </div>
                </Option>
              ))}
            </Field>
          </DialogContent>
          <DialogFooter>
            <Button
              type="submit"
              disabled={isCreating || isUploading || !uploadedFile}
            >
              {(isCreating || isUploading) && <Spinner />}
              {__("Add File")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <Dialog ref={editDialogRef} title={__("Edit File")}>
        <form onSubmit={handleUpdate}>
          <DialogContent padded className="space-y-4">
            <Field
              label={__("Name")}
              type="text"
              {...editForm.register("name")}
              error={editForm.formState.errors.name?.message}
            />
            <Field
              label={__("Category")}
              type="text"
              {...editForm.register("category")}
              error={editForm.formState.errors.category?.message}
            />
          </DialogContent>
          <DialogFooter>
            <Button
              type="submit"
              disabled={isUpdating}
            >
              {isUpdating && <Spinner />}
              {__("Save")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <Dialog ref={deleteDialogRef} title={__("Delete File")}>
        <DialogContent padded>
          <p>{__("Are you sure you want to delete this file? This action cannot be undone.")}</p>
        </DialogContent>
        <DialogFooter>
          <Button
            variant="danger"
            onClick={handleDeleteConfirm}
            disabled={isDeleting}
          >
            {isDeleting && <Spinner />}
            {__("Delete")}
          </Button>
        </DialogFooter>
      </Dialog>
    </div>
  );
}
