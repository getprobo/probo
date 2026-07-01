// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError, getRiskImpacts, getRiskLikelihoods } from "@probo/helpers";
import { useToggle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconPlusLarge,
  Label,
  Option,
  PropertyRow,
  Select,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, useMemo } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { FormRiskDialog_risk$key } from "#/__generated__/core/FormRiskDialog_risk.graphql";
import type { FormRiskDialogMutation } from "#/__generated__/core/FormRiskDialogMutation.graphql";
import type { FormRiskDialogUpdateRiskMutation } from "#/__generated__/core/FormRiskDialogUpdateRiskMutation.graphql";
import {
  ControlledField,
  ControlledSelect,
} from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import {
  type RiskData,
  type RiskForm,
  useRiskForm,
} from "#/hooks/forms/useRiskForm";
import { useFetchQuery } from "#/hooks/useFetchQuery";
import { useOrganizationId } from "#/hooks/useOrganizationId";

interface FormRiskDialogProps {
  trigger?: ReactNode;
  risk?: FormRiskDialog_risk$key;
  connection?: string;
  ref?: ReturnType<typeof useDialogRef>;
  onSuccess?: () => void;
}

type RiskTemplate = {
  category: string;
  name: string;
  description: string;
};

const formRiskFragment = graphql`
  fragment FormRiskDialog_risk on Risk {
    id
    name
    category
    description
    treatment
    inherentLikelihood
    inherentImpact
    residualLikelihood
    residualImpact
    inherentRiskScore
    residualRiskScore
    note
    owner {
      id
    }
  }
`;

const createRiskMutation = graphql`
  mutation FormRiskDialogMutation(
    $input: CreateRiskInput!
    $connections: [ID!]!
  ) {
    createRisk(input: $input) {
      riskEdge @prependEdge(connections: $connections) {
        node {
          ...FormRiskDialog_risk
        }
      }
    }
  }
`;

const updateRiskMutation = graphql`
  mutation FormRiskDialogUpdateRiskMutation($input: UpdateRiskInput!) {
    updateRisk(input: $input) {
      risk {
        ...FormRiskDialog_risk
      }
    }
  }
`;

export function FormRiskDialog({
  trigger,
  risk: riskKey,
  connection,
  ref: refProps,
  onSuccess,
}: FormRiskDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const ref = refProps ?? dialogRef;

  const risk = useFragment(formRiskFragment, riskKey ?? null);
  const formDefaults = risk
    ? {
        id: risk.id,
        name: risk.name,
        category: risk.category,
        description: risk.description,
        treatment: risk.treatment,
        inherentLikelihood: risk.inherentLikelihood,
        inherentImpact: risk.inherentImpact,
        residualLikelihood: risk.residualLikelihood,
        residualImpact: risk.residualImpact,
        inherentRiskScore: risk.inherentRiskScore,
        residualRiskScore: risk.residualRiskScore,
        note: risk.note,
        owner: risk.owner,
      }
    : undefined;
  const { control, handleSubmit, setValue, register, watch, formState, reset }
    = useRiskForm(formDefaults);
  const errors = formState.errors ?? {};
  const [createRisk, isCreating]
    = useMutation<FormRiskDialogMutation>(createRiskMutation);
  const [updateRisk, isUpdating]
    = useMutation<FormRiskDialogUpdateRiskMutation>(updateRiskMutation);
  const isLoading = isCreating || isUpdating;

  const onTemplateChange = (template: RiskTemplate) => {
    setValue("name", template.name);
    setValue("description", template.description);
  };

  const onSubmit = (data: RiskData) => {
    if (risk) {
      updateRisk({
        variables: {
          input: {
            id: risk.id,
            ...data,
            description: data.description || null,
          },
        },
        onCompleted() {
          toast({
            title: __("Success"),
            description: __("Risk updated successfully."),
            variant: "success",
          });
          ref?.current?.close();
        },
        onError(error) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to update risk"),
              error,
            ),
            variant: "error",
          });
        },
      });
      return;
    }
    createRisk({
      variables: {
        input: {
          ...data,
          description: data.description || null,
          organizationId,
        },
        connections: [connection!],
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Risk created successfully."),
          variant: "success",
        });
        ref?.current?.close();
        reset();
        onSuccess?.();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create risk"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  const [showNote, toggleNote] = useToggle(false);

  return (
    <Dialog
      ref={ref}
      trigger={trigger}
      title={(
        <Breadcrumb
          items={[__("Risks"), risk ? __("Edit Risk") : __("New Risk")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
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

          <div className="py-5 px-6 bg-subtle">
            <Label>{__("Properties")}</Label>

            <PropertyRow
              id="ownerId"
              label={__("Owner")}
              error={errors.ownerId?.message}
            >
              <PeopleSelectField
                name="ownerId"
                control={control}
                organizationId={organizationId}
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
  errors: RiskForm["formState"]["errors"];
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
          {getRiskImpacts(__).map(i => (
            <Option key={i.value} value={i.value.toString()}>
              {i.value}
              {" "}
              -
              {i.label}
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
          {getRiskLikelihoods(__).map(l => (
            <Option key={l.value} value={l.value.toString()}>
              {l.value}
              {" "}
              -
              {l.label}
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
    },
  );

  const categories = useMemo(
    () => Array.from(new Set(risks?.map(t => t.category))),
    [risks],
  );

  const selectedCategory = watch("category");

  const templates = useMemo(
    () => risks?.filter(r => r.category === selectedCategory) ?? [],
    [risks, selectedCategory],
  );

  const onTemplateChange = (template: string) => {
    const risk = risks?.find(r => r.name === template);
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
          {categories.map(category => (
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
          {templates?.map(template => (
            <Option key={template.name} value={template.name}>
              {template.name}
            </Option>
          ))}
        </Select>
      </div>
    </div>
  );
}
