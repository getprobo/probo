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
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const pdfDownloadSchema = z.object({
  withWatermark: z.boolean(),
  watermarkEmail: z
    .string()
    .email("Please enter a valid email address")
    .optional()
    .or(z.literal("")),
  withSignatures: z.boolean(),
});

type PdfDownloadFormData = z.infer<typeof pdfDownloadSchema>;

type Props = {
  children?: ReactNode;
  onDownload: (options: PdfDownloadFormData) => void;
  isLoading?: boolean;
  defaultEmail: string;
};

export type PdfDownloadDialogRef = {
  open: () => void;
  close: () => void;
};

export const PdfDownloadDialog = forwardRef<PdfDownloadDialogRef, Props>(
  ({ children, onDownload, isLoading = false, defaultEmail }, ref) => {
    const { t } = useTranslation();
    const dialogRef = useDialogRef();

    const { register, handleSubmit, formState, watch, setValue }
      = useFormWithSchema(pdfDownloadSchema, {
        defaultValues: {
          withWatermark: false,
          watermarkEmail: defaultEmail,
          withSignatures: true,
        },
      });

    const watchWatermark = watch("withWatermark");
    const watchSignatures = watch("withSignatures");

    useImperativeHandle(ref, () => ({
      open: () => dialogRef.current?.open(),
      close: () => dialogRef.current?.close(),
    }));

    const onSubmit = (data: PdfDownloadFormData) => {
      const options = {
        ...data,
        watermarkEmail: data.withWatermark ? data.watermarkEmail : undefined,
      };
      onDownload(options);
      dialogRef.current?.close();
    };

    return (
      <>
        <div onClick={() => dialogRef.current?.open()}>{children}</div>
        <Dialog
          className="max-w-md"
          ref={dialogRef}
          title={t("pdfDownloadDialog.title")}
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
                      {t("pdfDownloadDialog.includeSignatures.label")}
                    </label>
                    <p className="text-xs text-txt-secondary mt-1">
                      {t("pdfDownloadDialog.includeSignatures.description")}
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
                      {t("pdfDownloadDialog.watermark.label")}
                    </label>
                    <p className="text-xs text-txt-secondary mt-1">
                      {t("pdfDownloadDialog.watermark.description")}
                    </p>
                  </div>
                </div>

                {watchWatermark && (
                  <div className="ml-6">
                    <Field
                      label={t("pdfDownloadDialog.watermark.emailLabel")}
                      {...register("watermarkEmail")}
                      type="email"
                      placeholder={t("pdfDownloadDialog.watermark.emailPlaceholder")}
                      error={formState.errors.watermarkEmail?.message}
                      autoComplete="off"
                      required
                    />
                  </div>
                )}
              </div>
            </DialogContent>
            <DialogFooter>
              <Button type="submit" disabled={isLoading}>
                {isLoading
                  ? (
                      <>
                        <Spinner size={16} />
                        {t("pdfDownloadDialog.actions.downloading")}
                      </>
                    )
                  : (
                      t("pdfDownloadDialog.actions.download")
                    )}
              </Button>
            </DialogFooter>
          </form>
        </Dialog>
      </>
    );
  },
);

PdfDownloadDialog.displayName = "PdfDownloadDialog";
