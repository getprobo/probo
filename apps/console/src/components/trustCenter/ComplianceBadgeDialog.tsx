import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Dropzone,
  Field,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, type ReactNode, useImperativeHandle, useState } from "react";
import { z } from "zod";

import type { CompliancePageBadgeListItemFragment$data } from "#/__generated__/core/CompliancePageBadgeListItemFragment.graphql";
import {
  useCreateComplianceBadgeMutation,
  useUpdateComplianceBadgeMutation,
} from "#/hooks/graph/ComplianceBadgeGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const badgeSchema = z.object({
  name: z.string().min(1, "Name is required"),
  rank: z.number().int().positive().optional(),
});

type BadgeFormData = z.infer<typeof badgeSchema>;

export type ComplianceBadgeDialogRef = {
  openCreate: (trustCenterId: string, connectionId: string) => void;
  openEdit: (badge: CompliancePageBadgeListItemFragment$data) => void;
};

export const ComplianceBadgeDialog = forwardRef<ComplianceBadgeDialogRef, { children?: ReactNode }>(
  function ComplianceBadgeDialog({ children }, ref) {
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();
    const [mode, setMode] = useState<"create" | "edit">("create");
    const [trustCenterId, setTrustCenterId] = useState<string>("");
    const [connectionId, setConnectionId] = useState<string>("");
    const [editBadge, setEditBadge] = useState<CompliancePageBadgeListItemFragment$data | null>(null);
    const [uploadedFile, setUploadedFile] = useState<File | null>(null);

    const [createBadge, isCreating] = useCreateComplianceBadgeMutation();
    const [updateBadge, isUpdating] = useUpdateComplianceBadgeMutation();

    const { register, handleSubmit, formState: { errors }, reset } = useFormWithSchema(
      badgeSchema,
      {
        defaultValues: {
          name: "",
        },
      },
    );

    useImperativeHandle(ref, () => ({
      openCreate: (tId: string, cId: string) => {
        setMode("create");
        setTrustCenterId(tId);
        setConnectionId(cId);
        setEditBadge(null);
        setUploadedFile(null);
        reset({ name: "" });
        dialogRef.current?.open();
      },
      openEdit: (badge: CompliancePageBadgeListItemFragment$data) => {
        setMode("edit");
        setEditBadge(badge);
        setUploadedFile(null);
        reset({
          name: badge.name,
          rank: badge.rank,
        });
        dialogRef.current?.open();
      },
    }));

    const handleDrop = (files: File[]) => {
      if (files.length > 0) {
        setUploadedFile(files[0]);
      }
    };

    const onSubmit = async (data: BadgeFormData) => {
      if (mode === "create") {
        if (!uploadedFile) return;

        await createBadge({
          variables: {
            input: {
              trustCenterId,
              name: data.name,
              iconFile: null,
            },
            connections: [connectionId],
          },
          uploadables: {
            "input.iconFile": uploadedFile,
          },
          onSuccess: () => {
            reset();
            setUploadedFile(null);
            dialogRef.current?.close();
          },
        });
      } else if (editBadge) {
        const input: {
          id: string;
          name: string;
          rank?: number;
          iconFile?: null;
        } = {
          id: editBadge.id,
          name: data.name,
        };

        if (data.rank !== undefined) {
          input.rank = data.rank;
        }

        const uploadables: Record<string, File> = {};
        if (uploadedFile) {
          input.iconFile = null;
          uploadables["input.iconFile"] = uploadedFile;
        }

        await updateBadge({
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
    const title = mode === "create" ? __("Add Badge") : __("Edit Badge");

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
          className="max-w-lg"
          onClose={handleClose}
        >
          <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
            <DialogContent padded className="space-y-6">
              <Field
                {...register("name")}
                label={__("Badge Name")}
                type="text"
                required
                error={errors.name?.message}
                placeholder={__("e.g. ISO 27001, SOC 2 Type II")}
              />

              {mode === "edit" && (
                <Field
                  {...register("rank", { setValueAs: (v: string) => v === "" ? undefined : Number(v) })}
                  label={__("Rank")}
                  type="number"
                  min={1}
                  error={errors.rank?.message}
                  placeholder={__("Display order (1, 2, 3...)")}
                  help={__("Lower numbers appear first")}
                />
              )}

              <Field label={__("Icon")}>
                <Dropzone
                  description={__("Upload badge icon (PNG, JPG, SVG, WEBP up to 5MB)")}
                  isUploading={isSubmitting}
                  onDrop={handleDrop}
                  accept={{
                    "image/png": [".png"],
                    "image/jpeg": [".jpg", ".jpeg"],
                    "image/svg+xml": [".svg"],
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
                      {__("Current icon will be kept if no new file is uploaded")}
                    </p>
                  </div>
                )}
                {mode === "create" && !uploadedFile && (
                  <div className="mt-2 p-3 bg-warning-subtle rounded-lg">
                    <p className="text-sm">{__("Icon is required for new badges")}</p>
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
                {mode === "create" ? __("Add Badge") : __("Update Badge")}
              </Button>
            </DialogFooter>
          </form>
        </Dialog>
      </>
    );
  },
);
