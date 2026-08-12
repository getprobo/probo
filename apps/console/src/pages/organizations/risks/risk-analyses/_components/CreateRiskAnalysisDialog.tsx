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
  IconPlusLarge,
  Input,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";

import type { CreateRiskAnalysisDialogCreateMutation } from "#/__generated__/core/CreateRiskAnalysisDialogCreateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createMutation = graphql`
  mutation CreateRiskAnalysisDialogCreateMutation(
    $input: CreateRiskAnalysisInput!
    $connections: [ID!]!
  ) {
    createRiskAnalysis(input: $input) {
      riskAnalysisEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          period {
            start
            end
          }
          createdAt
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

export function CreateRiskAnalysisDialog(props: {
  connectionId: string;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const [createRiskAnalysis, isCreating] = useMutation<CreateRiskAnalysisDialogCreateMutation>(createMutation);
  const { register, handleSubmit, reset, formState } = useForm<FormData>({
    defaultValues: {
      name: "",
      description: "",
      periodStart: "",
      periodEnd: "",
    },
  });

  const onSubmit = (data: FormData) => {
    const periodStart = formatDatetime(data.periodStart);
    const periodEnd = formatDatetime(data.periodEnd);
    const period = periodStart || periodEnd
      ? {
          start: periodStart ?? null,
          end: periodEnd ?? null,
        }
      : null;

    createRiskAnalysis({
      variables: {
        input: {
          organizationId,
          name: data.name,
          description: data.description || null,
          period,
        },
        connections: [props.connectionId],
      },
      onCompleted: () => {
        reset();
        dialogRef.current?.close();
      },
    });
  };

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={(
        <Button icon={IconPlusLarge} variant="primary">
          {t("createRiskAnalysisDialog.title")}
        </Button>
      )}
      title={(
        <Breadcrumb
          items={[t("createRiskAnalysisDialog.breadcrumb.riskAnalyses"), t("createRiskAnalysisDialog.title")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={t("createRiskAnalysisDialog.fields.name")}
            {...register("name", { required: t("createRiskAnalysisDialog.validation.nameRequired") })}
            type="text"
            error={formState.errors.name?.message}
            placeholder={t("createRiskAnalysisDialog.placeholders.name")}
          />
          <Field
            label={t("createRiskAnalysisDialog.fields.description")}
            {...register("description")}
            type="textarea"
            rows={3}
            placeholder={t("createRiskAnalysisDialog.placeholders.description")}
          />
          <Field
            label={t("createRiskAnalysisDialog.fields.periodStart")}
            error={formState.errors.periodStart?.message}
          >
            <Input {...register("periodStart")} type="date" />
          </Field>
          <Field
            label={t("createRiskAnalysisDialog.fields.periodEnd")}
            error={formState.errors.periodEnd?.message}
          >
            <Input
              {...register("periodEnd", {
                validate: (value, formValues) =>
                  !value
                  || !formValues.periodStart
                  || value >= formValues.periodStart
                  || t("createRiskAnalysisDialog.validation.periodEndBeforeStart"),
              })}
              type="date"
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {t("createRiskAnalysisDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
