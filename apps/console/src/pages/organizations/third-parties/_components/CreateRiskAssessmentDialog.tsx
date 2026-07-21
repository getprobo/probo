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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();

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
            title: t("createThirdPartyRiskAssessmentDialog.messages.error"),
            description: formatError(
              t("createThirdPartyRiskAssessmentDialog.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("createThirdPartyRiskAssessmentDialog.messages.success"),
          description: t("createThirdPartyRiskAssessmentDialog.messages.created"),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: t("createThirdPartyRiskAssessmentDialog.messages.error"),
          description: formatError(
            t("createThirdPartyRiskAssessmentDialog.errors.create"),
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
          items={[t("createThirdPartyRiskAssessmentDialog.breadcrumb.riskAssessments"), t("createThirdPartyRiskAssessmentDialog.breadcrumb.newRiskAssessment")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <ControlledField
            label={t("createThirdPartyRiskAssessmentDialog.fields.dataSensitivity")}
            name="dataSensitivity"
            control={control}
            type="select"
          >
            <SentitivityOptions />
          </ControlledField>
          <ControlledField
            label={t("createThirdPartyRiskAssessmentDialog.fields.businessImpact")}
            name="businessImpact"
            control={control}
            type="select"
          >
            <ImpactOptions />
          </ControlledField>
          <Field
            label={t("createThirdPartyRiskAssessmentDialog.fields.notes")}
            {...register("notes")}
            type="textarea"
            error={formState.errors.notes?.message}
            help={t("createThirdPartyRiskAssessmentDialog.notesHelp")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {t("createThirdPartyRiskAssessmentDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
