// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { acceptImage } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
import { z } from "zod";

import type { CompliancePageReferenceListItemFragment$data } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import {
  useCreateCompliancePageReferenceMutation,
  useUpdateCompliancePageReferenceMutation,
} from "#/pages/organizations/compliance-page/_lib/compliancePageReferenceMutations";

const referenceSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string(),
  websiteUrl: z.string().url("Please enter a valid URL"),
  rank: z.number().int().positive().optional(),
});

type ReferenceFormData = z.infer<typeof referenceSchema>;

export type CompliancePageReferenceDialogRef = {
  openCreate: (compliancePageId: string, connectionId: string) => void;
  openEdit: (reference: CompliancePageReferenceListItemFragment$data, rank: number) => void;
};

export const CompliancePageReferenceDialog = forwardRef<CompliancePageReferenceDialogRef, { children?: ReactNode }>(
  function CompliancePageReferenceDialog({ children }, ref) {
    const { t } = useTranslation();
    const dialogRef = useDialogRef();
    const [mode, setMode] = useState<"create" | "edit">("create");
    const [compliancePageId, setCompliancePageId] = useState<string>("");
    const [connectionId, setConnectionId] = useState<string>("");
    const [editReference, setEditReference] = useState<CompliancePageReferenceListItemFragment$data | null>(null);
    const [uploadedFile, setUploadedFile] = useState<File | null>(null);

    const [createReference, isCreating] = useCreateCompliancePageReferenceMutation();
    const [updateReference, isUpdating] = useUpdateCompliancePageReferenceMutation();

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
      openCreate: (pageId: string, cId: string) => {
        setMode("create");
        setCompliancePageId(pageId);
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
      openEdit: (reference: CompliancePageReferenceListItemFragment$data, rank: number) => {
        setMode("edit");
        setEditReference(reference);
        setUploadedFile(null);
        reset({
          name: reference.name,
          description: reference.description ?? undefined,
          websiteUrl: reference.websiteUrl,
          rank,
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
              compliancePortalId: compliancePageId,
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
        });

        reset();
        setUploadedFile(null);
        dialogRef.current?.close();
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
        });

        reset();
        setUploadedFile(null);
        dialogRef.current?.close();
      }
    };

    const handleClose = () => {
      reset();
      setUploadedFile(null);
    };

    const isSubmitting = isCreating || isUpdating;
    const title = mode === "create" ? t("trustCenterReferenceDialog.actions.add") : t("trustCenterReferenceDialog.actions.edit");

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
                label={t("trustCenterReferenceDialog.fields.name.label")}
                type="text"
                required
                error={errors.name?.message}
                placeholder={t("trustCenterReferenceDialog.fields.name.placeholder")}
              />

              <Field label={t("trustCenterReferenceDialog.fields.description.label")} error={errors.description?.message}>
                <Textarea
                  {...register("description")}
                  placeholder={t("trustCenterReferenceDialog.fields.description.placeholder")}
                  rows={3}
                />
              </Field>

              <Field
                {...register("websiteUrl")}
                label={t("trustCenterReferenceDialog.fields.websiteUrl.label")}
                type="url"
                required
                error={errors.websiteUrl?.message}
                placeholder={t("trustCenterReferenceDialog.fields.websiteUrl.placeholder")}
              />

              {mode === "edit" && (
                <Field
                  {...register("rank", { valueAsNumber: true })}
                  label={t("trustCenterReferenceDialog.fields.rank.label")}
                  type="number"
                  min={1}
                  error={errors.rank?.message}
                  placeholder={t("trustCenterReferenceDialog.fields.rank.placeholder")}
                  help={t("trustCenterReferenceDialog.fields.rank.help")}
                />
              )}

              <Field label={t("trustCenterReferenceDialog.fields.logo.label")}>
                <Dropzone
                  description={t("trustCenterReferenceDialog.fields.logo.description")}
                  isUploading={isSubmitting}
                  onDrop={handleDrop}
                  accept={acceptImage}
                  maxSize={5}
                />
                {uploadedFile && (
                  <div className="mt-2 p-3 bg-tertiary-subtle rounded-lg">
                    <p className="text-sm font-medium">
                      {t("trustCenterReferenceDialog.fields.logo.selectedFile")}
                      :
                    </p>
                    <p className="text-sm text-txt-secondary">{uploadedFile.name}</p>
                  </div>
                )}
                {mode === "edit" && !uploadedFile && (
                  <div className="mt-2 p-3 bg-tertiary-subtle rounded-lg">
                    <p className="text-sm text-txt-secondary">
                      {t("trustCenterReferenceDialog.fields.logo.keepCurrent")}
                    </p>
                  </div>
                )}
                {mode === "create" && !uploadedFile && (
                  <div className="mt-2 p-3 bg-warning-subtle rounded-lg">
                    <p className="text-sm">
                      {t("trustCenterReferenceDialog.fields.logo.required")}
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
                {mode === "create" ? t("trustCenterReferenceDialog.actions.add") : t("trustCenterReferenceDialog.actions.update")}
              </Button>
            </DialogFooter>
          </form>
        </Dialog>
      </>
    );
  },
);
