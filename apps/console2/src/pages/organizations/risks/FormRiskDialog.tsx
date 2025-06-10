import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Label,
  Select,
  Option,
  Field,
  Card,
  PropertyRow,
  IconPlusLarge,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { useMemo, useState, type ReactNode } from "react";
import { useFetchQuery } from "/hooks/useFetchQuery";
import { graphql } from "relay-runtime";
import { PeopleSelect } from "/components/form/PeopleSelect";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useToggle } from "@probo/hooks";
import {
  ControlledField,
  ControlledSelect,
} from "/components/form/ControlledField";
import {
  useRiskForm,
  type RiskForm,
  type RiskKey,
} from "/hooks/forms/useRiskForm";
import type { FieldErrors } from "react-hook-form";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { FormRiskDialogMutation } from "./__generated__/FormRiskDialogMutation.graphql";
import type { FormRiskDialogUpdateRiskMutation } from "./__generated__/FormRiskDialogUpdateRiskMutation.graphql";
import { getRiskImpacts, getRiskLikelihoods } from "@probo/helpers";

type Props = {
  trigger?: ReactNode;
  risk?: RiskKey;
  connection?: string;
  ref?: ReturnType<typeof useDialogRef>;
};

type RiskTemplate = {
  category: string;
  name: string;
  description: string;
};

const createRiskMutation = graphql`
  mutation FormRiskDialogMutation(
    $input: CreateRiskInput!
    $connections: [ID!]!
  ) {
    createRisk(input: $input) {
      riskEdge @prependEdge(connections: $connections) {
        node {
          ...useRiskFormFragment
        }
      }
    }
  }
`;

const updateRiskMutation = graphql`
  mutation FormRiskDialogUpdateRiskMutation($input: UpdateRiskInput!) {
    updateRisk(input: $input) {
      risk {
        ...useRiskFormFragment
      }
    }
  }
`;

/**
 * Dialog to create or update a risk
 */
export default function FormRiskDialog({
  trigger,
  risk,
  connection,
  ref: refProps,
}: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const ref = refProps ?? useDialogRef();

  const { control, handleSubmit, setValue, register, watch, formState, reset } =
    useRiskForm(risk);
  const errors = formState.errors ?? {};
  const [createRisk, isLoadingCreate] =
    useMutationWithToasts<FormRiskDialogMutation>(createRiskMutation);
  const [updateRisk, isLoadingUpdate] =
    useMutationWithToasts<FormRiskDialogUpdateRiskMutation>(updateRiskMutation);
  const isLoading = isLoadingCreate || isLoadingUpdate;

  const onTemplateChange = (risk: RiskTemplate) => {
    setValue("name", risk.name);
    setValue("description", risk.description);
  };

  const onSubmit = handleSubmit((data) => {
    if (risk) {
      updateRisk({
        variables: {
          input: {
            id: risk.id,
            ...data,
          },
        },
        successMessage: __("Risk updated successfully."),
        errorMessage: __("Failed to update risk. Please try again."),
        onSuccess: () => {
          ref?.current?.close();
        },
      });
      return;
    }
    createRisk({
      variables: {
        input: {
          ...data,
          organizationId,
        },
        connections: [connection!],
      },
      successMessage: __("Risk created successfully."),
      errorMessage: __("Failed to create risk. Please try again."),
      onSuccess: () => {
        ref?.current?.close();
        reset();
      },
    });
  });

  const [showNote, toggleNote] = useToggle(false);

  return (
    <Dialog
      ref={ref}
      trigger={trigger}
      title={
        <Breadcrumb
          items={[__("Risks"), risk ? __("Edit Risk") : __("New Risk")]}
        />
      }
    >
      <form onSubmit={onSubmit}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          {/* Main form */}
          <div className="py-8 px-12 space-y-6">
            <TemplateSelector
              onChange={onTemplateChange}
              control={control}
              watch={watch}
            />
            <Field
              type="text"
              {...register("name")}
              error={errors.name?.message}
              label={__("Risk name")}
              placeholder={__("Service Outage")}
            />
            <Field
              {...register("description")}
              error={errors.description?.message}
              label={__("Description")}
              placeholder={__("Type your description here")}
              type="textarea"
            />

            <div className="grid grid-cols-2 gap-6">
              <ImpactAndLikelihood
                errors={errors}
                control={control}
                label={__("Inherent Risk")}
                prefix="inherent"
              />
              <ImpactAndLikelihood
                errors={errors}
                control={control}
                label={__("Residual Risk")}
                prefix="residual"
              />
            </div>
          </div>

          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{__("Properties")}</Label>

            <PropertyRow
              id="ownerId"
              label={__("Owner")}
              error={errors.ownerId?.message}
            >
              <PeopleSelect
                name="ownerId"
                control={control}
                organization={organizationId}
              />
            </PropertyRow>

            <PropertyRow
              id="treatment"
              label={__("Treatment strategy")}
              error={errors.treatment?.message}
            >
              <ControlledSelect
                control={control}
                name="treatment"
                variant="editor"
                placeholder={__("Select a treatment strategy")}
              >
                <Option value="AVOIDED">Avoid</Option>
                <Option value="MITIGATED">Mitigate</Option>
                <Option value="TRANSFERRED">Transfer</Option>
                <Option value="ACCEPTED">Accept</Option>
              </ControlledSelect>
            </PropertyRow>

            <PropertyRow
              id="note"
              label={__("Note")}
              error={errors.note?.message}
            >
              <Button
                type="button"
                variant="quaternary"
                icon={IconPlusLarge}
                onClick={toggleNote}
              />
              {showNote && (
                <Textarea
                  {...register("note")}
                  className="animate-in slide-in-from-top-2"
                  placeholder={__("Add any additional notes about this risk")}
                />
              )}
            </PropertyRow>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {risk ? __("Update risk") : __("Create risk")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function ImpactAndLikelihood({
  label,
  prefix,
  control,
  errors,
}: {
  label: string;
  prefix: "inherent" | "residual";
  control: RiskForm["control"];
  errors: FieldErrors<{
    inherentImpact: string;
    inherentLikelihood: string;
    residualImpact: string;
    residualLikelihood: string;
  }>;
}) {
  const { __ } = useTranslate();
  return (
    <div>
      <Label>{label}</Label>
      <Card padded className="space-y-4 p-4">
        <ControlledField
          control={control}
          name={`${prefix}Impact`}
          type="select"
          label={__("Impact")}
          placeholder={__("Select impact level")}
          error={errors?.[`${prefix}Impact`]?.message}
        >
          {getRiskImpacts(__).map((i) => (
            <Option key={i.value} value={i.value.toString()}>
              {i.value} - {i.label}
            </Option>
          ))}
        </ControlledField>
        <ControlledField
          control={control}
          name={`${prefix}Likelihood`}
          type="select"
          label={__("Likelihood")}
          placeholder={__("Select likelihood level")}
          error={errors?.[`${prefix}Likelihood`]?.message}
        >
          {getRiskLikelihoods(__).map((l) => (
            <Option key={l.value} value={l.value.toString()}>
              {l.value} - {l.label}
            </Option>
          ))}
        </ControlledField>
      </Card>
    </div>
  );
}

function TemplateSelector({
  onChange,
  control,
  watch,
}: {
  onChange: (risk: RiskTemplate) => void;
  control: RiskForm["control"];
  watch: RiskForm["watch"];
}) {
  const { __ } = useTranslate();
  const { data: risks } = useFetchQuery<RiskTemplate[]>(
    "/data/risks/risks.json",
    {
      staleTime: 100_000,
    }
  );

  const categories = useMemo(
    () => Array.from(new Set(risks?.map((t) => t.category))),
    [risks]
  );

  const selectedCategory = watch("category");

  const templates = useMemo(
    () => risks?.filter((r) => r.category === selectedCategory) ?? [],
    [risks, selectedCategory]
  );

  const onTemplateChange = (template: string) => {
    const risk = risks?.find((r) => r.name === template);
    if (!risk) {
      throw new Error("Risk not found");
    }
    onChange(risk);
  };
  return (
    <div>
      <Label>{__("Risk category")}</Label>
      <div className="grid grid-cols-2 gap-2">
        <ControlledSelect
          control={control}
          name="category"
          placeholder={__("Select a category")}
        >
          {categories.map((category) => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </ControlledSelect>
        <Select
          key={selectedCategory}
          variant={templates?.length === 0 ? "dashed" : "default"}
          placeholder={__("Select template")}
          onValueChange={onTemplateChange}
        >
          {templates?.map((template) => (
            <Option key={template.name} value={template.name}>
              {template.name}
            </Option>
          ))}
        </Select>
      </div>
    </div>
  );
}
