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

import { formatDatetime, toDateInput } from "@probo/helpers";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { UpdateRiskAnalysisDialog_riskAnalysis$key } from "#/__generated__/core/UpdateRiskAnalysisDialog_riskAnalysis.graphql";
import type { UpdateRiskAnalysisDialogMutation } from "#/__generated__/core/UpdateRiskAnalysisDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const updateRiskAnalysisDialogFragment = graphql`
  fragment UpdateRiskAnalysisDialog_riskAnalysis on RiskAnalysis {
    id
    name
    description
    period {
      start
      end
    }
  }
`;

const updateMutation = graphql`
  mutation UpdateRiskAnalysisDialogMutation(
    $input: UpdateRiskAnalysisInput!
  ) {
    updateRiskAnalysis(input: $input) {
      riskAnalysis {
        id
        name
        description
        period {
          start
          end
        }
        matrixSize {
          rows
          cols
        }
        updatedAt
      }
    }
  }
`;

type FormData = {
  name: string;
  description: string;
  periodStart: string;
  periodEnd: string;
};

interface UpdateRiskAnalysisDialogProps {
  riskAnalysisKey: UpdateRiskAnalysisDialog_riskAnalysis$key;
  dialogRef: ReturnType<typeof useDialogRef>;
}

export function UpdateRiskAnalysisDialog({
  riskAnalysisKey,
  dialogRef,
}: UpdateRiskAnalysisDialogProps) {
  const { t } = useTranslation();
  const riskAnalysis = useFragment(updateRiskAnalysisDialogFragment, riskAnalysisKey);
  const [updateRiskAnalysis, isUpdating] = useMutation<UpdateRiskAnalysisDialogMutation>(updateMutation);
  const { register, handleSubmit, formState } = useForm<FormData>({
    values: {
      name: riskAnalysis.name,
      description: riskAnalysis.description ?? "",
      periodStart: toDateInput(riskAnalysis.period?.start),
      periodEnd: toDateInput(riskAnalysis.period?.end),
    },
  });

  const onSubmit = async (data: FormData) => {
    try {
      await updateRiskAnalysis({
        variables: {
          input: {
            id: riskAnalysis.id,
            name: data.name,
            description: data.description || null,
            period: {
              start: formatDatetime(data.periodStart) ?? null,
              end: formatDatetime(data.periodEnd) ?? null,
            },
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
      className="max-w-lg"
      ref={dialogRef}
      title={(
        <Breadcrumb
          items={[
            t("updateRiskAnalysisDialog.breadcrumb.riskAnalyses"),
            t("updateRiskAnalysisDialog.title"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={t("updateRiskAnalysisDialog.fields.name")}
            {...register("name", { required: t("updateRiskAnalysisDialog.validation.nameRequired") })}
            type="text"
            error={formState.errors.name?.message}
            placeholder={t("updateRiskAnalysisDialog.placeholders.name")}
          />
          <Field
            label={t("updateRiskAnalysisDialog.fields.description")}
            {...register("description")}
            type="textarea"
            rows={3}
            placeholder={t("updateRiskAnalysisDialog.placeholders.description")}
          />
          <Field
            label={t("updateRiskAnalysisDialog.fields.periodStart")}
            error={formState.errors.periodStart?.message}
          >
            <Input {...register("periodStart")} type="date" />
          </Field>
          <Field
            label={t("updateRiskAnalysisDialog.fields.periodEnd")}
            error={formState.errors.periodEnd?.message}
          >
            <Input
              {...register("periodEnd", {
                validate: (value, formValues) =>
                  !value
                  || !formValues.periodStart
                  || value >= formValues.periodStart
                  || t("updateRiskAnalysisDialog.validation.periodEndBeforeStart"),
              })}
              type="date"
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isUpdating}>
            {t("updateRiskAnalysisDialog.actions.save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
