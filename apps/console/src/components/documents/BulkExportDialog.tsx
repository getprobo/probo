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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const bulkExportSchema = z.object({
  withWatermark: z.boolean(),
  watermarkEmail: z.string().optional().or(z.literal("")),
  withSignatures: z.boolean(),
}).refine((data) => {
  if (data.withWatermark && (!data.watermarkEmail || data.watermarkEmail === "")) {
    return false;
  }
  if (data.withWatermark && data.watermarkEmail && !z.string().email().safeParse(data.watermarkEmail).success) {
    return false;
  }
  return true;
}, {
  message: "Please enter a valid email address",
  path: ["watermarkEmail"],
});

type BulkExportFormData = z.infer<typeof bulkExportSchema>;

type Props = {
  children: ReactNode;
  onExport: (options: BulkExportFormData) => Promise<void>;
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
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();

    const { register, handleSubmit, formState, watch, setValue } = useFormWithSchema(
      bulkExportSchema,
      {
        defaultValues: {
          withWatermark: false,
          watermarkEmail: defaultEmail,
          withSignatures: true,
        },
      },
    );

    const watchWatermark = watch("withWatermark");
    const watchSignatures = watch("withSignatures");

    useImperativeHandle(ref, () => ({
      open: () => dialogRef.current?.open(),
      close: () => dialogRef.current?.close(),
    }));

    const onSubmit = async (data: BulkExportFormData) => {
      const options = {
        ...data,
        watermarkEmail: data.withWatermark ? data.watermarkEmail : undefined,
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
          title={sprintf(__("Export %s Documents"), selectedCount)}
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
                      {__("Include signatures")}
                    </label>
                    <p className="text-xs text-txt-secondary mt-1">
                      {__("Show signature information and approval details in the PDFs")}
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
                      {__("Add watermark")}
                    </label>
                    <p className="text-xs text-txt-secondary mt-1">
                      {__("Add confidential watermark with email and timestamp to all PDFs")}
                    </p>
                  </div>
                </div>

                {watchWatermark && (
                  <div className="ml-6">
                    <Field
                      label={__("Watermark email")}
                      {...register("watermarkEmail")}
                      type="email"
                      placeholder={__("Enter email address")}
                      error={formState.errors.watermarkEmail?.message}
                      autoComplete="off"
                      required
                    />
                  </div>
                )}
              </div>

              <div className="bg-level-1 p-3 rounded-lg border border-border-subtle">
                <p className="text-sm text-txt-secondary">
                  {__("The documents will be exported as individual PDFs in a ZIP file. You will receive an email when the export is ready for download.")}
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
                        {__("Exporting...")}
                      </>
                    )
                  : (
                      __("Export Documents")
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
