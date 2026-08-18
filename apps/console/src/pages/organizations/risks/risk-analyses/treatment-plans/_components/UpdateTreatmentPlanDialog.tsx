// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
  Breadcrumb,
  Button,
  Dialog,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { UpdateTreatmentPlanDialog_treatmentPlan$key } from "#/__generated__/core/UpdateTreatmentPlanDialog_treatmentPlan.graphql";
import type {
  RiskTreatment,
  UpdateTreatmentPlanDialogMutation,
} from "#/__generated__/core/UpdateTreatmentPlanDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { MatrixSize } from "../../_components/matrixSize";

import { TreatmentPlanDialogFields } from "./TreatmentPlanDialogFields";

export const updateTreatmentPlanDialogFragment = graphql`
  fragment UpdateTreatmentPlanDialog_treatmentPlan on TreatmentPlan {
    id
    treatment
    owner {
      id
    }
    inherentLikelihood
    inherentImpact
    residualLikelihood
    residualImpact
    risk {
      name
    }
  }
`;

const updateMutation = graphql`
  mutation UpdateTreatmentPlanDialogMutation($input: UpdateTreatmentPlanInput!) {
    updateTreatmentPlan(input: $input) {
      treatmentPlan {
        id
        netLikelihood
        netImpact
        netRiskScore
        ...UpdateTreatmentPlanDialog_treatmentPlan
        ...TreatmentPlanListItem_treatmentPlan
        ...RiskTreatmentPlanListItem_treatmentPlan
      }
    }
  }
`;

type FormData = {
  treatment: RiskTreatment;
  ownerId: string;
  inherentLikelihood: string;
  inherentImpact: string;
  residualLikelihood: string;
  residualImpact: string;
};

export function UpdateTreatmentPlanDialog({
  treatmentPlanKey,
  matrixSize,
  trigger,
  dialogRef: dialogRefFromParent,
}: {
  treatmentPlanKey: UpdateTreatmentPlanDialog_treatmentPlan$key;
  matrixSize: MatrixSize;
  trigger?: ReactNode;
  dialogRef?: ReturnType<typeof useDialogRef>;
}) {
  const { t } = useTranslation();
  const localDialogRef = useDialogRef();
  const dialogRef = dialogRefFromParent ?? localDialogRef;
  const treatmentPlan = useFragment(updateTreatmentPlanDialogFragment, treatmentPlanKey);
  const [updateTreatmentPlan, isUpdating]
    = useMutation<UpdateTreatmentPlanDialogMutation>(updateMutation);
  const { control, handleSubmit, setError } = useForm<FormData>({
    values: {
      treatment: treatmentPlan.treatment,
      ownerId: treatmentPlan.owner.id,
      inherentLikelihood: String(treatmentPlan.inherentLikelihood),
      inherentImpact: String(treatmentPlan.inherentImpact),
      residualLikelihood: String(treatmentPlan.residualLikelihood),
      residualImpact: String(treatmentPlan.residualImpact),
    },
  });

  const onSubmit = async (data: FormData) => {
    if (!data.ownerId) {
      setError("ownerId", {
        type: "required",
        message: t("updateTreatmentPlanDialog.validation.ownerRequired"),
      });
      return;
    }

    try {
      await updateTreatmentPlan({
        variables: {
          input: {
            id: treatmentPlan.id,
            treatment: data.treatment,
            ownerId: data.ownerId,
            inherentLikelihood: Number(data.inherentLikelihood),
            inherentImpact: Number(data.inherentImpact),
            residualLikelihood: Number(data.residualLikelihood),
            residualImpact: Number(data.residualImpact),
          },
        },
      });
      dialogRef.current?.close();
    } catch {
      // Error toast is handled by useMutation.
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={trigger}
      title={(
        <Breadcrumb
          items={[
            t("updateTreatmentPlanDialog.breadcrumb.treatmentPlans"),
            t("updateTreatmentPlanDialog.breadcrumb.edit"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <TreatmentPlanDialogFields
          control={control}
          copy="updateTreatmentPlanDialog"
          matrixSize={matrixSize}
        >
          <Field
            disabled
            label={t("updateTreatmentPlanDialog.fields.risk")}
            type="text"
            value={treatmentPlan.risk.name}
          />
        </TreatmentPlanDialogFields>
        <DialogFooter>
          <Button type="submit" disabled={isUpdating}>
            {t("updateTreatmentPlanDialog.actions.update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
