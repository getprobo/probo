import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Dropzone,
  Field,
  Spinner,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, type ReactNode, useImperativeHandle, useState } from "react";
import { z } from "zod";

import type { CompliancePageReferenceListItemFragment$data } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import {
  useCreateTrustCenterReferenceMutation,
  useUpdateTrustCenterReferenceMutation,
} from "#/hooks/graph/TrustCenterReferenceGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const referenceSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string(),
  websiteUrl: z.string().url("Please enter a valid URL"),
  rank: z.number().int().positive().optional(),
});

type ReferenceFormData = z.infer<typeof referenceSchema>;

export type TrustCenterReferenceDialogRef = {
  openCreate: (trustCenterId: string, connectionId: string) => void;
  openEdit: (reference: CompliancePageReferenceListItemFragment$data) => void;
};

export const TrustCenterReferenceDialog = forwardRef<TrustCenterReferenceDialogRef, { children?: ReactNode }>(
  function TrustCenterReferenceDialog({ children }, ref) {
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();
    const [mode, setMode] = useState<"create" | "edit">("create");
    const [trustCenterId, setTrustCenterId] = useState<string>("");
    const [connectionId, setConnectionId] = useState<string>("");
    const [editReference, setEditReference] = useState<CompliancePageReferenceListItemFragment$data | null>(null);
    const [uploadedFile, setUploadedFile] = useState<File | null>(null);

    const [createReference, isCreating] = useCreateTrustCenterReferenceMutation();
    const [updateReference, isUpdating] = useUpdateTrustCenterReferenceMutation();

    const { register, handleSubmit, formState: { errors }, reset } = useFormWithSchema(
      referenceSchema,
      {
        defaultValues: {
          name: "",
          description: "",
          websiteUrl: "",
        },
      },
    );

    useImperativeHandle(ref, () => ({
      openCreate: (tId: string, cId: string) => {
        setMode("create");
        setTrustCenterId(tId);
        setConnectionId(cId);
        setEditReference(null);
        setUploadedFile(null);
        reset({
          name: "",
          description: "",
          websiteUrl: "",
        });
        dialogRef.current?.open();
      },
      openEdit: (reference: CompliancePageReferenceListItemFragment$data) => {
        setMode("edit");
        setEditReference(reference);
        setUploadedFile(null);
        reset({
          name: reference.name,
          description: reference.description ?? undefined,
          websiteUrl: reference.websiteUrl,
          rank: reference.rank,
        });
        dialogRef.current?.open();
      },
    }));

    const handleDrop = (files: File[]) => {
      if (files.length > 0) {
        const file = files[0];
        setUploadedFile(file);
      }
    };

    const onSubmit = async (data: ReferenceFormData) => {
      if (mode === "create") {
        if (!uploadedFile) {
          return;
        }

        await createReference({
          variables: {
            input: {
              trustCenterId,
              name: data.name,
              description: data.description || null,
              websiteUrl: data.websiteUrl,
              logoFile: null,
            },
            connections: [connectionId],
          },
          uploadables: {
            "input.logoFile": uploadedFile,
          },
          onSuccess: () => {
            reset();
            setUploadedFile(null);
            dialogRef.current?.close();
          },
        });
      } else if (editReference) {
        const input: {
          id: string;
          name: string;
          description: string | null;
          websiteUrl: string;
          rank?: number;
          logoFile?: null;
        } = {
          id: editReference.id,
          name: data.name,
          description: data.description || null,
          websiteUrl: data.websiteUrl,
        };

        if (data.rank !== undefined) {
          input.rank = data.rank;
        }

        const uploadables: Record<string, File> = {};

        if (uploadedFile) {
          input.logoFile = null;
          uploadables["input.logoFile"] = uploadedFile;
        }

        await updateReference({
          variables: { input },
          uploadables: Object.keys(uploadables).length > 0 ? uploadables : undefined,
          onSuccess: () => {
            reset();
            setUploadedFile(null);
            dialogRef.current?.close();
          },
        });
      }
    };

    const handleClose = () => {
      reset();
      setUploadedFile(null);
    };

    const isSubmitting = isCreating || isUpdating;
    const title = mode === "create" ? __("Add Reference") : __("Edit Reference");

    return (
      <>
        {children && (
          <span onClick={() => mode === "create" && dialogRef.current?.open()}>
            {children}
          </span>
        )}

        <Dialog
          ref={dialogRef}
          title={title}
          className="max-w-2xl"
          onClose={handleClose}
        >
          <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
            <DialogContent padded className="space-y-6">
              <Field
                {...register("name")}
                label={__("Reference Name")}
                type="text"
                required
                error={errors.name?.message}
                placeholder={__("Company or organization name")}
              />

              <Field label={__("Description")} error={errors.description?.message}>
                <Textarea
                  {...register("description")}
                  placeholder={__("Brief description of the reference")}
                  rows={3}
                />
              </Field>

              <Field
                {...register("websiteUrl")}
                label={__("Website URL")}
                type="url"
                required
                error={errors.websiteUrl?.message}
                placeholder={__("https://example.com")}
              />

              {mode === "edit" && (
                <Field
                  {...register("rank", { valueAsNumber: true })}
                  label={__("Rank")}
                  type="number"
                  min={1}
                  error={errors.rank?.message}
                  placeholder={__("Display order (1, 2, 3...)")}
                  help={__("Lower numbers appear first")}
                />
              )}

              <Field label={__("Logo")}>
                <Dropzone
                  description={__("Upload logo image (PNG, JPG, WEBP up to 5MB)")}
                  isUploading={isSubmitting}
                  onDrop={handleDrop}
                  accept={{
                    "image/png": [".png"],
                    "image/jpeg": [".jpg", ".jpeg"],
                    "image/webp": [".webp"],
                  }}
                  maxSize={5}
                />
                {uploadedFile && (
                  <div className="mt-2 p-3 bg-tertiary-subtle rounded-lg">
                    <p className="text-sm font-medium">
                      {__("Selected file")}
                      :
                    </p>
                    <p className="text-sm text-txt-secondary">{uploadedFile.name}</p>
                  </div>
                )}
                {mode === "edit" && !uploadedFile && (
                  <div className="mt-2 p-3 bg-tertiary-subtle rounded-lg">
                    <p className="text-sm text-txt-secondary">
                      {__("Current logo will be kept if no new file is uploaded")}
                    </p>
                  </div>
                )}
                {mode === "create" && !uploadedFile && (
                  <div className="mt-2 p-3 bg-warning-subtle rounded-lg">
                    <p className="text-sm">
                      {__("Logo is required for new references")}
                    </p>
                  </div>
                )}
              </Field>
            </DialogContent>

            <DialogFooter>
              <Button
                type="submit"
                disabled={isSubmitting || (mode === "create" && !uploadedFile)}
                icon={isSubmitting ? Spinner : undefined}
              >
                {mode === "create" ? __("Add Reference") : __("Update Reference")}
              </Button>
            </DialogFooter>
          </form>
        </Dialog>
      </>
    );
  },
);
