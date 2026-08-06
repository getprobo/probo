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

import {
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, type ReactNode, useImperativeHandle } from "react";
import { useTranslation } from "react-i18next";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import {
  getWatermarkTextLength,
  maxWatermarkTextLength,
  truncateWatermarkText,
  type WatermarkExportFormData,
  watermarkExportSchema,
} from "./watermark";

type Props = {
  children: ReactNode;
  onExport: (options: WatermarkExportFormData) => Promise<void>;
  isLoading?: boolean;
  defaultEmail: string;
  selectedCount: number;
};

export type BulkExportDialogRef = {
  open: () => void;
  close: () => void;
};

export const BulkExportDialog = forwardRef<BulkExportDialogRef, Props>(
  ({ children, onExport, isLoading = false, defaultEmail, selectedCount }, ref) => {
    const { t } = useTranslation();
    const dialogRef = useDialogRef();

    const { register, handleSubmit, formState, watch, setValue } = useFormWithSchema(
      watermarkExportSchema,
      {
        defaultValues: {
          withWatermark: false,
          watermarkText: truncateWatermarkText(defaultEmail),
          withSignatures: true,
        },
      },
    );

    const watchWatermark = watch("withWatermark");
    const watchWatermarkText = watch("watermarkText") ?? "";
    const watchSignatures = watch("withSignatures");

    useImperativeHandle(ref, () => ({
      open: () => dialogRef.current?.open(),
      close: () => dialogRef.current?.close(),
    }));

    const onSubmit = async (data: WatermarkExportFormData) => {
      const options = {
        ...data,
        watermarkText: data.withWatermark ? data.watermarkText : undefined,
      };
      await onExport(options);
      dialogRef.current?.close();
    };

    return (
      <>
        <div onClick={() => dialogRef.current?.open()}>{children}</div>
        <Dialog
          className="max-w-md"
          ref={dialogRef}
          title={t("bulkExportDialog.title", { count: selectedCount })}
        >
          <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
            <DialogContent className="space-y-4" padded>
              <div className="space-y-4">
                <div className="flex items-start gap-3">
                  <Checkbox
                    checked={watchSignatures}
                    onChange={checked => setValue("withSignatures", checked)}
                  />
                  <div className="flex-1">
                    <label className="text-sm font-medium text-txt-primary cursor-pointer">
                      {t("bulkExportDialog.includeSignatures.label")}
                    </label>
                    <p className="text-xs text-txt-secondary mt-1">
                      {t("bulkExportDialog.includeSignatures.description")}
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3">
                  <Checkbox
                    checked={watchWatermark}
                    onChange={checked => setValue("withWatermark", checked)}
                  />
                  <div className="flex-1">
                    <label className="text-sm font-medium text-txt-primary cursor-pointer">
                      {t("bulkExportDialog.watermark.label")}
                    </label>
                    <p className="text-xs text-txt-secondary mt-1">
                      {t("bulkExportDialog.watermark.description")}
                    </p>
                  </div>
                </div>

                {watchWatermark && (
                  <div className="ml-6">
                    <Field
                      label={t("bulkExportDialog.watermark.textLabel")}
                      {...register("watermarkText")}
                      type="text"
                      placeholder={t("bulkExportDialog.watermark.textPlaceholder")}
                      error={formState.errors.watermarkText?.message}
                      help={t("bulkExportDialog.watermark.byteCount", {
                        count: getWatermarkTextLength(watchWatermarkText),
                        max: maxWatermarkTextLength,
                      })}
                      autoComplete="off"
                      required
                    />
                  </div>
                )}
              </div>

              <div className="bg-level-1 p-3 rounded-lg border border-border-subtle">
                <p className="text-sm text-txt-secondary">
                  {t("bulkExportDialog.notice")}
                </p>
              </div>
            </DialogContent>
            <DialogFooter>
              <Button
                type="submit"
                disabled={isLoading}
              >
                {isLoading
                  ? (
                      <>
                        <Spinner size={16} />
                        {t("bulkExportDialog.actions.exporting")}
                      </>
                    )
                  : (
                      t("bulkExportDialog.actions.export")
                    )}
              </Button>
            </DialogFooter>
          </form>
        </Dialog>
      </>
    );
  },
);

BulkExportDialog.displayName = "BulkExportDialog";
