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
  ActionDropdown,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Field,
  IconPencil,
  IconTrashCan,
  Input,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";

import type { UpdateRiskAnalysisDialogMutation } from "#/__generated__/core/UpdateRiskAnalysisDialogMutation.graphql";

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

export function UpdateRiskAnalysisDialog(props: {
  riskAnalysis: {
    id: string;
    name: string;
    description: string | null | undefined;
    period: {
      start: string | null | undefined;
      end: string | null | undefined;
    } | null | undefined;
  };
  canDelete?: boolean;
  onDelete?: () => void;
}) {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [updateRiskAnalysis, isUpdating] = useMutation<UpdateRiskAnalysisDialogMutation>(updateMutation);
  const { register, handleSubmit, formState } = useForm<FormData>({
    values: {
      name: props.riskAnalysis.name,
      description: props.riskAnalysis.description ?? "",
      periodStart: toDateInput(props.riskAnalysis.period?.start),
      periodEnd: toDateInput(props.riskAnalysis.period?.end),
    },
  });

  const onSubmit = (data: FormData) => {
    updateRiskAnalysis({
      variables: {
        input: {
          id: props.riskAnalysis.id,
          name: data.name,
          description: data.description || null,
          period: {
            start: formatDatetime(data.periodStart) ?? null,
            end: formatDatetime(data.periodEnd) ?? null,
          },
        },
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
    });
  };

  return (
    <>
      <ActionDropdown variant="secondary">
        <DropdownItem
          icon={IconPencil}
          onSelect={() => dialogRef.current?.open()}
        >
          {t("riskAnalysisDetailPage.actions.edit")}
        </DropdownItem>
        {props.canDelete && props.onDelete && (
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onClick={props.onDelete}
          >
            {t("riskAnalysisDetailPage.actions.delete")}
          </DropdownItem>
        )}
      </ActionDropdown>
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
    </>
  );
}
