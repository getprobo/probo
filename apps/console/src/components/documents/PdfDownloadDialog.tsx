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
    const { __ } = useTranslate();
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
          title={__("Download PDF Options")}
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
                      {__(
                        "Show signature information and approval details in the PDF",
                      )}
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
                      {__(
                        "Add confidential watermark with email and timestamp",
                      )}
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
            </DialogContent>
            <DialogFooter>
              <Button type="submit" disabled={isLoading}>
                {isLoading
                  ? (
                      <>
                        <Spinner size={16} />
                        {__("Downloading...")}
                      </>
                    )
                  : (
                      __("Download PDF")
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
