import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  Button,
  Field,
  useToast,
  useDialogRef,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { z } from "zod";
import { useRequestReportAccess, type TrustCenterAudit } from "/hooks/useTrustCenterQueries";

type Props = {
  trigger: React.ReactNode;
  report: NonNullable<TrustCenterAudit["report"]>;
  audit: TrustCenterAudit;
  trustCenterId: string;
  organizationName: string;
  isAuthenticated: boolean;
};

export function RequestReportAccessDialog({
  trigger,
  report,
  audit,
  trustCenterId,
  organizationName,
  isAuthenticated
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const dialogRef = useDialogRef();

  const mutation = useRequestReportAccess();

  const schema = z.object({
    name: isAuthenticated
      ? z.string().optional()
      : z.string().min(1, __("Name is required")).min(2, __("Name must be at least 2 characters long")),
    email: isAuthenticated
      ? z.string().optional()
      : z.string().min(1, __("Email is required")).email(__("Please enter a valid email address")),
  });

  const { register, handleSubmit, formState, reset } = useFormWithSchema(schema, {
    defaultValues: { name: "", email: "" },
  });

  const onSubmit = handleSubmit(async (data) => {
    setIsSubmitting(true);
    mutation.mutate(
      {
        trustCenterId,
        email: data.email || "",
        name: data.name || "",
        reportId: report.id,
      },
      {
        onSuccess: () => {
          toast({
            title: __("Request Submitted"),
            description: sprintf(__("Your access request for %s compliance report has been submitted successfully."), audit.framework.name),
          });

          reset();
          dialogRef.current?.close();
          setIsSubmitting(false);
          window.location.reload();
        },
        onError: () => {
          toast({
            title: __("Error"),
            description: __("Failed to submit access request. Please try again."),
            variant: "error",
          });
          setIsSubmitting(false);
        },
      }
    );
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={trigger}
      title={__("Request Access to Compliance Report")}
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <div className="text-sm text-txt-secondary">
            {sprintf(__("Request access to %s compliance report from %s. Your request will be reviewed and you will receive an email notification with access instructions if approved."), audit.framework.name, organizationName)}
          </div>

          {!isAuthenticated && (
            <>
              <Field
                label={__("Your Name")}
                required
                error={formState.errors.name?.message}
                {...register("name")}
                placeholder={__("Enter your full name")}
                disabled={isSubmitting}
              />

              <Field
                label={__("Email Address")}
                required
                type="email"
                error={formState.errors.email?.message}
                {...register("email")}
                placeholder={__("Enter your email address")}
                disabled={isSubmitting}
              />
            </>
          )}
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            disabled={isSubmitting || mutation.isPending}
          >
            {isSubmitting || mutation.isPending ? __("Submitting...") : __("Submit Request")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
