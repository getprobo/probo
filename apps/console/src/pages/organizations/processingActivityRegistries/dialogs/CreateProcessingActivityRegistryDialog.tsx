import { type ReactNode } from "react";
import {
  Button,
  Field,
  useToast,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
  Textarea,
  Breadcrumb,
  Label,
  Checkbox,
  Select,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useCreateProcessingActivityRegistry } from "../../../../hooks/graph/ProcessingActivityRegistryGraph";
import { Controller } from "react-hook-form";
import {
  SpecialOrCriminalDataOptions,
  LawfulBasisOptions,
  TransferSafeguardsOptions,
  DataProtectionImpactAssessmentOptions,
  TransferImpactAssessmentOptions,
} from "../../../../components/form/ProcessingActivityRegistryEnumOptions";

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
  dataProtectionImpactAssessment: z.enum(["NEEDED", "NOT_NEEDED"] as const),
  transferImpactAssessment: z.enum(["NEEDED", "NOT_NEEDED"] as const),
});

type FormData = z.infer<typeof schema>;

interface CreateProcessingActivityRegistryDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionId?: string;
}

export function CreateProcessingActivityRegistryDialog({
  children,
  organizationId,
  connectionId,
}: CreateProcessingActivityRegistryDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createRegistry = useCreateProcessingActivityRegistry(connectionId);

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
      dataProtectionImpactAssessment: "NOT_NEEDED" as const,
      transferImpactAssessment: "NOT_NEEDED" as const,
    },
  });

  const onSubmit = handleSubmit(async (formData: FormData) => {
    try {
      await createRegistry({
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
        dataProtectionImpactAssessment: formData.dataProtectionImpactAssessment || undefined,
        transferImpactAssessment: formData.transferImpactAssessment || undefined,
      });

      toast({
        title: __("Success"),
        description: __("Processing activity entry created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to create processing activity entry"),
        variant: "error",
      });
    }
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Processing Activities"), __("Create Entry")]} />}
      className="max-w-4xl"
    >
      <form onSubmit={onSubmit}>
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
                error={formState.errors.consentEvidenceLink?.message}
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
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            variant="primary"
            disabled={formState.isSubmitting}
          >
            {formState.isSubmitting ? __("Creating...") : __("Create Entry")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
