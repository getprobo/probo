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
  useToast,
} from "@probo/ui";
import { useMemo, useState, type ReactNode } from "react";
import { useFetchQuery } from "../../../hooks/useFetchQuery";
import { graphql } from "relay-runtime";
import { UserSelect } from "../../../components/form/UserSelect";
import { useOrganizationId } from "../../../hooks/useOrganizationId";
import { useToggle } from "@probo/hooks";
import {
  ControlledField,
  ControlledSelect,
} from "../../../components/form/ControlledField";
import { useRiskForm, type RiskForm } from "./forms/useRiskForm";
import { useMutation } from "react-relay";

type Props = {
  trigger: ReactNode;
};

type Risk = {
  category: string;
  name: string;
  description: string;
};

const createRiskMutation = graphql`
  mutation NewRiskDialogMutation(
    $input: CreateRiskInput!
    $connections: [ID!]!
  ) {
    createRisk(input: $input) {
      riskEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          category
          inherentLikelihood
          inherentImpact
          residualLikelihood
          residualImpact
          treatment
          createdAt
          updatedAt
        }
      }
    }
  }
`;

export default function NewRiskDialog({ trigger }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const { control, handleSubmit, setValue, register, watch, formState } =
    useRiskForm(organizationId);
  const errors = formState.errors ?? {};
  const [createRisk, isLoading] = useMutation(createRiskMutation);
  const { toast } = useToast();

  const onTemplateChange = (risk: Risk) => {
    setValue("name", risk.name);
    setValue("description", risk.description);
  };

  const onSubmit = handleSubmit((data) => {
    createRisk({
      variables: {
        input: data,
      },
      onCompleted: (response, error) => {
        if (error) {
          toast({
            title: __("Error"),
            description: __("Failed to create risk. Please try again."),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Risk created successfully."),
          variant: "success",
        });

        setOpen(false);
      },
    });
  });

  const [showNote, toggleNote] = useToggle(false);
  const [open, setOpen] = useState(false);

  return (
    <Dialog
      open={open}
      onOpenChange={setOpen}
      trigger={trigger}
      title={<Breadcrumb items={[__("Risks"), __("New Risk")]}></Breadcrumb>}
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
              <UserSelect
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
                  className="animate-in slide-in-from-top-2"
                  placeholder={__("Add any additional notes about this risk")}
                />
              )}
            </PropertyRow>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {__("Create risk")}
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
  prefix: string;
  control: RiskForm["control"];
  errors: Record<string, { message: string }>;
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
          <Option value="1">1 - Negligible</Option>
          <Option value="2">2 - Low</Option>
          <Option value="3">3 - Moderate</Option>
          <Option value="4">4 - Significant</Option>
          <Option value="5">5 - Catastrophic</Option>
        </ControlledField>
        <ControlledField
          control={control}
          name={`${prefix}Likelihood`}
          type="select"
          label={__("Likelihood")}
          placeholder={__("Select likelihood level")}
          error={errors?.[`${prefix}Likelihood`]?.message}
        >
          <Option value="1">1 - Improbable</Option>
          <Option value="2">2 - Remote</Option>
          <Option value="3">3 - Occasional</Option>
          <Option value="4">4 - Probable</Option>
          <Option value="5">5 - Frequent</Option>
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
  onChange: (risk: Risk) => void;
  control: RiskForm["control"];
  watch: RiskForm["watch"];
}) {
  const { __ } = useTranslate();
  const { data: risks } = useFetchQuery<Risk[]>("/data/risks/risks.json", {
    staleTime: 100_000,
  });

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
