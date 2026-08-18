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
  DialogContent,
  Label,
  Option,
  PropertyRow,
} from "@probo/ui";
import { type ReactNode } from "react";
import { type Control, type FieldValues, type Path, useFormState } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { ControlledSelect } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import type { MatrixSize } from "../../_components/matrixSize";

import { TreatmentPlanScoreFields } from "./TreatmentPlanScoreFields";

export type TreatmentPlanDialogValues = {
  treatment: string;
  ownerId: string;
  inherentLikelihood: string;
  inherentImpact: string;
  residualLikelihood: string;
  residualImpact: string;
};

type CopyNamespace = "createTreatmentPlanDialog" | "updateTreatmentPlanDialog";

interface TreatmentPlanDialogFieldsProps<
  TFieldValues extends FieldValues & TreatmentPlanDialogValues,
> {
  children?: ReactNode;
  control: Control<TFieldValues>;
  copy: CopyNamespace;
  disabled?: boolean;
  matrixSize: MatrixSize;
}

export function TreatmentPlanDialogFields<
  TFieldValues extends FieldValues & TreatmentPlanDialogValues,
>({
  children,
  control,
  copy,
  disabled,
  matrixSize,
}: TreatmentPlanDialogFieldsProps<TFieldValues>) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { errors } = useFormState({ control });

  return (
    <DialogContent className="grid grid-cols-[1fr_420px]">
      <div className="space-y-6 px-12 py-8">
        {children}
        <div className="grid grid-cols-2 gap-6">
          <TreatmentPlanScoreFields
            control={control}
            disabled={disabled}
            label={t(`${copy}.fields.inherentRisk`)}
            matrixSize={matrixSize}
            prefix="inherent"
          />
          <TreatmentPlanScoreFields
            control={control}
            disabled={disabled}
            label={t(`${copy}.fields.residualRisk`)}
            matrixSize={matrixSize}
            prefix="residual"
          />
        </div>
      </div>
      <div className="bg-subtle px-6 py-5">
        <Label>{t(`${copy}.properties`)}</Label>
        <PropertyRow
          id="treatment"
          label={t(`${copy}.fields.treatment`)}
        >
          <ControlledSelect
            control={control}
            disabled={disabled}
            name={"treatment" as Path<TFieldValues>}
            placeholder={t(`${copy}.placeholders.treatment`)}
            variant="editor"
          >
            <Option value="MITIGATED">{t("formRiskDialog.treatments.mitigated")}</Option>
            <Option value="ACCEPTED">{t("formRiskDialog.treatments.accepted")}</Option>
            <Option value="AVOIDED">{t("formRiskDialog.treatments.avoided")}</Option>
            <Option value="TRANSFERRED">{t("formRiskDialog.treatments.transferred")}</Option>
          </ControlledSelect>
        </PropertyRow>
        <PropertyRow
          id="ownerId"
          label={t(`${copy}.fields.owner`)}
          error={typeof errors.ownerId?.message === "string"
            ? errors.ownerId.message
            : undefined}
        >
          <PeopleSelectField
            control={control}
            disabled={disabled}
            name={"ownerId" as Path<TFieldValues>}
            organizationId={organizationId}
          />
        </PropertyRow>
      </div>
    </DialogContent>
  );
}
