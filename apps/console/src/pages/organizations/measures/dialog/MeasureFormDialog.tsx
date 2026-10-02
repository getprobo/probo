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

import {
  internalControlCadences,
  internalControlImplementationStatuses,
  internalControlNatures,
  internalControlTypes,
  measureStates,
} from "@probo/helpers";
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

import type { MeasureFormDialogCreateMutation } from "#/__generated__/core/MeasureFormDialogCreateMutation.graphql";
import type { MeasureFormDialogMeasureFragment$key } from "#/__generated__/core/MeasureFormDialogMeasureFragment.graphql";
import type { MeasureGraphUpdateMutation$variables } from "#/__generated__/core/MeasureGraphUpdateMutation.graphql";
import { ControlledSelect } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useUpdateMeasure } from "#/hooks/graph/MeasureGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { z } from "#/lib/zod";

const blankableEnum = <T extends readonly [string, ...string[]]>(values: T) =>
  z.union([z.enum(values), z.literal("")]);

const measureFragment = graphql`
  fragment MeasureFormDialogMeasureFragment on Measure {
    id
    description
    name
    category
    state
    code
    controlType
    nature
    operatingFrequency
    evidenceCadence
    testingCadence
    implementationStatus
    owner {
      id
    }
    reviewer {
      id
    }
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
          id
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
  onCreated?: (measureId: string) => void | Promise<void>;
};

export default function MeasureFormDialog(props: Props) {
  const { children, measure: measureKey, connection, onCreated, ...rest } = props;
  const { t } = useTranslation();
  const ref = useDialogRef();
  const dialogRef = rest.ref ?? ref;
  const measure = useFragment(measureFragment, measureKey);
  const organizationId = useOrganizationId();
  const [updateMeasure] = useUpdateMeasure();
  const [createMeasure] = useMutationWithToasts<MeasureFormDialogCreateMutation>(
    measureCreateMutation,
    {
      successMessage: t("measureFormDialog.messages.created"),
      errorMessage: t("measureFormDialog.errors.create"),
    },
  );
  const measureSchema = z.object({
    name: z.string().min(1, t("measureFormDialog.validation.nameRequired")),
    description: z.string().optional().nullable(),
    category: z.string().min(1, t("measureFormDialog.validation.categoryRequired")),
    state: z.enum(measureStates),
    code: z.string().optional().nullable(),
    controlType: blankableEnum(internalControlTypes),
    nature: blankableEnum(internalControlNatures),
    operatingFrequency: blankableEnum(internalControlCadences),
    evidenceCadence: blankableEnum(internalControlCadences),
    testingCadence: blankableEnum(internalControlCadences),
    implementationStatus: z.enum(internalControlImplementationStatuses),
    ownerId: z.string().optional().nullable(),
    reviewerId: z.string().optional().nullable(),
  }).superRefine((value, ctx) => {
    if (value.ownerId && value.reviewerId && value.ownerId === value.reviewerId) {
      ctx.addIssue({
        code: "custom",
        path: ["reviewerId"],
        message: t("measureFormDialog.validation.reviewerDistinct"),
      });
    }
  });

  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(measureSchema, {
      values: {
        name: measure?.name ?? "",
        description: measure?.description ?? "",
        category: measure?.category ?? "",
        state: measure?.state ?? "NOT_STARTED",
        code: measure?.code ?? "",
        controlType: measure?.controlType ?? "",
        nature: measure?.nature ?? "",
        operatingFrequency: measure?.operatingFrequency ?? "",
        evidenceCadence: measure?.evidenceCadence ?? "",
        testingCadence: measure?.testingCadence ?? "",
        implementationStatus: measure?.implementationStatus ?? "NOT_IMPLEMENTED",
        ownerId: measure?.owner?.id ?? "",
        reviewerId: measure?.reviewer?.id ?? "",
      },
    });

  const onSubmit = async (data: z.infer<typeof measureSchema>) => {
    const fields = {
      code: data.code || null,
      controlType: data.controlType || null,
      nature: data.nature || null,
      operatingFrequency: data.operatingFrequency || null,
      evidenceCadence: data.evidenceCadence || null,
      testingCadence: data.testingCadence || null,
      ownerId: data.ownerId || null,
      reviewerId: data.reviewerId || null,
    };
    if (measure) {
      const input: MeasureGraphUpdateMutation$variables["input"] = {
        id: measure.id,
        name: data.name,
        description: data.description || null,
        category: data.category,
        ...fields,
      };
      // Send only the status that changed. Repeating both would collapse an
      // operating control back to implemented, or a not-started control into
      // not implemented.
      if (data.implementationStatus !== measure.implementationStatus) {
        input.implementationStatus = data.implementationStatus;
      } else if (data.state !== measure.state) {
        input.state = data.state;
      }
      await updateMeasure({
        variables: {
          input,
        },
      });
    } else {
      let createdId: string | undefined;
      await createMeasure({
        variables: {
          input: {
            organizationId,
            name: data.name,
            description: data.description || null,
            category: data.category,
            implementationStatus: data.implementationStatus,
            ...fields,
          },
          connections: connection ? [connection] : [],
        },
        onCompleted: (response, errors) => {
          if (!errors) {
            createdId = response.createMeasure.measureEdge.node.id;
          }
        },
      });
      if (createdId) {
        try {
          await onCreated?.(createdId);
        } catch {
          // Linking errors are reported by the caller.
        }
      }
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
            <PropertyRow
              label={t("measureFormDialog.fields.code")}
              error={formState.errors.code?.message}
            >
              <Input
                {...register("code")}
                placeholder={t("measureFormDialog.fields.codePlaceholder")}
              />
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.controlType")}
              error={formState.errors.controlType?.message}
            >
              <ControlledSelect control={control} name="controlType">
                <Option value="">{t("measureFormDialog.fields.notSet")}</Option>
                {internalControlTypes.map(value => (
                  <Option key={value} value={value}>
                    {t(`measureFormDialog.controlTypes.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.nature")}
              error={formState.errors.nature?.message}
            >
              <ControlledSelect control={control} name="nature">
                <Option value="">{t("measureFormDialog.fields.notSet")}</Option>
                {internalControlNatures.map(value => (
                  <Option key={value} value={value}>
                    {t(`measureFormDialog.natures.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.operatingFrequency")}
              error={formState.errors.operatingFrequency?.message}
            >
              <ControlledSelect control={control} name="operatingFrequency">
                <Option value="">{t("measureFormDialog.fields.notSet")}</Option>
                {internalControlCadences.map(value => (
                  <Option key={value} value={value}>
                    {t(`measureFormDialog.cadences.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.evidenceCadence")}
              error={formState.errors.evidenceCadence?.message}
            >
              <ControlledSelect control={control} name="evidenceCadence">
                <Option value="">{t("measureFormDialog.fields.notSet")}</Option>
                {internalControlCadences.map(value => (
                  <Option key={value} value={value}>
                    {t(`measureFormDialog.cadences.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.testingCadence")}
              error={formState.errors.testingCadence?.message}
            >
              <ControlledSelect control={control} name="testingCadence">
                <Option value="">{t("measureFormDialog.fields.notSet")}</Option>
                {internalControlCadences.map(value => (
                  <Option key={value} value={value}>
                    {t(`measureFormDialog.cadences.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.implementationStatus")}
              error={formState.errors.implementationStatus?.message}
            >
              <ControlledSelect control={control} name="implementationStatus">
                {internalControlImplementationStatuses.map(value => (
                  <Option key={value} value={value}>
                    {t(`measureFormDialog.implementationStatuses.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.owner")}
              error={formState.errors.ownerId?.message}
            >
              <PeopleSelectField
                name="ownerId"
                control={control}
                organizationId={organizationId}
                optional
              />
            </PropertyRow>
            <PropertyRow
              label={t("measureFormDialog.fields.reviewer")}
              error={formState.errors.reviewerId?.message}
            >
              <PeopleSelectField
                name="reviewerId"
                control={control}
                organizationId={organizationId}
                optional
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
