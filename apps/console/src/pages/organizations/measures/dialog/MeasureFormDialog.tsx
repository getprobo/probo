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

import { measureStates } from "@probo/helpers";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  type DialogRef,
  Field,
  Input,
  Label,
  Option,
  PropertyRow,
  useDialogRef,
} from "@probo/ui";
import { Breadcrumb } from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { MeasureFormDialogMeasureFragment$key } from "#/__generated__/core/MeasureFormDialogMeasureFragment.graphql";
import { ControlledSelect } from "#/components/form/ControlledField";
import { useUpdateMeasure } from "#/hooks/graph/MeasureGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const measureFragment = graphql`
  fragment MeasureFormDialogMeasureFragment on Measure {
    id
    description
    name
    category
    state
  }
`;

const measureCreateMutation = graphql`
  mutation MeasureFormDialogCreateMutation(
    $input: CreateMeasureInput!
    $connections: [ID!]!
  ) {
    createMeasure(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          ...MeasureFormDialogMeasureFragment
        }
      }
    }
  }
`;

type Props = {
  children?: ReactNode;
  measure?: MeasureFormDialogMeasureFragment$key;
  connection?: string;
  ref?: DialogRef;
};

export default function MeasureFormDialog(props: Props) {
  const { children, measure: measureKey, connection, ...rest } = props;
  const { t } = useTranslation();
  const ref = useDialogRef();
  const dialogRef = rest.ref ?? ref;
  const measure = useFragment(measureFragment, measureKey);
  const organizationId = useOrganizationId();
  const [updateMeasure] = useUpdateMeasure();
  const [createMeasure] = useMutationWithToasts(measureCreateMutation, {
    successMessage: t("measureFormDialog.messages.created"),
    errorMessage: t("measureFormDialog.errors.create"),
  });
  const mutate = measureKey ? updateMeasure : createMeasure;
  const measureSchema = z.object({
    name: z.string().min(1, t("measureFormDialog.validation.nameRequired")),
    description: z.string().optional().nullable(),
    category: z.string().min(1, t("measureFormDialog.validation.categoryRequired")),
    state: z.enum(measureStates),
  });

  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(measureSchema, {
      values: {
        name: measure?.name ?? "",
        description: measure?.description ?? "",
        category: measure?.category ?? "",
        state: measure?.state ?? "NOT_STARTED",
      },
    });

  const onSubmit = async (data: z.infer<typeof measureSchema>) => {
    if (measure) {
      await mutate({
        variables: {
          input: {
            id: measure.id,
            name: data.name,
            description: data.description || null,
            category: data.category,
            state: data.state,
          },
        },
      });
    } else {
      await mutate({
        variables: {
          input: {
            organizationId,
            name: data.name,
            description: data.description || null,
            category: data.category,
          },
          connections: [connection!],
        },
      });
      reset();
    }
    dialogRef.current?.close();
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={(
        <Breadcrumb
          items={[
            t("measureFormDialog.breadcrumb.measures"),
            measure ? t("measureFormDialog.breadcrumb.editMeasure") : t("measureFormDialog.breadcrumb.newMeasure"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-6">
            <Field
              {...register("name")}
              error={formState.errors.name?.message}
              label={t("measureFormDialog.fields.name")}
              placeholder={t("measureFormDialog.fields.namePlaceholder")}
              required
            />
            <Field
              {...register("description")}
              error={formState.errors.description?.message}
              label={t("measureFormDialog.fields.description")}
              placeholder={t("measureFormDialog.fields.descriptionPlaceholder")}
              type="textarea"
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{t("measureFormDialog.properties")}</Label>
            <PropertyRow
              label={t("measureFormDialog.fields.category")}
              error={formState.errors.category?.message}
            >
              <Input
                {...register("category")}
                required
                placeholder={t("measureFormDialog.fields.categoryPlaceholder")}
              />
            </PropertyRow>
            {measure && (
              <PropertyRow
                label={t("measureFormDialog.fields.state")}
                error={formState.errors.state?.message}
              >
                <ControlledSelect
                  control={control}
                  name="state"
                  placeholder={t("measureFormDialog.fields.statePlaceholder")}
                >
                  {measureStates.map(state => (
                    <Option key={state} value={state}>
                      {t(`measureFormDialog.states.${state.toLowerCase()}`)}
                    </Option>
                  ))}
                </ControlledSelect>
              </PropertyRow>
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">
            {measure ? t("measureFormDialog.actions.update") : t("measureFormDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
