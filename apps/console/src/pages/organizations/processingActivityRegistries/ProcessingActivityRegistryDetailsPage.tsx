import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import {
  processingActivityRegistryNodeQuery,
  useDeleteProcessingActivityRegistry,
  useUpdateProcessingActivityRegistry,
  ProcessingActivityRegistriesConnectionKey,
} from "../../../hooks/graph/ProcessingActivityRegistryGraph";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  DropdownItem,
  Field,
  Card,
  Textarea,
  useToast,
  Label,
  Checkbox,
  Select,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { AuditSelectField } from "/components/form/AuditSelectField";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Controller } from "react-hook-form";
import z from "zod";
import {
  SpecialOrCriminalDataOptions,
  LawfulBasisOptions,
  TransferSafeguardsOptions,
  DataProtectionImpactAssessmentOptions,
  TransferImpactAssessmentOptions,
} from "../../../components/form/ProcessingActivityRegistryEnumOptions";

import type { ProcessingActivityRegistryGraphNodeQuery } from "/hooks/graph/__generated__/ProcessingActivityRegistryGraphNodeQuery.graphql";

const updateRegistrySchema = z.object({
  name: z.string().min(1, "Name is required"),
  purpose: z.string().optional(),
  dataSubjectCategory: z.string().optional(),
  personalDataCategory: z.string().optional(),
  specialOrCriminalData: z.enum(["YES", "NO", "POSSIBLE"] as const),
  consentEvidenceLink: z.string().optional(),
  lawfulBasis: z.enum(["CONSENT", "CONTRACTUAL_NECESSITY", "LEGAL_OBLIGATION", "LEGITIMATE_INTEREST", "PUBLIC_TASK", "VITAL_INTERESTS"] as const),
  recipients: z.string().optional(),
  location: z.string().optional(),
  internationalTransfers: z.boolean(),
  transferSafeguards: z.string(),
  retentionPeriod: z.string().optional(),
  securityMeasures: z.string().optional(),
  dataProtectionImpactAssessment: z.enum(["NEEDED", "NOT_NEEDED"] as const),
  transferImpactAssessment: z.enum(["NEEDED", "NOT_NEEDED"] as const),
  auditId: z.string().min(1, "Audit is required"),
});

type Props = {
  queryRef: PreloadedQuery<ProcessingActivityRegistryGraphNodeQuery>;
};

export default function ProcessingActivityRegistryDetailsPage(props: Props) {
  const data = usePreloadedQuery<ProcessingActivityRegistryGraphNodeQuery>(processingActivityRegistryNodeQuery, props.queryRef);
  const registry = data.node;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  if (!registry) {
    return <div>{__("Processing activity registry entry not found")}</div>;
  }

  const updateRegistry = useUpdateProcessingActivityRegistry();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivityRegistriesConnectionKey
  );

  const deleteRegistry = useDeleteProcessingActivityRegistry({ id: registry.id!, name: registry.name! }, connectionId);

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateRegistrySchema,
    {
      defaultValues: {
        name: registry.name || "",
        purpose: registry.purpose || "",
        dataSubjectCategory: registry.dataSubjectCategory || "",
        personalDataCategory: registry.personalDataCategory || "",
        specialOrCriminalData: registry.specialOrCriminalData || "NO" as const,
        consentEvidenceLink: registry.consentEvidenceLink || "",
        lawfulBasis: registry.lawfulBasis || "LEGITIMATE_INTEREST" as const,
        recipients: registry.recipients || "",
        location: registry.location || "",
        internationalTransfers: registry.internationalTransfers || false,
        transferSafeguards: registry.transferSafeguards || "__NONE__",
        retentionPeriod: registry.retentionPeriod || "",
        securityMeasures: registry.securityMeasures || "",
        dataProtectionImpactAssessment: registry.dataProtectionImpactAssessment || "NOT_NEEDED" as const,
        transferImpactAssessment: registry.transferImpactAssessment || "NOT_NEEDED" as const,
        auditId: registry.audit?.id || "",
      },
    }
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateRegistry({
        id: registry.id!,
        auditId: formData.auditId || undefined,
        name: formData.name,
        purpose: formData.purpose || undefined,
        dataSubjectCategory: formData.dataSubjectCategory || undefined,
        personalDataCategory: formData.personalDataCategory || undefined,
        specialOrCriminalData: formData.specialOrCriminalData || undefined,
        consentEvidenceLink: formData.consentEvidenceLink || undefined,
        lawfulBasis: formData.lawfulBasis || undefined,
        recipients: formData.recipients || undefined,
        location: formData.location || undefined,
        internationalTransfers: formData.internationalTransfers,
        transferSafeguards: formData.transferSafeguards === "__NONE__" ? undefined : formData.transferSafeguards || undefined,
        retentionPeriod: formData.retentionPeriod || undefined,
        securityMeasures: formData.securityMeasures || undefined,
        dataProtectionImpactAssessment: formData.dataProtectionImpactAssessment || undefined,
        transferImpactAssessment: formData.transferImpactAssessment || undefined,
      });

      toast({
        title: __("Success"),
        description: __("Processing activity registry entry updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to update processing activity registry entry"),
        variant: "error",
      });
    }
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Breadcrumb
          items={[
            { label: __("Processing Activity Registries"), to: "../processing-activity-registries" },
            { label: registry.name! },
          ]}
        />
        <ActionDropdown>
          <DropdownItem onClick={deleteRegistry} variant="danger">
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </div>

      <Card>
        <div className="p-6">
          <div className="mb-6">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">{registry.name}</h1>
            </div>
          </div>

          <form onSubmit={onSubmit} className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div className="space-y-4">
                <Field
                  label={__("Name")}
                  {...register("name")}
                  error={formState.errors.name?.message}
                  required
                />

                <AuditSelectField
                  organizationId={organizationId}
                  control={control}
                  name="auditId"
                  label={__("Audit")}
                  error={formState.errors.auditId?.message}
                  required
                />

                <div>
                  <Label>{__("Purpose")}</Label>
                  <Textarea
                    {...register("purpose")}
                    placeholder={__("Describe the purpose of processing")}
                    rows={3}
                  />
                </div>

                <Field
                  label={__("Data Subject Category")}
                  {...register("dataSubjectCategory")}
                  placeholder={__("e.g., employees, customers, prospects")}
                />

                <Field
                  label={__("Personal Data Category")}
                  {...register("personalDataCategory")}
                  placeholder={__("e.g., contact details, financial data")}
                />

                <div>
                  <Label htmlFor="specialOrCriminalData">{__("Special or Criminal Data")} *</Label>
                  <Controller
                    control={control}
                    name="specialOrCriminalData"
                    render={({ field }) => (
                      <Select
                        id="specialOrCriminalData"
                        placeholder={__("Select special or criminal data status")}
                        onValueChange={field.onChange}
                        value={field.value}
                        className="w-full"
                      >
                        <SpecialOrCriminalDataOptions />
                      </Select>
                    )}
                  />
                  {formState.errors.specialOrCriminalData && (
                    <p className="text-sm text-txt-danger mt-1">{formState.errors.specialOrCriminalData.message}</p>
                  )}
                </div>

                <Field
                  label={__("Consent Evidence Link")}
                  {...register("consentEvidenceLink")}
                  placeholder={__("Link to consent evidence if applicable")}
                />

                <div>
                  <Label htmlFor="lawfulBasis">{__("Lawful Basis")} *</Label>
                  <Controller
                    control={control}
                    name="lawfulBasis"
                    render={({ field }) => (
                      <Select
                        id="lawfulBasis"
                        placeholder={__("Select lawful basis for processing")}
                        onValueChange={field.onChange}
                        value={field.value}
                        className="w-full"
                      >
                        <LawfulBasisOptions />
                      </Select>
                    )}
                  />
                  {formState.errors.lawfulBasis && (
                    <p className="text-sm text-txt-danger mt-1">{formState.errors.lawfulBasis.message}</p>
                  )}
                </div>
              </div>

              <div className="space-y-4">
                <Field
                  label={__("Recipients")}
                  {...register("recipients")}
                  placeholder={__("Who receives the data")}
                />

                <Field
                  label={__("Location")}
                  {...register("location")}
                  placeholder={__("Where is the data processed")}
                />

                <Controller
                  control={control}
                  name="internationalTransfers"
                  render={({ field }) => (
                                         <div>
                       <Label>{__("International Transfers")}</Label>
                       <div className="mt-2 flex items-center gap-2">
                         <Checkbox
                           checked={field.value}
                           onChange={field.onChange}
                         />
                         <span>{__("Data is transferred internationally")}</span>
                       </div>
                     </div>
                  )}
                />

                <div>
                  <Label htmlFor="transferSafeguards">{__("Transfer Safeguards")}</Label>
                  <Controller
                    control={control}
                    name="transferSafeguards"
                    render={({ field }) => (
                      <Select
                        id="transferSafeguards"
                        placeholder={__("Select transfer safeguards")}
                        onValueChange={field.onChange}
                        value={field.value}
                        className="w-full"
                      >
                        <TransferSafeguardsOptions />
                      </Select>
                    )}
                  />
                  {formState.errors.transferSafeguards && (
                    <p className="text-sm text-txt-danger mt-1">{formState.errors.transferSafeguards.message}</p>
                  )}
                </div>

                <Field
                  label={__("Retention Period")}
                  {...register("retentionPeriod")}
                  placeholder={__("How long is data retained")}
                />

                <div>
                  <Label>{__("Security Measures")}</Label>
                  <Textarea
                    {...register("securityMeasures")}
                    placeholder={__("Technical and organizational measures")}
                    rows={3}
                  />
                </div>

                <div>
                  <Label htmlFor="dataProtectionImpactAssessment">{__("Data Protection Impact Assessment")} *</Label>
                  <Controller
                    control={control}
                    name="dataProtectionImpactAssessment"
                    render={({ field }) => (
                      <Select
                        id="dataProtectionImpactAssessment"
                        placeholder={__("Is DPIA needed?")}
                        onValueChange={field.onChange}
                        value={field.value}
                        className="w-full"
                      >
                        <DataProtectionImpactAssessmentOptions />
                      </Select>
                    )}
                  />
                  {formState.errors.dataProtectionImpactAssessment && (
                    <p className="text-sm text-txt-danger mt-1">{formState.errors.dataProtectionImpactAssessment.message}</p>
                  )}
                </div>

                <div>
                  <Label htmlFor="transferImpactAssessment">{__("Transfer Impact Assessment")} *</Label>
                  <Controller
                    control={control}
                    name="transferImpactAssessment"
                    render={({ field }) => (
                      <Select
                        id="transferImpactAssessment"
                        placeholder={__("Is TIA needed?")}
                        onValueChange={field.onChange}
                        value={field.value}
                        className="w-full"
                      >
                        <TransferImpactAssessmentOptions />
                      </Select>
                    )}
                  />
                  {formState.errors.transferImpactAssessment && (
                    <p className="text-sm text-txt-danger mt-1">{formState.errors.transferImpactAssessment.message}</p>
                  )}
                </div>
              </div>
            </div>

            <div className="flex justify-end pt-4">
              <Button
                type="submit"
                variant="primary"
                disabled={formState.isSubmitting}
              >
                {formState.isSubmitting ? __("Saving...") : __("Save Changes")}
              </Button>
            </div>
          </form>
        </div>
      </Card>
    </div>
  );
}
