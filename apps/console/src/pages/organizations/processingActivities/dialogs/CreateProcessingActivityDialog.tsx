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

import { formatDatetime, formatError, type GraphQLError } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
import { z } from "zod";

import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
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

type FormData = {
  name: string; purpose?: string; dataSubjectCategory?: string; personalDataCategory?: string;
  specialOrCriminalData: "YES" | "NO" | "POSSIBLE"; consentEvidenceLink?: string;
  lawfulBasis: "CONSENT" | "CONTRACTUAL_NECESSITY" | "LEGAL_OBLIGATION" | "LEGITIMATE_INTEREST" | "PUBLIC_TASK" | "VITAL_INTERESTS";
  recipients?: string; location?: string; internationalTransfers: boolean; transferSafeguards: string;
  retentionPeriod?: string; securityMeasures?: string; dataProtectionImpactAssessmentNeeded: "NEEDED" | "NOT_NEEDED";
  transferImpactAssessmentNeeded: "NEEDED" | "NOT_NEEDED"; lastReviewDate?: string; nextReviewDate?: string;
  role: "CONTROLLER" | "PROCESSOR"; dataProtectionOfficerId?: string; thirdPartyIds?: string[];
};

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
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createProcessingActivity = useCreateProcessingActivity(connectionId);
  const schema = z.object({
    name: z.string().min(1, t("createProcessingActivityDialog.validation.nameRequired")), purpose: z.string().optional(), dataSubjectCategory: z.string().optional(), personalDataCategory: z.string().optional(), specialOrCriminalData: z.enum(["YES", "NO", "POSSIBLE"] as const), consentEvidenceLink: z.string().optional(), lawfulBasis: z.enum(["CONSENT", "CONTRACTUAL_NECESSITY", "LEGAL_OBLIGATION", "LEGITIMATE_INTEREST", "PUBLIC_TASK", "VITAL_INTERESTS"] as const), recipients: z.string().optional(), location: z.string().optional(), internationalTransfers: z.boolean(), transferSafeguards: z.string(), retentionPeriod: z.string().optional(), securityMeasures: z.string().optional(), dataProtectionImpactAssessmentNeeded: z.enum(["NEEDED", "NOT_NEEDED"] as const), transferImpactAssessmentNeeded: z.enum(["NEEDED", "NOT_NEEDED"] as const), lastReviewDate: z.string().optional(), nextReviewDate: z.string().optional(), role: z.enum(["CONTROLLER", "PROCESSOR"] as const), dataProtectionOfficerId: z.string().optional(), thirdPartyIds: z.array(z.string()).optional(),
  });

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
      thirdPartyIds: [],
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
        thirdPartyIds: formData.thirdPartyIds,
      });

      toast({
        title: t("createProcessingActivityDialog.messages.success"),
        description: t("createProcessingActivityDialog.messages.created"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: t("createProcessingActivityDialog.messages.error"),
        description: formatError(t("createProcessingActivityDialog.errors.create"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[t("createProcessingActivityDialog.breadcrumb.activities"), t("createProcessingActivityDialog.breadcrumb.create")]} />}
      className="max-w-4xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="space-y-4">
              <Field
                label={t("processingActivityDetailsPage.fields.name")}
                {...register("name")}
                placeholder={t("createProcessingActivityDialog.placeholders.name")}
                error={formState.errors.name?.message}
                required
              />

              <div>
                <Label htmlFor="role">{t("processingActivityDetailsPage.fields.role")}</Label>
                <Controller
                  control={control}
                  name="role"
                  render={({ field }) => (
                    <Select
                      id="role"
                      placeholder={t("processingActivityDetailsPage.placeholders.role")}
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
                <Label>{t("processingActivityDetailsPage.fields.purpose")}</Label>
                <Textarea
                  {...register("purpose")}
                  placeholder={t("processingActivityDetailsPage.placeholders.purpose")}
                  rows={3}
                />
              </div>

              <Field
                label={t("processingActivityDetailsPage.fields.dataSubjectCategory")}
                {...register("dataSubjectCategory")}
                placeholder={t("processingActivityDetailsPage.placeholders.dataSubjectCategory")}
                error={formState.errors.dataSubjectCategory?.message}
              />

              <Field
                label={t("processingActivityDetailsPage.fields.personalDataCategory")}
                {...register("personalDataCategory")}
                placeholder={t("processingActivityDetailsPage.placeholders.personalDataCategory")}
                error={formState.errors.personalDataCategory?.message}
              />

              <div>
                <Label htmlFor="specialOrCriminalData">
                  {t("processingActivityDetailsPage.fields.specialOrCriminalData")}
                  {" "}
                  *
                </Label>
                <Controller
                  control={control}
                  name="specialOrCriminalData"
                  render={({ field }) => (
                    <Select
                      id="specialOrCriminalData"
                      placeholder={t("processingActivityDetailsPage.placeholders.specialOrCriminalData")}
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
                label={t("processingActivityDetailsPage.fields.consentEvidenceLink")}
                {...register("consentEvidenceLink")}
                placeholder={t("processingActivityDetailsPage.placeholders.consentEvidenceLink")}
                error={formState.errors.consentEvidenceLink?.message}
              />

              <div>
                <Label htmlFor="lawfulBasis">
                  {t("processingActivityDetailsPage.fields.lawfulBasis")}
                  {" "}
                  *
                </Label>
                <Controller
                  control={control}
                  name="lawfulBasis"
                  render={({ field }) => (
                    <Select
                      id="lawfulBasis"
                      placeholder={t("processingActivityDetailsPage.placeholders.lawfulBasis")}
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
                <Label htmlFor="lastReviewDate">{t("processingActivityDetailsPage.fields.lastReviewDate")}</Label>
                <Input
                  id="lastReviewDate"
                  type="date"
                  {...register("lastReviewDate")}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="nextReviewDate">{t("processingActivityDetailsPage.fields.nextReviewDate")}</Label>
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
                label={t("processingActivityDetailsPage.fields.dataProtectionOfficer")}
              />
            </div>

            <div className="space-y-4">
              <Field
                label={t("processingActivityDetailsPage.fields.recipients")}
                {...register("recipients")}
                placeholder={t("processingActivityDetailsPage.placeholders.recipients")}
                error={formState.errors.recipients?.message}
              />

              <Field
                label={t("processingActivityDetailsPage.fields.location")}
                {...register("location")}
                placeholder={t("processingActivityDetailsPage.placeholders.location")}
                error={formState.errors.location?.message}
              />

              <Controller
                control={control}
                name="internationalTransfers"
                render={({ field }) => (
                  <div>
                    <Label>{t("processingActivityDetailsPage.fields.internationalTransfers")}</Label>
                    <div className="mt-2 flex items-center gap-2">
                      <Checkbox
                        checked={field.value ?? false}
                        onChange={field.onChange}
                      />
                      <span>{t("processingActivityDetailsPage.fields.internationalTransfersDescription")}</span>
                    </div>
                  </div>
                )}
              />

              <div>
                <Label htmlFor="transferSafeguards">{t("processingActivityDetailsPage.fields.transferSafeguards")}</Label>
                <Controller
                  control={control}
                  name="transferSafeguards"
                  render={({ field }) => (
                    <Select
                      id="transferSafeguards"
                      placeholder={t("processingActivityDetailsPage.placeholders.transferSafeguards")}
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
                label={t("processingActivityDetailsPage.fields.retentionPeriod")}
                {...register("retentionPeriod")}
                placeholder={t("processingActivityDetailsPage.placeholders.retentionPeriod")}
                error={formState.errors.retentionPeriod?.message}
              />

              <div>
                <Label>{t("processingActivityDetailsPage.fields.securityMeasures")}</Label>
                <Textarea
                  {...register("securityMeasures")}
                  placeholder={t("processingActivityDetailsPage.placeholders.securityMeasures")}
                  rows={2}
                />
              </div>

              <div>
                <Label htmlFor="dataProtectionImpactAssessmentNeeded">
                  {t("processingActivityDetailsPage.fields.dpiaNeeded")}
                  {" "}
                  *
                </Label>
                <Controller
                  control={control}
                  name="dataProtectionImpactAssessmentNeeded"
                  render={({ field }) => (
                    <Select
                      id="dataProtectionImpactAssessmentNeeded"
                      placeholder={t("processingActivityDetailsPage.placeholders.dpiaNeeded")}
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
                  {t("processingActivityDetailsPage.fields.tiaNeeded")}
                  {" "}
                  *
                </Label>
                <Controller
                  control={control}
                  name="transferImpactAssessmentNeeded"
                  render={({ field }) => (
                    <Select
                      id="transferImpactAssessmentNeeded"
                      placeholder={t("processingActivityDetailsPage.placeholders.tiaNeeded")}
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

          <ThirdPartiesMultiSelectField
            organizationId={organizationId}
            control={control}
            name="thirdPartyIds"
            selectedThirdParties={[]}
            label={t("processingActivityDetailsPage.fields.thirdParties")}
          />
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            variant="primary"
            disabled={formState.isSubmitting}
          >
            {formState.isSubmitting ? t("createProcessingActivityDialog.actions.creating") : t("createProcessingActivityDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
