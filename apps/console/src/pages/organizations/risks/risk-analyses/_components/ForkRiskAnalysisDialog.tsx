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

import { formatDatetime } from "@probo/helpers";
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

import type { ForkRiskAnalysisDialog_riskAnalysis$key } from "#/__generated__/core/ForkRiskAnalysisDialog_riskAnalysis.graphql";
import type { ForkRiskAnalysisDialogMutation } from "#/__generated__/core/ForkRiskAnalysisDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { formatMatrixSize } from "./matrixSize";

export const forkRiskAnalysisDialogFragment = graphql`
  fragment ForkRiskAnalysisDialog_riskAnalysis on RiskAnalysis {
    id
    name
    description
    matrixSize {
      rows
      cols
    }
  }
`;

const forkMutation = graphql`
  mutation ForkRiskAnalysisDialogMutation(
    $input: ForkRiskAnalysisInput!
    $connections: [ID!]!
  ) {
    forkRiskAnalysis(input: $input) {
      riskAnalysisEdge @prependEdge(connections: $connections) {
        node {
          id
          ...RiskAnalysisListItem_riskAnalysis
        }
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

interface ForkRiskAnalysisDialogProps {
  riskAnalysisKey: ForkRiskAnalysisDialog_riskAnalysis$key;
  connectionId: string;
  dialogRef?: ReturnType<typeof useDialogRef>;
}

export function ForkRiskAnalysisDialog({
  riskAnalysisKey,
  connectionId,
  dialogRef: dialogRefFromParent,
}: ForkRiskAnalysisDialogProps) {
  const { t } = useTranslation();
  const localDialogRef = useDialogRef();
  const dialogRef = dialogRefFromParent ?? localDialogRef;
  const riskAnalysis = useFragment(forkRiskAnalysisDialogFragment, riskAnalysisKey);
  const [forkRiskAnalysis, isForking] = useMutation<ForkRiskAnalysisDialogMutation>(forkMutation);
  const { register, handleSubmit, formState } = useForm<FormData>({
    values: {
      name: riskAnalysis.name,
      description: riskAnalysis.description ?? "",
      periodStart: "",
      periodEnd: "",
    },
  });

  const onSubmit = async (data: FormData) => {
    const periodStart = formatDatetime(data.periodStart);
    const periodEnd = formatDatetime(data.periodEnd);
    const period = periodStart || periodEnd
      ? {
          start: periodStart ?? null,
          end: periodEnd ?? null,
        }
      : null;

    try {
      await forkRiskAnalysis({
        variables: {
          input: {
            riskAnalysisId: riskAnalysis.id,
            name: data.name,
            description: data.description || null,
            period,
          },
          connections: [connectionId],
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
            t("forkRiskAnalysisDialog.breadcrumb.riskAnalyses"),
            t("forkRiskAnalysisDialog.title"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={t("forkRiskAnalysisDialog.fields.name")}
            {...register("name", { required: t("forkRiskAnalysisDialog.validation.nameRequired") })}
            type="text"
            error={formState.errors.name?.message}
            placeholder={t("forkRiskAnalysisDialog.placeholders.name")}
          />
          <Field
            label={t("forkRiskAnalysisDialog.fields.description")}
            {...register("description")}
            type="textarea"
            rows={3}
            placeholder={t("forkRiskAnalysisDialog.placeholders.description")}
          />
          <Field
            disabled
            label={t("forkRiskAnalysisDialog.fields.matrixSize")}
            type="text"
            value={formatMatrixSize(
              riskAnalysis.matrixSize.rows,
              riskAnalysis.matrixSize.cols,
            )}
          />
          <Field
            label={t("forkRiskAnalysisDialog.fields.periodStart")}
            error={formState.errors.periodStart?.message}
          >
            <Input {...register("periodStart")} type="date" />
          </Field>
          <Field
            label={t("forkRiskAnalysisDialog.fields.periodEnd")}
            error={formState.errors.periodEnd?.message}
          >
            <Input
              {...register("periodEnd", {
                validate: (value, formValues) =>
                  !value
                  || !formValues.periodStart
                  || value >= formValues.periodStart
                  || t("forkRiskAnalysisDialog.validation.periodEndBeforeStart"),
              })}
              type="date"
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isForking}>
            {t("forkRiskAnalysisDialog.actions.fork")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
