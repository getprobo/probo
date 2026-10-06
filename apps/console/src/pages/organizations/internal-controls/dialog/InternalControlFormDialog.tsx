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
  internalControlImplementationStatuses,
  internalControlNatures,
  internalControlOperatingModes,
  internalControlStates,
  internalControlTypes,
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
import { type ReactNode, Suspense } from "react";
import { Controller, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { InternalControlFormDialogCreateMutation } from "#/__generated__/core/InternalControlFormDialogCreateMutation.graphql";
import type { InternalControlFormDialogInternalControlFragment$key } from "#/__generated__/core/InternalControlFormDialogInternalControlFragment.graphql";
import type { InternalControlGraphUpdateMutation$variables } from "#/__generated__/core/InternalControlGraphUpdateMutation.graphql";
import { ControlledSelect } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useUpdateInternalControl } from "#/hooks/graph/InternalControlGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { z } from "#/lib/zod";
import { TaskDurationField } from "#/pages/organizations/tasks/_components/TaskDurationField";

const controlDurationUnits = ["D", "W", "MO", "Y"] as const;

const blankableEnum = <T extends readonly [string, ...string[]]>(values: T) =>
  z.union([z.enum(values), z.literal("")]);

const internalControlFragment = graphql`
  fragment InternalControlFormDialogInternalControlFragment on InternalControl {
    id
    description
    name
    category
    state
    code
    controlType
    nature
    operatingFrequency {
      mode
      interval
      event
    }
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

const internalControlCreateMutation = graphql`
  mutation InternalControlFormDialogCreateMutation(
    $input: CreateInternalControlInput!
    $connections: [ID!]!
  ) {
    createInternalControl(input: $input) {
      internalControlEdge @prependEdge(connections: $connections) {
        node {
          id
          ...InternalControlFormDialogInternalControlFragment
        }
      }
    }
  }
`;

type Props = {
  children?: ReactNode;
  internalControl?: InternalControlFormDialogInternalControlFragment$key;
  connection?: string;
  ref?: DialogRef;
  onCreated?: (internalControlId: string) => void | Promise<void>;
};

export default function InternalControlFormDialog(props: Props) {
  const { children, internalControl: internalControlKey, connection, onCreated, ...rest } = props;
  const { t } = useTranslation();
  const ref = useDialogRef();
  const dialogRef = rest.ref ?? ref;
  const internalControl = useFragment(internalControlFragment, internalControlKey);
  const organizationId = useOrganizationId();
  const [updateInternalControl] = useUpdateInternalControl();
  const [createInternalControl] = useMutationWithToasts<InternalControlFormDialogCreateMutation>(
    internalControlCreateMutation,
    {
      successMessage: t("internalControlFormDialog.messages.created"),
      errorMessage: t("internalControlFormDialog.errors.create"),
    },
  );
  const internalControlSchema = z.object({
    name: z.string().min(1, t("internalControlFormDialog.validation.nameRequired")),
    description: z.string().optional().nullable(),
    category: z.string().min(1, t("internalControlFormDialog.validation.categoryRequired")),
    state: z.enum(internalControlStates),
    code: z.string().optional().nullable(),
    controlType: blankableEnum(internalControlTypes),
    nature: blankableEnum(internalControlNatures),
    operatingMode: blankableEnum(internalControlOperatingModes),
    operatingInterval: z.string().nullable(),
    operatingEvent: z.string(),
    evidenceCadence: z.string().nullable(),
    testingCadence: z.string().nullable(),
    implementationStatus: z.enum(internalControlImplementationStatuses),
    ownerId: z.string().optional().nullable(),
    reviewerId: z.string().optional().nullable(),
  }).superRefine((value, ctx) => {
    if (value.ownerId && value.reviewerId && value.ownerId === value.reviewerId) {
      ctx.addIssue({
        code: "custom",
        path: ["reviewerId"],
        message: t("internalControlFormDialog.validation.reviewerDistinct"),
      });
    }
    if (value.operatingMode === "PERIODIC" && !value.operatingInterval) {
      ctx.addIssue({
        code: "custom",
        path: ["operatingInterval"],
        message: t("internalControlFormDialog.validation.operatingIntervalRequired"),
      });
    }
  });

  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(internalControlSchema, {
      values: {
        name: internalControl?.name ?? "",
        description: internalControl?.description ?? "",
        category: internalControl?.category ?? "",
        state: internalControl?.state ?? "NOT_STARTED",
        code: internalControl?.code ?? "",
        controlType: internalControl?.controlType ?? "",
        nature: internalControl?.nature ?? "",
        operatingMode: internalControl?.operatingFrequency?.mode ?? "",
        operatingInterval: internalControl?.operatingFrequency?.interval ?? null,
        operatingEvent: internalControl?.operatingFrequency?.event ?? "",
        evidenceCadence: internalControl?.evidenceCadence ?? null,
        testingCadence: internalControl?.testingCadence ?? null,
        implementationStatus: internalControl?.implementationStatus ?? "NOT_IMPLEMENTED",
        ownerId: internalControl?.owner?.id ?? "",
        reviewerId: internalControl?.reviewer?.id ?? "",
      },
    });

  const operatingMode = useWatch({ control, name: "operatingMode" });

  const onSubmit = async (data: z.infer<typeof internalControlSchema>) => {
    const fields = {
      code: data.code || null,
      controlType: data.controlType || null,
      nature: data.nature || null,
      operatingFrequency: operatingFrequencyInput(data),
      evidenceCadence: data.evidenceCadence || null,
      testingCadence: data.testingCadence || null,
      ownerId: data.ownerId || null,
      reviewerId: data.reviewerId || null,
    };
    if (internalControl) {
      const input: InternalControlGraphUpdateMutation$variables["input"] = {
        id: internalControl.id,
        name: data.name,
        description: data.description || null,
        category: data.category,
        ...fields,
      };
      // Send only the status that changed. Repeating both would collapse an
      // operating control back to implemented, or a not-started control into
      // not implemented.
      if (data.implementationStatus !== internalControl.implementationStatus) {
        input.implementationStatus = data.implementationStatus;
      } else if (data.state !== internalControl.state) {
        input.state = data.state;
      }
      await updateInternalControl({
        variables: {
          input,
        },
      });
    } else {
      let createdId: string | undefined;
      await createInternalControl({
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
            createdId = response.createInternalControl.internalControlEdge.node.id;
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
            t("internalControlFormDialog.breadcrumb.internalControls"),
            internalControl ? t("internalControlFormDialog.breadcrumb.editInternalControl") : t("internalControlFormDialog.breadcrumb.newInternalControl"),
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
              label={t("internalControlFormDialog.fields.name")}
              placeholder={t("internalControlFormDialog.fields.namePlaceholder")}
              required
            />
            <Field
              {...register("description")}
              error={formState.errors.description?.message}
              label={t("internalControlFormDialog.fields.description")}
              placeholder={t("internalControlFormDialog.fields.descriptionPlaceholder")}
              type="textarea"
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{t("internalControlFormDialog.properties")}</Label>
            <PropertyRow
              label={t("internalControlFormDialog.fields.category")}
              error={formState.errors.category?.message}
            >
              <Input
                {...register("category")}
                required
                placeholder={t("internalControlFormDialog.fields.categoryPlaceholder")}
              />
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.code")}
              error={formState.errors.code?.message}
            >
              <Input
                {...register("code")}
                placeholder={t("internalControlFormDialog.fields.codePlaceholder")}
              />
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.controlType")}
              error={formState.errors.controlType?.message}
            >
              <ControlledSelect control={control} name="controlType">
                <Option value="">{t("internalControlFormDialog.fields.notSet")}</Option>
                {internalControlTypes.map(value => (
                  <Option key={value} value={value}>
                    {t(`internalControlFormDialog.controlTypes.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.nature")}
              error={formState.errors.nature?.message}
            >
              <ControlledSelect control={control} name="nature">
                <Option value="">{t("internalControlFormDialog.fields.notSet")}</Option>
                {internalControlNatures.map(value => (
                  <Option key={value} value={value}>
                    {t(`internalControlFormDialog.natures.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.operatingFrequency")}
              error={formState.errors.operatingMode?.message}
            >
              <ControlledSelect control={control} name="operatingMode">
                <Option value="">{t("internalControlFormDialog.fields.notSet")}</Option>
                {internalControlOperatingModes.map(value => (
                  <Option key={value} value={value}>
                    {t(`internalControlFormDialog.operatingModes.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            {operatingMode === "PERIODIC" && (
              <PropertyRow
                label={t("internalControlFormDialog.operatingModes.periodic")}
                error={formState.errors.operatingInterval?.message}
              >
                <Suspense fallback={null}>
                  <Controller
                    control={control}
                    name="operatingInterval"
                    render={({ field }) => (
                      <TaskDurationField
                        value={field.value}
                        onValueChange={field.onChange}
                        defaultValue="P3M"
                        units={controlDurationUnits}
                        addLabel={t("internalControlFormDialog.actions.addDuration")}
                        clearLabel={t("internalControlFormDialog.actions.clearDuration")}
                      />
                    )}
                  />
                </Suspense>
              </PropertyRow>
            )}
            {operatingMode === "EVENT" && (
              <PropertyRow
                label={t("internalControlFormDialog.fields.operatingEvent")}
                error={formState.errors.operatingEvent?.message}
              >
                <Input
                  {...register("operatingEvent")}
                  placeholder={t("internalControlFormDialog.fields.operatingEventPlaceholder")}
                />
              </PropertyRow>
            )}
            <PropertyRow
              label={t("internalControlFormDialog.fields.evidenceCadence")}
              error={formState.errors.evidenceCadence?.message}
            >
              <Suspense fallback={null}>
                <Controller
                  control={control}
                  name="evidenceCadence"
                  render={({ field }) => (
                    <TaskDurationField
                      value={field.value}
                      onValueChange={field.onChange}
                      defaultValue="P1M"
                      units={controlDurationUnits}
                      addLabel={t("internalControlFormDialog.actions.addDuration")}
                      clearLabel={t("internalControlFormDialog.actions.clearDuration")}
                    />
                  )}
                />
              </Suspense>
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.testingCadence")}
              error={formState.errors.testingCadence?.message}
            >
              <Suspense fallback={null}>
                <Controller
                  control={control}
                  name="testingCadence"
                  render={({ field }) => (
                    <TaskDurationField
                      value={field.value}
                      onValueChange={field.onChange}
                      defaultValue="P3M"
                      units={controlDurationUnits}
                      addLabel={t("internalControlFormDialog.actions.addDuration")}
                      clearLabel={t("internalControlFormDialog.actions.clearDuration")}
                    />
                  )}
                />
              </Suspense>
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.implementationStatus")}
              error={formState.errors.implementationStatus?.message}
            >
              <ControlledSelect control={control} name="implementationStatus">
                {internalControlImplementationStatuses.map(value => (
                  <Option key={value} value={value}>
                    {t(`internalControlFormDialog.implementationStatuses.${value.toLowerCase()}`)}
                  </Option>
                ))}
              </ControlledSelect>
            </PropertyRow>
            <PropertyRow
              label={t("internalControlFormDialog.fields.owner")}
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
              label={t("internalControlFormDialog.fields.reviewer")}
              error={formState.errors.reviewerId?.message}
            >
              <PeopleSelectField
                name="reviewerId"
                control={control}
                organizationId={organizationId}
                optional
              />
            </PropertyRow>
            {internalControl && (
              <PropertyRow
                label={t("internalControlFormDialog.fields.state")}
                error={formState.errors.state?.message}
              >
                <ControlledSelect
                  control={control}
                  name="state"
                  placeholder={t("internalControlFormDialog.fields.statePlaceholder")}
                >
                  {internalControlStates.map(state => (
                    <Option key={state} value={state}>
                      {t(`internalControlFormDialog.states.${state.toLowerCase()}`)}
                    </Option>
                  ))}
                </ControlledSelect>
              </PropertyRow>
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">
            {internalControl ? t("internalControlFormDialog.actions.update") : t("internalControlFormDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function operatingFrequencyInput(data: {
  operatingMode: "" | (typeof internalControlOperatingModes)[number];
  operatingInterval: string | null;
  operatingEvent: string;
}) {
  if (data.operatingMode === "") {
    return null;
  }

  if (data.operatingMode === "PERIODIC") {
    return {
      mode: data.operatingMode,
      interval: data.operatingInterval,
    };
  }

  if (data.operatingMode === "EVENT") {
    const event = data.operatingEvent.trim();
    return {
      mode: data.operatingMode,
      ...(event ? { event } : {}),
    };
  }

  return { mode: data.operatingMode };
}
