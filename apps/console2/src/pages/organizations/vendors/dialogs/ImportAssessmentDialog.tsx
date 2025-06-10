import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { ReactNode } from "react";

const schema = z.object({
  url: z.string().url(),
});

const importAssessmentMutation = graphql`
  mutation ImportAssessmentDialogMutation($input: AssessVendorInput!) {
    assessVendor(input: $input) {
      vendor {
        id
        name
        websiteUrl
        ...useVendorFormFragment
        ...VendorComplianceTabFragment
        ...VendorRiskAssessmentTabFragment
      }
    }
  }
`;

type Props = {
  vendorId: string;
  children: ReactNode;
};

export function ImportAssessmentDialog({ vendorId, children }: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const { register, handleSubmit, reset, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        url: "",
      },
    }
  );
  const [assess, isAssessing] = useMutationWithToasts(
    importAssessmentMutation,
    {
      successMessage: __("Vendor assessed successfully."),
      errorMessage: __("Failed to assess vendor. Please try again."),
    }
  );

  const onSubmit = handleSubmit((data) => {
    assess({
      variables: {
        input: {
          id: vendorId,
          websiteUrl: data.url,
        },
      },
      onSuccess: () => {
        dialogRef.current?.close();
        reset();
      },
    });
  });
  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={__("Assessment from website")}
      className="max-w-lg"
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded>
          <Field
            required
            label={__("URL")}
            type="text"
            {...register("url")}
            error={formState.errors.url?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isAssessing}>
            {__("Assess")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
