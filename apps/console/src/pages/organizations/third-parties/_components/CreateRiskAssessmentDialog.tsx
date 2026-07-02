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

import { formatError } from "@probo/helpers";
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
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateRiskAssessmentDialogMutation } from "#/__generated__/core/CreateRiskAssessmentDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  children: ReactNode;
  connection: string;
  thirdPartyId: string;
};

const createRiskAssessmentMutation = graphql`
  mutation CreateRiskAssessmentDialogMutation(
    $input: CreateThirdPartyRiskAssessmentInput!
    $connections: [ID!]!
  ) {
    createThirdPartyRiskAssessment(input: $input) {
      thirdPartyRiskAssessmentEdge @prependEdge(connections: $connections) {
        node {
          ...ThirdPartyRiskAssessmentRow_assessment
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
  thirdPartyId,
}: Props) {
  const { __ } = useTranslate();

  const { register, handleSubmit, formState, reset, control }
    = useFormWithSchema(schema, {
      defaultValues: {
        dataSensitivity: "LOW",
        businessImpact: "LOW",
      },
    });
  const { toast } = useToast();
  const [createRiskAssessment, isCreating]
    = useMutation<CreateRiskAssessmentDialogMutation>(
      createRiskAssessmentMutation,
    );

  const onSubmit = (data: z.infer<typeof schema>) => {
    const nextYear = new Date();
    nextYear.setFullYear(nextYear.getFullYear() + 1);
    createRiskAssessment({
      variables: {
        input: {
          ...data,
          notes: data.notes || null,
          thirdPartyId,
          expiresAt: nextYear.toISOString(),
        },
        connections: [connection],
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to create Risk Assessment"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Risk Assessment created successfully."),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create Risk Assessment"),
            error,
          ),
          variant: "error",
        });
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
          <Button type="submit" disabled={isCreating}>
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
