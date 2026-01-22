import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  ImpactOptions,
  SentitivityOptions,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { ControlledField } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

type Props = {
  children: ReactNode;
  connection: string;
  vendorId: string;
};

const createRiskAssessmentMutation = graphql`
  mutation CreateRiskAssessmentDialogMutation(
    $input: CreateVendorRiskAssessmentInput!
    $connections: [ID!]!
  ) {
    createVendorRiskAssessment(input: $input) {
      vendorRiskAssessmentEdge @prependEdge(connections: $connections) {
        node {
          ...VendorRiskAssessmentTabFragment_assessment
        }
      }
    }
  }
`;

const schema = z.object({
  dataSensitivity: z.enum(["NONE", "LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  businessImpact: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  notes: z.string().nullable().optional(),
});

/**
 * Dialog to create or update a riskassessment
 */
export function CreateRiskAssessmentDialog({
  children,
  connection,
  vendorId,
}: Props) {
  const { __ } = useTranslate();

  const { register, handleSubmit, formState, reset, control }
    = useFormWithSchema(schema, {
      defaultValues: {
        dataSensitivity: "LOW",
        businessImpact: "LOW",
      },
    });
  const [createRiskAssessment, isLoading] = useMutationWithToasts(
    createRiskAssessmentMutation,
    {
      successMessage: __("Risk Assessment created successfully."),
      errorMessage: __("Failed to create Risk Assessment"),
    },
  );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const nextYear = new Date();
    nextYear.setFullYear(nextYear.getFullYear() + 1);
    await createRiskAssessment({
      variables: {
        input: {
          ...data,
          notes: data.notes || null,
          vendorId,
          expiresAt: nextYear.toISOString(),
        },
        connections: [connection],
      },
      onSuccess: () => {
        dialogRef.current?.close();
        reset();
      },
    });
  };

  const dialogRef = useDialogRef();

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={children}
      title={(
        <Breadcrumb
          items={[__("Risk Assessments"), __("New Risk Assessment")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <ControlledField
            label={__("Data Sensitivity")}
            name="dataSensitivity"
            control={control}
            type="select"
          >
            <SentitivityOptions />
          </ControlledField>
          <ControlledField
            label={__("Business Impact")}
            name="businessImpact"
            control={control}
            type="select"
          >
            <ImpactOptions />
          </ControlledField>
          <Field
            label={__("Notes")}
            {...register("notes")}
            type="textarea"
            error={formState.errors.notes?.message}
            help={__(
              "Add any context or details about this risk assessment that might be helpful for future reference.",
            )}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
