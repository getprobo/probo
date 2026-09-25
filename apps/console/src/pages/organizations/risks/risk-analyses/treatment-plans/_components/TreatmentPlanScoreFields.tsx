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

import { getRiskImpacts, getRiskLikelihoods } from "@probo/helpers";
import { Card, Label, Option } from "@probo/ui";
import { type Control, type FieldValues, type Path } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { ControlledField } from "#/components/form/ControlledField";

import type { MatrixSize } from "../../_components/matrixSize";

import type { TreatmentPlanDialogValues } from "./TreatmentPlanDialogFields";

interface TreatmentPlanScoreFieldsProps<
  TFieldValues extends FieldValues & TreatmentPlanDialogValues,
> {
  control: Control<TFieldValues>;
  disabled?: boolean;
  label: string;
  matrixSize: MatrixSize;
  prefix: "inherent" | "residual";
}

export function TreatmentPlanScoreFields<
  TFieldValues extends FieldValues & TreatmentPlanDialogValues,
>({
  control,
  disabled,
  label,
  matrixSize,
  prefix,
}: TreatmentPlanScoreFieldsProps<TFieldValues>) {
  const { t } = useTranslation();
  const impacts = getRiskImpacts(t, matrixSize.cols);
  const likelihoods = getRiskLikelihoods(t, matrixSize.rows);

  return (
    <div>
      <Label>{label}</Label>
      <Card padded className="space-y-4 p-4">
        <ControlledField
          control={control}
          disabled={disabled}
          label={t("formRiskDialog.fields.impact")}
          name={`${prefix}Impact` as Path<TFieldValues>}
          placeholder={t("formRiskDialog.placeholders.impact")}
          type="select"
        >
          {impacts.map(impact => (
            <Option key={impact.value} value={impact.value.toString()}>
              {t("formRiskDialog.scoreOption", {
                value: impact.value,
                label: impact.label,
              })}
            </Option>
          ))}
        </ControlledField>
        <ControlledField
          control={control}
          disabled={disabled}
          label={t("formRiskDialog.fields.likelihood")}
          name={`${prefix}Likelihood` as Path<TFieldValues>}
          placeholder={t("formRiskDialog.placeholders.likelihood")}
          type="select"
        >
          {likelihoods.map(likelihood => (
            <Option key={likelihood.value} value={likelihood.value.toString()}>
              {t("formRiskDialog.scoreOption", {
                value: likelihood.value,
                label: likelihood.label,
              })}
            </Option>
          ))}
        </ControlledField>
      </Card>
    </div>
  );
}
