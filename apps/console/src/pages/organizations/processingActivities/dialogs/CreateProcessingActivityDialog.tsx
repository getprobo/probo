import { formatDatetime, formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Select,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { Controller } from "react-hook-form";
import { z } from "zod";

import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { VendorsMultiSelectField } from "#/components/form/VendorsMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import {
  DataProtectionImpactAssessmentOptions,
  LawfulBasisOptions,
  RoleOptions,
  SpecialOrCriminalDataOptions,
  TransferImpactAssessmentOptions,
  TransferSafeguardsOptions,
} from "../../../../components/form/ProcessingActivityEnumOptions";
import { useCreateProcessingActivity } from "../../../../hooks/graph/ProcessingActivityGraph";

const schema = z.object({
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
  dataProtectionImpactAssessmentNeeded: z.enum(["NEEDED", "NOT_NEEDED"] as const),
  transferImpactAssessmentNeeded: z.enum(["NEEDED", "NOT_NEEDED"] as const),
  lastReviewDate: z.string().optional(),
  nextReviewDate: z.string().optional(),
  role: z.enum(["CONTROLLER", "PROCESSOR"] as const),
  dataProtectionOfficerId: z.string().optional(),
  vendorIds: z.array(z.string()).optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateProcessingActivityDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionId?: string;
}

export function CreateProcessingActivityDialog({
  children,
  organizationId,
  connectionId,
}: CreateProcessingActivityDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createProcessingActivity = useCreateProcessingActivity(connectionId);

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      name: "",
      purpose: "",
      dataSubjectCategory: "",
      personalDataCategory: "",
      specialOrCriminalData: "NO" as const,
      consentEvidenceLink: "",
      lawfulBasis: "LEGITIMATE_INTEREST" as const,
      recipients: "",
      location: "",
      internationalTransfers: false,
      transferSafeguards: "__NONE__",
      retentionPeriod: "",
      securityMeasures: "",
      dataProtectionImpactAssessmentNeeded: "NOT_NEEDED" as const,
      transferImpactAssessmentNeeded: "NOT_NEEDED" as const,
      lastReviewDate: "",
      nextReviewDate: "",
      role: "PROCESSOR" as const,
      dataProtectionOfficerId: "",
      vendorIds: [],
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createProcessingActivity({
        organizationId,
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
        dataProtectionImpactAssessmentNeeded: formData.dataProtectionImpactAssessmentNeeded || undefined,
        transferImpactAssessmentNeeded: formData.transferImpactAssessmentNeeded || undefined,
        lastReviewDate: formatDatetime(formData.lastReviewDate),
        nextReviewDate: formatDatetime(formData.nextReviewDate),
        role: formData.role,
        dataProtectionOfficerId: formData.dataProtectionOfficerId || undefined,
        vendorIds: formData.vendorIds,
      });

      toast({
        title: __("Success"),
        description: __("Processing activity created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to create processing activity"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Processing Activities"), __("Create Processing Activity")]} />}
      className="max-w-4xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="space-y-4">
              <Field
                label={__("Name")}
                {...register("name")}
                placeholder={__("Processing activity name")}
                error={formState.errors.name?.message}
                required
              />

              <div>
                <Label htmlFor="role">{__("Role")}</Label>
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
                    >
                      <RoleOptions />
                    </Select>
                  )}
                />
                {formState.errors.role && (
                  <p className="text-sm text-txt-danger mt-1">{formState.errors.role.message}</p>
                )}
              </div>

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
                error={formState.errors.dataSubjectCategory?.message}
              />

              <Field
                label={__("Personal Data Category")}
                {...register("personalDataCategory")}
                placeholder={__("e.g., contact details, financial data")}
                error={formState.errors.personalDataCategory?.message}
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
                error={formState.errors.consentEvidenceLink?.message}
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
                    >
                      <LawfulBasisOptions />
                    </Select>
                  )}
                />
                {formState.errors.lawfulBasis && (
                  <p className="text-sm text-txt-danger mt-1">{formState.errors.lawfulBasis.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="lastReviewDate">{__("Last Review Date")}</Label>
                <Input
                  id="lastReviewDate"
                  type="date"
                  {...register("lastReviewDate")}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="nextReviewDate">{__("Next Review Date")}</Label>
                <Input
                  id="nextReviewDate"
                  type="date"
                  {...register("nextReviewDate")}
                />
              </div>

              <PeopleSelectField
                organizationId={organizationId}
                control={control}
                name="dataProtectionOfficerId"
                label={__("Data Protection Officer")}
              />
            </div>

            <div className="space-y-4">
              <Field
                label={__("Recipients")}
                {...register("recipients")}
                placeholder={__("Who receives the data")}
                error={formState.errors.recipients?.message}
              />

              <Field
                label={__("Location")}
                {...register("location")}
                placeholder={__("Where is the data processed")}
                error={formState.errors.location?.message}
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
                error={formState.errors.retentionPeriod?.message}
              />

              <div>
                <Label>{__("Security Measures")}</Label>
                <Textarea
                  {...register("securityMeasures")}
                  placeholder={__("Technical and organizational measures")}
                  rows={2}
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
                    >
                      <DataProtectionImpactAssessmentOptions />
                    </Select>
                  )}
                />
                {formState.errors.dataProtectionImpactAssessmentNeeded && (
                  <p className="text-sm text-txt-danger mt-1">{formState.errors.dataProtectionImpactAssessmentNeeded.message}</p>
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
                    >
                      <TransferImpactAssessmentOptions />
                    </Select>
                  )}
                />
                {formState.errors.transferImpactAssessmentNeeded && (
                  <p className="text-sm text-txt-danger mt-1">{formState.errors.transferImpactAssessmentNeeded.message}</p>
                )}
              </div>
            </div>
          </div>

          <VendorsMultiSelectField
            organizationId={organizationId}
            control={control}
            name="vendorIds"
            selectedVendors={[]}
            label={__("Vendors")}
          />
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            variant="primary"
            disabled={formState.isSubmitting}
          >
            {formState.isSubmitting ? __("Creating...") : __("Create Processing Activity")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
