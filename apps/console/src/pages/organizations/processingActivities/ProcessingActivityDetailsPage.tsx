import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import {
  processingActivityNodeQuery,
  useDeleteProcessingActivity,
  useUpdateProcessingActivity,
  ProcessingActivitiesConnectionKey,
} from "../../../hooks/graph/ProcessingActivityGraph";
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
import { useParams } from "react-router";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Controller } from "react-hook-form";
import { formatError, type GraphQLError } from "@probo/helpers";
import z from "zod";
import { validateSnapshotConsistency } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import { VendorsMultiSelectField } from "/components/form/VendorsMultiSelectField";
import {
  SpecialOrCriminalDataOptions,
  LawfulBasisOptions,
  TransferSafeguardsOptions,
  DataProtectionImpactAssessmentOptions,
  TransferImpactAssessmentOptions,
} from "../../../components/form/ProcessingActivityEnumOptions";

import type { ProcessingActivityGraphNodeQuery } from "/hooks/graph/__generated__/ProcessingActivityGraphNodeQuery.graphql";
import { use } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";

const updateProcessingActivitySchema = z.object({
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
  vendorIds: z.array(z.string()).optional(),
});

type Props = {
  queryRef: PreloadedQuery<ProcessingActivityGraphNodeQuery>;
};

export default function ProcessingActivityDetailsPage(props: Props) {
  const data = usePreloadedQuery<ProcessingActivityGraphNodeQuery>(processingActivityNodeQuery, props.queryRef);
  const activity = data.node;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const { isAuthorized } = use(PermissionsContext);

  if (!activity) {
    return <div>{__("Processing activity not found")}</div>;
  }

  validateSnapshotConsistency(activity, snapshotId);

  const updateActivity = useUpdateProcessingActivity();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivitiesConnectionKey,
    { filter: { snapshotId: snapshotId || null } }
  );

  const deleteActivity = useDeleteProcessingActivity({ id: activity.id!, name: activity.name! }, connectionId);

  const vendors = activity?.vendors?.edges.map((edge) => edge.node) ?? [];
  const vendorIds = vendors.map((vendor) => vendor.id);

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateProcessingActivitySchema,
    {
      defaultValues: {
        name: activity.name || "",
        purpose: activity.purpose || "",
        dataSubjectCategory: activity.dataSubjectCategory || "",
        personalDataCategory: activity.personalDataCategory || "",
        specialOrCriminalData: activity.specialOrCriminalData || "NO" as const,
        consentEvidenceLink: activity.consentEvidenceLink || "",
        lawfulBasis: activity.lawfulBasis || "LEGITIMATE_INTEREST" as const,
        recipients: activity.recipients || "",
        location: activity.location || "",
        internationalTransfers: activity.internationalTransfers || false,
        transferSafeguards: activity.transferSafeguards || "__NONE__",
        retentionPeriod: activity.retentionPeriod || "",
        securityMeasures: activity.securityMeasures || "",
        dataProtectionImpactAssessment: activity.dataProtectionImpactAssessment || "NOT_NEEDED" as const,
        transferImpactAssessment: activity.transferImpactAssessment || "NOT_NEEDED" as const,
        vendorIds: vendorIds,
      },
    }
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateActivity({
        id: activity.id!,
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
        vendorIds: formData.vendorIds,
      });

      toast({
        title: __("Success"),
        description: __("Processing activity updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to update processing activity"), error as GraphQLError),
        variant: "error",
      });
    }
  });

  const breadcrumbProcessingActivitiesUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/processing-activities`
    : `/organizations/${organizationId}/processing-activities`;

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <div className="flex items-center justify-between">
        <Breadcrumb
          items={[
            { label: __("Processing Activities"), to: breadcrumbProcessingActivitiesUrl },
            { label: activity.name! },
          ]}
        />
        {!isSnapshotMode && (
          isAuthorized("ProcessingActivity", "deleteProcessingActivity") && (
            <ActionDropdown>
              <DropdownItem onClick={deleteActivity} variant="danger">
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          )
        )}
      </div>

      <Card>
        <div className="p-6">
          <div className="mb-6">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">{activity.name}</h1>
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
                  disabled={isSnapshotMode}
                />

                <div>
                  <Label>{__("Purpose")}</Label>
                  <Textarea
                    {...register("purpose")}
                    placeholder={__("Describe the purpose of processing")}
                    rows={3}
                    disabled={isSnapshotMode}
                  />
                </div>

                <Field
                  label={__("Data Subject Category")}
                  {...register("dataSubjectCategory")}
                  placeholder={__("e.g., employees, customers, prospects")}
                  disabled={isSnapshotMode}
                />

                <Field
                  label={__("Personal Data Category")}
                  {...register("personalDataCategory")}
                  placeholder={__("e.g., contact details, financial data")}
                  disabled={isSnapshotMode}
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
                        disabled={isSnapshotMode}
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
                  disabled={isSnapshotMode}
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
                        disabled={isSnapshotMode}
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
                  disabled={isSnapshotMode}
                />

                <Field
                  label={__("Location")}
                  {...register("location")}
                  placeholder={__("Where is the data processed")}
                  disabled={isSnapshotMode}
                />

                <Controller
                  control={control}
                  name="internationalTransfers"
                  render={({ field }) => (
                    <div>
                      <Label>{__("International Transfers")}</Label>
                      <div className="mt-2 flex items-center gap-2">
                        <Checkbox
                          checked={field.value ?? false}
                          onChange={field.onChange}
                          disabled={isSnapshotMode}
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
                        disabled={isSnapshotMode}
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
                  disabled={isSnapshotMode}
                />

                <div>
                  <Label>{__("Security Measures")}</Label>
                  <Textarea
                    {...register("securityMeasures")}
                    placeholder={__("Technical and organizational measures")}
                    rows={3}
                    disabled={isSnapshotMode}
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
                        disabled={isSnapshotMode}
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
                        disabled={isSnapshotMode}
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

            <VendorsMultiSelectField
              organizationId={organizationId}
              control={control}
              name="vendorIds"
              selectedVendors={vendors}
              label={__("Vendors")}
              disabled={isSnapshotMode}
            />

            {!isSnapshotMode && (
              <div className="flex justify-end pt-4">
                <Authorized entity="ProcessingActivity" action="updateProcessingActivity">
                  <Button
                    type="submit"
                    variant="primary"
                    disabled={formState.isSubmitting}
                  >
                    {formState.isSubmitting ? __("Saving...") : __("Save Changes")}
                  </Button>
                </Authorized>
              </div>
            )}
          </form>
        </div>
      </Card>
    </div>
  );
}
