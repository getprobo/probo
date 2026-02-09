import { formatError, type GraphQLError } from "@probo/helpers";
import {
  formatDatetime,
  toDateInput,
  validateSnapshotConsistency,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  Card,
  Checkbox,
  DropdownItem,
  Field,
  Input,
  Label,
  Option,
  Select,
  TabItem,
  Tabs,
  Textarea,
  useToast,
} from "@probo/ui";
import { useEffect, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import { z } from "zod";

import type { ProcessingActivityGraphNodeQuery } from "#/__generated__/core/ProcessingActivityGraphNodeQuery.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { VendorsMultiSelectField } from "#/components/form/VendorsMultiSelectField";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  DataProtectionImpactAssessmentOptions,
  LawfulBasisOptions,
  RoleOptions,
  SpecialOrCriminalDataOptions,
  TransferImpactAssessmentOptions,
  TransferSafeguardsOptions,
} from "../../../components/form/ProcessingActivityEnumOptions";
import {
  ProcessingActivitiesConnectionKey,
  type ProcessingActivityDPIAResidualRisk,
  processingActivityNodeQuery,
  useCreateDataProtectionImpactAssessment,
  useCreateTransferImpactAssessment,
  useDeleteDataProtectionImpactAssessment,
  useDeleteProcessingActivity,
  useDeleteTransferImpactAssessment,
  useUpdateDataProtectionImpactAssessment,
  useUpdateProcessingActivity,
  useUpdateTransferImpactAssessment,
} from "../../../hooks/graph/ProcessingActivityGraph";

const updateProcessingActivitySchema = z.object({
  name: z.string().min(1, "Name is required"),
  purpose: z.string().optional(),
  dataSubjectCategory: z.string().optional(),
  personalDataCategory: z.string().optional(),
  specialOrCriminalData: z.enum(["YES", "NO", "POSSIBLE"] as const),
  consentEvidenceLink: z.string().optional(),
  lawfulBasis: z.enum([
    "CONSENT",
    "CONTRACTUAL_NECESSITY",
    "LEGAL_OBLIGATION",
    "LEGITIMATE_INTEREST",
    "PUBLIC_TASK",
    "VITAL_INTERESTS",
  ] as const),
  recipients: z.string().optional(),
  location: z.string().optional(),
  internationalTransfers: z.boolean(),
  transferSafeguards: z.string(),
  retentionPeriod: z.string().optional(),
  securityMeasures: z.string().optional(),
  dataProtectionImpactAssessmentNeeded: z.enum([
    "NEEDED",
    "NOT_NEEDED",
  ] as const),
  transferImpactAssessmentNeeded: z.enum(["NEEDED", "NOT_NEEDED"] as const),
  lastReviewDate: z.string().optional(),
  nextReviewDate: z.string().optional(),
  role: z.enum(["CONTROLLER", "PROCESSOR"] as const),
  dataProtectionOfficerId: z.string().optional(),
  vendorIds: z.array(z.string()).optional(),
});

type Props = {
  queryRef: PreloadedQuery<ProcessingActivityGraphNodeQuery>;
};

export default function ProcessingActivityDetailsPage(props: Props) {
  const { node: activity }
    = usePreloadedQuery<ProcessingActivityGraphNodeQuery>(
      processingActivityNodeQuery,
      props.queryRef,
    );
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  // Get initial tab from URL hash
  const getInitialTab = (): "overview" | "dpia" | "tia" => {
    const hash = window.location.hash.slice(1);
    if (hash === "dpia" || hash === "tia") return hash;
    return "overview";
  };

  const [activeTab, setActiveTab] = useState<"overview" | "dpia" | "tia">(
    getInitialTab,
  );
  const [dpiaSubmitting, setDpiaSubmitting] = useState(false);
  const [tiaSubmitting, setTiaSubmitting] = useState(false);
  const [showDpiaForm, setShowDpiaForm] = useState(
    Boolean(activity?.dataProtectionImpactAssessment?.id),
  );
  const [showTiaForm, setShowTiaForm] = useState(
    Boolean(activity?.transferImpactAssessment?.id),
  );
  const [dpiaDeleted, setDpiaDeleted] = useState(false);
  const [tiaDeleted, setTiaDeleted] = useState(false);

  const canCreateOrUpdateDPIA = activity.dataProtectionImpactAssessment
    ? activity.dataProtectionImpactAssessment.canUpdate
    : activity.canCreateDPIA;
  const canCreateOrUpdateTIA = activity.transferImpactAssessment
    ? activity.transferImpactAssessment.canUpdate
    : activity.canCreateTIA;

  useEffect(() => {
    window.location.hash = activeTab === "overview" ? "" : activeTab;
  }, [activeTab]);

  validateSnapshotConsistency(activity, snapshotId);

  const updateActivity = useUpdateProcessingActivity();
  const createDPIA = useCreateDataProtectionImpactAssessment();
  const updateDPIA = useUpdateDataProtectionImpactAssessment();
  const deleteDPIA = useDeleteDataProtectionImpactAssessment(
    { id: activity?.dataProtectionImpactAssessment?.id || "" },
    {
      onSuccess: () => {
        setDpiaDeleted(true);
        setShowDpiaForm(false);
        dpiaForm.reset({
          description: "",
          necessityAndProportionality: "",
          potentialRisk: "",
          mitigations: "",
          residualRisk: "",
        });
      },
    },
  );
  const createTIA = useCreateTransferImpactAssessment();
  const updateTIA = useUpdateTransferImpactAssessment();
  const deleteTIA = useDeleteTransferImpactAssessment(
    { id: activity?.transferImpactAssessment?.id || "" },
    {
      onSuccess: () => {
        setTiaDeleted(true);
        setShowTiaForm(false);
        tiaForm.reset({
          dataSubjects: "",
          legalMechanism: "",
          transfer: "",
          localLawRisk: "",
          supplementaryMeasures: "",
        });
      },
    },
  );

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ProcessingActivitiesConnectionKey,
    { filter: { snapshotId: snapshotId || null } },
  );

  const deleteActivity = useDeleteProcessingActivity(
    { id: activity.id!, name: activity.name! },
    connectionId,
  );

  const vendors = activity?.vendors?.edges.map(edge => edge.node) ?? [];
  const vendorIds = vendors.map(vendor => vendor.id);

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateProcessingActivitySchema,
    {
      defaultValues: {
        name: activity.name || "",
        purpose: activity.purpose || "",
        dataSubjectCategory: activity.dataSubjectCategory || "",
        personalDataCategory: activity.personalDataCategory || "",
        specialOrCriminalData:
          activity.specialOrCriminalData || ("NO" as const),
        consentEvidenceLink: activity.consentEvidenceLink || "",
        lawfulBasis: activity.lawfulBasis || ("LEGITIMATE_INTEREST" as const),
        recipients: activity.recipients || "",
        location: activity.location || "",
        internationalTransfers: activity.internationalTransfers || false,
        transferSafeguards: activity.transferSafeguards || "__NONE__",
        retentionPeriod: activity.retentionPeriod || "",
        securityMeasures: activity.securityMeasures || "",
        dataProtectionImpactAssessmentNeeded:
          activity.dataProtectionImpactAssessmentNeeded
          || ("NOT_NEEDED" as const),
        transferImpactAssessmentNeeded:
          activity.transferImpactAssessmentNeeded || ("NOT_NEEDED" as const),
        lastReviewDate: toDateInput(activity.lastReviewDate),
        nextReviewDate: toDateInput(activity.nextReviewDate),
        role: activity.role || ("CONTROLLER" as const),
        dataProtectionOfficerId: activity.dataProtectionOfficer?.id || "",
        vendorIds: vendorIds,
      },
    },
  );

  const dpiaForm = useForm({
    defaultValues: {
      description: activity?.dataProtectionImpactAssessment?.description || "",
      necessityAndProportionality:
        activity?.dataProtectionImpactAssessment?.necessityAndProportionality
        || "",
      potentialRisk:
        activity?.dataProtectionImpactAssessment?.potentialRisk || "",
      mitigations: activity?.dataProtectionImpactAssessment?.mitigations || "",
      residualRisk: (activity?.dataProtectionImpactAssessment?.residualRisk
        || "") as ProcessingActivityDPIAResidualRisk | "",
    },
  });

  const tiaForm = useForm({
    defaultValues: {
      dataSubjects: activity?.transferImpactAssessment?.dataSubjects || "",
      legalMechanism: activity?.transferImpactAssessment?.legalMechanism || "",
      transfer: activity?.transferImpactAssessment?.transfer || "",
      localLawRisk: activity?.transferImpactAssessment?.localLawRisk || "",
      supplementaryMeasures:
        activity?.transferImpactAssessment?.supplementaryMeasures || "",
    },
  });

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
        transferSafeguards:
          formData.transferSafeguards === "__NONE__"
            ? undefined
            : formData.transferSafeguards || undefined,
        retentionPeriod: formData.retentionPeriod || undefined,
        securityMeasures: formData.securityMeasures || undefined,
        dataProtectionImpactAssessmentNeeded:
          formData.dataProtectionImpactAssessmentNeeded || undefined,
        transferImpactAssessmentNeeded:
          formData.transferImpactAssessmentNeeded || undefined,
        lastReviewDate: formatDatetime(formData.lastReviewDate) ?? null,
        nextReviewDate: formatDatetime(formData.nextReviewDate) ?? null,
        role: formData.role,
        dataProtectionOfficerId: formData.dataProtectionOfficerId || null,
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
        description: formatError(
          __("Failed to update processing activity"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const onDPIASubmit = dpiaForm.handleSubmit(async (formData) => {
    setDpiaSubmitting(true);
    try {
      const isCreating
        = !activity?.dataProtectionImpactAssessment?.id || dpiaDeleted;
      if (!isCreating) {
        // Update existing DPIA
        await updateDPIA({
          id: activity.dataProtectionImpactAssessment.id,
          description: formData.description || undefined,
          necessityAndProportionality:
            formData.necessityAndProportionality || undefined,
          potentialRisk: formData.potentialRisk || undefined,
          mitigations: formData.mitigations || undefined,
          residualRisk:
            (formData.residualRisk as ProcessingActivityDPIAResidualRisk)
            || undefined,
        });
        toast({
          title: __("Success"),
          description: __("DPIA updated successfully"),
          variant: "success",
        });
      } else {
        // Create new DPIA
        await createDPIA({
          processingActivityId: activity.id!,
          description: formData.description || undefined,
          necessityAndProportionality:
            formData.necessityAndProportionality || undefined,
          potentialRisk: formData.potentialRisk || undefined,
          mitigations: formData.mitigations || undefined,
          residualRisk:
            (formData.residualRisk as ProcessingActivityDPIAResidualRisk)
            || undefined,
        });
        setDpiaDeleted(false);
        setShowDpiaForm(true);
        toast({
          title: __("Success"),
          description: __("DPIA created successfully"),
          variant: "success",
        });
      }
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to save DPIA"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    } finally {
      setDpiaSubmitting(false);
    }
  });

  const onTIASubmit = tiaForm.handleSubmit(async (formData) => {
    setTiaSubmitting(true);
    try {
      const isCreating = !activity?.transferImpactAssessment?.id || tiaDeleted;
      if (!isCreating) {
        // Update existing TIA
        await updateTIA({
          id: activity.transferImpactAssessment.id,
          dataSubjects: formData.dataSubjects || undefined,
          legalMechanism: formData.legalMechanism || undefined,
          transfer: formData.transfer || undefined,
          localLawRisk: formData.localLawRisk || undefined,
          supplementaryMeasures: formData.supplementaryMeasures || undefined,
        });
        toast({
          title: __("Success"),
          description: __("TIA updated successfully"),
          variant: "success",
        });
      } else {
        // Create new TIA
        await createTIA({
          processingActivityId: activity.id!,
          dataSubjects: formData.dataSubjects || undefined,
          legalMechanism: formData.legalMechanism || undefined,
          transfer: formData.transfer || undefined,
          localLawRisk: formData.localLawRisk || undefined,
          supplementaryMeasures: formData.supplementaryMeasures || undefined,
        });
        setTiaDeleted(false);
        setShowTiaForm(true);
        toast({
          title: __("Success"),
          description: __("TIA created successfully"),
          variant: "success",
        });
      }
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to save TIA"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    } finally {
      setTiaSubmitting(false);
    }
  });

  const breadcrumbProcessingActivitiesUrl
    = isSnapshotMode && snapshotId
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
            {
              label: __("Processing Activities"),
              to: breadcrumbProcessingActivitiesUrl,
            },
            { label: activity.name! },
          ]}
        />
        {!isSnapshotMode && activity.canDelete && (
          <ActionDropdown>
            <DropdownItem onClick={deleteActivity} variant="danger">
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <div className="mb-6">
        <h1 className="text-2xl font-bold">{activity.name}</h1>
      </div>

      <Tabs>
        <TabItem
          active={activeTab === "overview"}
          onClick={() => setActiveTab("overview")}
        >
          {__("Overview")}
        </TabItem>
        <TabItem
          active={activeTab === "dpia"}
          onClick={() => setActiveTab("dpia")}
        >
          {__("Data Protection Impact Assessment")}
        </TabItem>
        <TabItem
          active={activeTab === "tia"}
          onClick={() => setActiveTab("tia")}
        >
          {__("Transfer Impact Assessment")}
        </TabItem>
      </Tabs>

      {activeTab === "overview" && (
        <Card>
          <div className="p-6">
            <form onSubmit={e => void onSubmit(e)} className="space-y-6">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <Field
                    label={__("Name")}
                    {...register("name")}
                    error={formState.errors.name?.message}
                    required
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />

                  <div>
                    <Label htmlFor="role">
                      {__("Role")}
                      {" "}
                      *
                    </Label>
                    <Controller
                      control={control}
                      name="role"
                      render={({ field }) => (
                        <Select
                          id="role"
                          placeholder={__("Select role")}
                          onValueChange={field.onChange}
                          value={field.value}
                          className="w-full"
                          disabled={isSnapshotMode || !activity.canUpdate}
                        >
                          <RoleOptions />
                        </Select>
                      )}
                    />
                    {formState.errors.role && (
                      <p className="text-sm text-txt-danger mt-1">
                        {formState.errors.role.message}
                      </p>
                    )}
                  </div>

                  <div>
                    <Label>{__("Purpose")}</Label>
                    <Textarea
                      {...register("purpose")}
                      placeholder={__("Describe the purpose of processing")}
                      rows={3}
                      disabled={isSnapshotMode || !activity.canUpdate}
                    />
                  </div>

                  <Field
                    label={__("Data Subject Category")}
                    {...register("dataSubjectCategory")}
                    placeholder={__("e.g., employees, customers, prospects")}
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />

                  <Field
                    label={__("Personal Data Category")}
                    {...register("personalDataCategory")}
                    placeholder={__("e.g., contact details, financial data")}
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />

                  <div>
                    <Label htmlFor="specialOrCriminalData">
                      {__("Special or Criminal Data")}
                      {" "}
                      *
                    </Label>
                    <Controller
                      control={control}
                      name="specialOrCriminalData"
                      render={({ field }) => (
                        <Select
                          id="specialOrCriminalData"
                          placeholder={__(
                            "Select special or criminal data status",
                          )}
                          onValueChange={field.onChange}
                          value={field.value}
                          className="w-full"
                          disabled={isSnapshotMode || !activity.canUpdate}
                        >
                          <SpecialOrCriminalDataOptions />
                        </Select>
                      )}
                    />
                    {formState.errors.specialOrCriminalData && (
                      <p className="text-sm text-txt-danger mt-1">
                        {formState.errors.specialOrCriminalData.message}
                      </p>
                    )}
                  </div>

                  <Field
                    label={__("Consent Evidence Link")}
                    {...register("consentEvidenceLink")}
                    placeholder={__("Link to consent evidence if applicable")}
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />

                  <div>
                    <Label htmlFor="lawfulBasis">
                      {__("Lawful Basis")}
                      {" "}
                      *
                    </Label>
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
                          disabled={isSnapshotMode || !activity.canUpdate}
                        >
                          <LawfulBasisOptions />
                        </Select>
                      )}
                    />
                    {formState.errors.lawfulBasis && (
                      <p className="text-sm text-txt-danger mt-1">
                        {formState.errors.lawfulBasis.message}
                      </p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="lastReviewDate">
                      {__("Last Review Date")}
                    </Label>
                    <Input
                      id="lastReviewDate"
                      type="date"
                      {...register("lastReviewDate")}
                      disabled={isSnapshotMode || !activity.canUpdate}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="nextReviewDate">
                      {__("Next Review Date")}
                    </Label>
                    <Input
                      id="nextReviewDate"
                      type="date"
                      {...register("nextReviewDate")}
                      disabled={isSnapshotMode || !activity.canUpdate}
                    />
                  </div>

                  <PeopleSelectField
                    organizationId={organizationId}
                    control={control}
                    name="dataProtectionOfficerId"
                    label={__("Data Protection Officer")}
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />
                </div>

                <div className="space-y-4">
                  <Field
                    label={__("Recipients")}
                    {...register("recipients")}
                    placeholder={__("Who receives the data")}
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />

                  <Field
                    label={__("Location")}
                    {...register("location")}
                    placeholder={__("Where is the data processed")}
                    disabled={isSnapshotMode || !activity.canUpdate}
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
                            disabled={isSnapshotMode || !activity.canUpdate}
                          />
                          <span>
                            {__("Data is transferred internationally")}
                          </span>
                        </div>
                      </div>
                    )}
                  />

                  <div>
                    <Label htmlFor="transferSafeguards">
                      {__("Transfer Safeguards")}
                    </Label>
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
                          disabled={isSnapshotMode || !activity.canUpdate}
                        >
                          <TransferSafeguardsOptions />
                        </Select>
                      )}
                    />
                    {formState.errors.transferSafeguards && (
                      <p className="text-sm text-txt-danger mt-1">
                        {formState.errors.transferSafeguards.message}
                      </p>
                    )}
                  </div>

                  <Field
                    label={__("Retention Period")}
                    {...register("retentionPeriod")}
                    placeholder={__("How long is data retained")}
                    disabled={isSnapshotMode || !activity.canUpdate}
                  />

                  <div>
                    <Label>{__("Security Measures")}</Label>
                    <Textarea
                      {...register("securityMeasures")}
                      placeholder={__("Technical and organizational measures")}
                      rows={3}
                      disabled={isSnapshotMode || !activity.canUpdate}
                    />
                  </div>

                  <div>
                    <Label htmlFor="dataProtectionImpactAssessmentNeeded">
                      {__("Data Protection Impact Assessment")}
                      {" "}
                      *
                    </Label>
                    <Controller
                      control={control}
                      name="dataProtectionImpactAssessmentNeeded"
                      render={({ field }) => (
                        <Select
                          id="dataProtectionImpactAssessmentNeeded"
                          placeholder={__("Is DPIA needed?")}
                          onValueChange={field.onChange}
                          value={field.value}
                          className="w-full"
                          disabled={isSnapshotMode || !activity.canUpdate}
                        >
                          <DataProtectionImpactAssessmentOptions />
                        </Select>
                      )}
                    />
                    {formState.errors.dataProtectionImpactAssessmentNeeded && (
                      <p className="text-sm text-txt-danger mt-1">
                        {
                          formState.errors.dataProtectionImpactAssessmentNeeded
                            .message
                        }
                      </p>
                    )}
                  </div>

                  <div>
                    <Label htmlFor="transferImpactAssessmentNeeded">
                      {__("Transfer Impact Assessment")}
                      {" "}
                      *
                    </Label>
                    <Controller
                      control={control}
                      name="transferImpactAssessmentNeeded"
                      render={({ field }) => (
                        <Select
                          id="transferImpactAssessmentNeeded"
                          placeholder={__("Is TIA needed?")}
                          onValueChange={field.onChange}
                          value={field.value}
                          className="w-full"
                          disabled={isSnapshotMode || !activity.canUpdate}
                        >
                          <TransferImpactAssessmentOptions />
                        </Select>
                      )}
                    />
                    {formState.errors.transferImpactAssessmentNeeded && (
                      <p className="text-sm text-txt-danger mt-1">
                        {
                          formState.errors.transferImpactAssessmentNeeded
                            .message
                        }
                      </p>
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
                disabled={isSnapshotMode || !activity.canUpdate}
              />

              {!isSnapshotMode && (
                <div className="flex justify-end pt-4">
                  {activity.canUpdate && (
                    <Button
                      type="submit"
                      variant="primary"
                      disabled={formState.isSubmitting}
                    >
                      {formState.isSubmitting
                        ? __("Saving...")
                        : __("Save Changes")}
                    </Button>
                  )}
                </div>
              )}
            </form>
          </div>
        </Card>
      )}

      {activeTab === "dpia" && (
        <Card>
          <div className="p-6">
            {!showDpiaForm
              && (!activity?.dataProtectionImpactAssessment?.id || dpiaDeleted)
              ? (
                  <div className="flex flex-col items-center justify-center py-16 w-full">
                    <h2 className="text-xl font-semibold mb-6 text-center">
                      {__("Data Protection Impact Assessment")}
                    </h2>
                    {!isSnapshotMode && activity.canCreateDPIA && (
                      <Button
                        variant="primary"
                        onClick={() => setShowDpiaForm(true)}
                      >
                        {__("Create DPIA")}
                      </Button>
                    )}
                  </div>
                )
              : (
                  <>
                    <div className="flex items-center justify-between mb-6">
                      <h2 className="text-xl font-semibold">
                        {__("Data Protection Impact Assessment")}
                      </h2>
                      {!isSnapshotMode
                        && activity?.dataProtectionImpactAssessment?.id
                        && !dpiaDeleted
                        && activity.dataProtectionImpactAssessment.canDelete && (
                        <Button variant="danger" onClick={deleteDPIA}>
                          {__("Delete DPIA")}
                        </Button>
                      )}
                    </div>

                    <form onSubmit={e => void onDPIASubmit(e)} className="space-y-6">
                      <div>
                        <Label htmlFor="dpia-description">
                          {__("Description")}
                        </Label>
                        <Textarea
                          id="dpia-description"
                          {...dpiaForm.register("description")}
                          placeholder={__(
                            "Describe the processing activity and its purpose",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateDPIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="dpia-necessity">
                          {__("Necessity and Proportionality")}
                        </Label>
                        <Textarea
                          id="dpia-necessity"
                          {...dpiaForm.register("necessityAndProportionality")}
                          placeholder={__(
                            "Explain why the processing is necessary and proportionate",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateDPIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="dpia-potential-risk">
                          {__("Potential Risk")}
                        </Label>
                        <Textarea
                          id="dpia-potential-risk"
                          {...dpiaForm.register("potentialRisk")}
                          placeholder={__(
                            "Describe the potential risks to data subjects",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateDPIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="dpia-mitigations">
                          {__("Mitigations")}
                        </Label>
                        <Textarea
                          id="dpia-mitigations"
                          {...dpiaForm.register("mitigations")}
                          placeholder={__(
                            "Describe measures to mitigate the identified risks",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateDPIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="dpia-residual-risk">
                          {__("Residual Risk")}
                        </Label>
                        <Controller
                          control={dpiaForm.control}
                          name="residualRisk"
                          render={({ field }) => (
                            <Select
                              id="dpia-residual-risk"
                              placeholder={__("Select residual risk level")}
                              onValueChange={field.onChange}
                              value={field.value}
                              className="w-full"
                              disabled={isSnapshotMode || !canCreateOrUpdateDPIA}
                            >
                              <Option value="LOW">{__("Low")}</Option>
                              <Option value="MEDIUM">{__("Medium")}</Option>
                              <Option value="HIGH">{__("High")}</Option>
                            </Select>
                          )}
                        />
                      </div>

                      {!isSnapshotMode && (
                        <div className="flex justify-end gap-3 pt-4">
                          {(!activity?.dataProtectionImpactAssessment?.id
                            || dpiaDeleted) && (
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => setShowDpiaForm(false)}
                            >
                              {__("Cancel")}
                            </Button>
                          )}
                          {(activity?.dataProtectionImpactAssessment?.id
                            && !dpiaDeleted
                            ? activity.dataProtectionImpactAssessment.canUpdate
                            : activity.canCreateDPIA) && (
                            <Button
                              type="submit"
                              variant="primary"
                              disabled={dpiaSubmitting}
                            >
                              {dpiaSubmitting
                                ? __("Saving...")
                                : activity?.dataProtectionImpactAssessment?.id
                                  && !dpiaDeleted
                                  ? __("Update DPIA")
                                  : __("Create DPIA")}
                            </Button>
                          )}
                        </div>
                      )}
                    </form>
                  </>
                )}
          </div>
        </Card>
      )}

      {activeTab === "tia" && (
        <Card>
          <div className="p-6">
            {!showTiaForm
              && (!activity?.transferImpactAssessment?.id || tiaDeleted)
              ? (
                  <div className="flex flex-col items-center justify-center py-16 w-full">
                    <h2 className="text-xl font-semibold mb-6 text-center">
                      {__("Transfer Impact Assessment")}
                    </h2>
                    {!isSnapshotMode && activity.canCreateTIA && (
                      <Button
                        variant="primary"
                        onClick={() => setShowTiaForm(true)}
                      >
                        {__("Create TIA")}
                      </Button>
                    )}
                  </div>
                )
              : (
                  <>
                    <div className="flex items-center justify-between mb-6">
                      <h2 className="text-xl font-semibold">
                        {__("Transfer Impact Assessment")}
                      </h2>
                      {!isSnapshotMode
                        && activity?.transferImpactAssessment?.id
                        && !tiaDeleted
                        && activity.transferImpactAssessment.canDelete && (
                        <Button variant="danger" onClick={deleteTIA}>
                          {__("Delete TIA")}
                        </Button>
                      )}
                    </div>

                    <form onSubmit={e => void onTIASubmit(e)} className="space-y-6">
                      <div>
                        <Label htmlFor="tia-data-subjects">
                          {__("Data Subjects")}
                        </Label>
                        <Textarea
                          id="tia-data-subjects"
                          {...tiaForm.register("dataSubjects")}
                          placeholder={__(
                            "Describe the data subjects involved in the transfer",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateTIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="tia-legal-mechanism">
                          {__("Legal Mechanism")}
                        </Label>
                        <Textarea
                          id="tia-legal-mechanism"
                          {...tiaForm.register("legalMechanism")}
                          placeholder={__(
                            "Describe the legal mechanism for the transfer (e.g., SCCs, BCRs, adequacy decision)",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateTIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="tia-transfer">{__("Transfer")}</Label>
                        <Textarea
                          id="tia-transfer"
                          {...tiaForm.register("transfer")}
                          placeholder={__(
                            "Describe the nature and details of the data transfer",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateTIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="tia-local-law-risk">
                          {__("Local Law Risk")}
                        </Label>
                        <Textarea
                          id="tia-local-law-risk"
                          {...tiaForm.register("localLawRisk")}
                          placeholder={__(
                            "Assess the risks related to the local laws of the destination country",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateTIA}
                        />
                      </div>

                      <div>
                        <Label htmlFor="tia-supplementary-measures">
                          {__("Supplementary Measures")}
                        </Label>
                        <Textarea
                          id="tia-supplementary-measures"
                          {...tiaForm.register("supplementaryMeasures")}
                          placeholder={__(
                            "Describe any supplementary measures taken to ensure adequate protection",
                          )}
                          rows={4}
                          disabled={isSnapshotMode || !canCreateOrUpdateTIA}
                        />
                      </div>

                      {!isSnapshotMode && (
                        <div className="flex justify-end gap-3 pt-4">
                          {(!activity?.transferImpactAssessment?.id
                            || tiaDeleted) && (
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => setShowTiaForm(false)}
                            >
                              {__("Cancel")}
                            </Button>
                          )}
                          {(activity?.transferImpactAssessment?.id && !tiaDeleted
                            ? activity.transferImpactAssessment.canUpdate
                            : activity.canCreateTIA) && (
                            <Button
                              type="submit"
                              variant="primary"
                              disabled={tiaSubmitting}
                            >
                              {tiaSubmitting
                                ? __("Saving...")
                                : activity?.transferImpactAssessment?.id
                                  && !tiaDeleted
                                  ? __("Update TIA")
                                  : __("Create TIA")}
                            </Button>
                          )}
                        </div>
                      )}
                    </form>
                  </>
                )}
          </div>
        </Card>
      )}
    </div>
  );
}
