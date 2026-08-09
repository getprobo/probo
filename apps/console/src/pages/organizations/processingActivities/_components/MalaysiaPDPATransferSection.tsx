// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError, type GraphQLError } from "@probo/helpers";
import {
  Badge,
  Button,
  Field,
  Label,
  Option,
  Select,
  Textarea,
  useToast,
} from "@probo/ui";
import { Controller, useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { MalaysiaPDPATransferSection_activity$key } from "#/__generated__/core/MalaysiaPDPATransferSection_activity.graphql";
import {
  type MalaysiaPDPATransferInput,
  useCreateTransferImpactAssessment,
  useUpdateTransferImpactAssessment,
} from "#/hooks/graph/ProcessingActivityGraph";

const fragment = graphql`
  fragment MalaysiaPDPATransferSection_activity on ProcessingActivity {
    id
    canCreateTIA: permission(
      action: "core:transfer-impact-assessment:create"
    )
    organization {
      id
      thirdParties(
        first: 100
        orderBy: { direction: ASC, field: NAME }
        filter: { level: 1 }
      ) {
        edges { node { id name } }
      }
    }
    transferImpactAssessment {
      id
      malaysiaTransferBasis
      malaysiaDestinationCountry
      malaysiaRecipientThirdPartyId
      malaysiaReceiverRegistrationNumber
      malaysiaReceiverContact
      malaysiaTransferPurpose
      malaysiaPersonalDataCategories
      malaysiaSafeguards
      malaysiaApprovalStatus
      malaysiaApprovedByProfileId
      malaysiaApprovalNotes
      malaysiaReviewedAt
      malaysiaNextReviewAt
      malaysiaReviewEvidence
      malaysiaRuleVersion
      malaysiaRuleSource
      canUpdate: permission(
        action: "core:transfer-impact-assessment:update"
      )
    }
  }
`;

type TransferForm = MalaysiaPDPATransferInput;

type Props = {
  activityKey: MalaysiaPDPATransferSection_activity$key;
  deleted?: boolean;
  onTIAAvailable: () => void;
};

const transferBases: TransferForm["basis"][] = [
  "SUBSTANTIALLY_SIMILAR_LAW",
  "ADEQUATE_EQUIVALENT_PROTECTION",
  "DATA_SUBJECT_CONSENT",
  "DATA_SUBJECT_CONTRACT",
  "THIRD_PARTY_CONTRACT",
  "LEGAL_PROCEEDINGS",
  "ADVERSE_ACTION",
  "REASONABLE_PRECAUTIONS",
  "VITAL_INTERESTS",
];

export function MalaysiaPDPATransferSection({
  activityKey,
  deleted = false,
  onTIAAvailable,
}: Props) {
  const activity = useFragment(fragment, activityKey);
  const tia = deleted ? null : activity.transferImpactAssessment;
  const thirdParties = activity.organization.thirdParties.edges.map(
    edge => edge.node,
  );
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const createTIA = useCreateTransferImpactAssessment();
  const updateTIA = useUpdateTransferImpactAssessment();
  const form = useForm<TransferForm>({
    defaultValues: {
      basis: tia?.malaysiaTransferBasis ?? "SUBSTANTIALLY_SIMILAR_LAW",
      destinationCountry: tia?.malaysiaDestinationCountry ?? "",
      recipientThirdPartyId: tia?.malaysiaRecipientThirdPartyId ?? "",
      receiverRegistrationNumber:
        tia?.malaysiaReceiverRegistrationNumber ?? "",
      receiverContact: tia?.malaysiaReceiverContact ?? "",
      transferPurpose: tia?.malaysiaTransferPurpose ?? "",
      personalDataCategories: tia?.malaysiaPersonalDataCategories ?? "",
      safeguards: tia?.malaysiaSafeguards ?? "",
      approvalStatus: tia?.malaysiaApprovalStatus ?? "PENDING",
      approvalNotes: tia?.malaysiaApprovalNotes ?? "",
      reviewEvidence: tia?.malaysiaReviewEvidence ?? "",
    },
  });
  const approvalStatus = useWatch({
    control: form.control,
    name: "approvalStatus",
  });
  const canSave = tia ? tia.canUpdate : activity.canCreateTIA;
  const statusVariant = tia?.malaysiaApprovalStatus === "APPROVED"
    ? "success"
    : tia?.malaysiaApprovalStatus === "REJECTED"
      ? "danger"
      : "warning";

  const formatDate = (value: string | null | undefined) => value
    ? new Intl.DateTimeFormat(i18n.language, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(new Date(value))
    : t("processingActivityDetailsPage.malaysiaTransfer.notAvailable");

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const malaysiaPDPA: MalaysiaPDPATransferInput = {
        ...values,
        destinationCountry: values.destinationCountry.trim().toUpperCase(),
        receiverRegistrationNumber:
          values.receiverRegistrationNumber || undefined,
        approvalNotes: values.approvalNotes || undefined,
        reviewEvidence: values.reviewEvidence || undefined,
      };
      if (tia) {
        await updateTIA({ id: tia.id, malaysiaPDPA });
      } else {
        await createTIA({ processingActivityId: activity.id, malaysiaPDPA });
        onTIAAvailable();
      }
      toast({
        title: t("processingActivityDetailsPage.malaysiaTransfer.savedTitle"),
        description: t("processingActivityDetailsPage.malaysiaTransfer.savedDescription"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("processingActivityDetailsPage.messages.error"),
        description: formatError(
          t("processingActivityDetailsPage.malaysiaTransfer.saveError"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  return (
    <section className="mb-8 space-y-6 border-b pb-8">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-xl font-semibold">
            {t("processingActivityDetailsPage.malaysiaTransfer.title")}
          </h2>
          <p className="mt-1 text-sm text-txt-secondary">
            {t("processingActivityDetailsPage.malaysiaTransfer.description")}
          </p>
        </div>
        {tia?.malaysiaApprovalStatus && (
          <Badge variant={statusVariant}>
            {t(`processingActivityDetailsPage.malaysiaTransfer.status.${tia.malaysiaApprovalStatus}`)}
          </Badge>
        )}
      </div>

      {tia && (
        <div className="grid grid-cols-1 gap-3 rounded-md bg-level-1 p-4 text-sm md:grid-cols-2">
          <div>
            <span className="text-txt-secondary">
              {t("processingActivityDetailsPage.malaysiaTransfer.reviewedAt")}
            </span>
            <p>{formatDate(tia.malaysiaReviewedAt)}</p>
          </div>
          <div>
            <span className="text-txt-secondary">
              {t("processingActivityDetailsPage.malaysiaTransfer.nextReviewAt")}
            </span>
            <p>{formatDate(tia.malaysiaNextReviewAt)}</p>
          </div>
          {tia.malaysiaApprovedByProfileId && (
            <div className="md:col-span-2">
              <span className="text-txt-secondary">
                {t("processingActivityDetailsPage.malaysiaTransfer.approvedBy")}
              </span>
              <p className="break-all">{tia.malaysiaApprovedByProfileId}</p>
            </div>
          )}
          <p className="text-xs text-txt-tertiary md:col-span-2">
            {t("processingActivityDetailsPage.malaysiaTransfer.ruleVersion", {
              version: tia.malaysiaRuleVersion,
            })}
          </p>
          {tia.malaysiaRuleSource && (
            <a
              className="text-xs text-txt-tertiary underline md:col-span-2"
              href={tia.malaysiaRuleSource}
              rel="noreferrer"
              target="_blank"
            >
              {t("processingActivityDetailsPage.malaysiaTransfer.ruleSource")}
            </a>
          )}
        </div>
      )}

      <form className="space-y-5" onSubmit={event => void onSubmit(event)}>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <Label htmlFor="malaysia-transfer-basis">
              {t("processingActivityDetailsPage.malaysiaTransfer.basis")}
            </Label>
            <Controller
              control={form.control}
              name="basis"
              render={({ field }) => (
                <Select
                  id="malaysia-transfer-basis"
                  value={field.value}
                  onValueChange={field.onChange}
                  disabled={!canSave}
                >
                  {transferBases.map(basis => (
                    <Option key={basis} value={basis}>
                      {t(`processingActivityDetailsPage.malaysiaTransfer.bases.${basis}`)}
                    </Option>
                  ))}
                </Select>
              )}
            />
          </div>

          <Field
            label={t("processingActivityDetailsPage.malaysiaTransfer.destinationCountry")}
            maxLength={2}
            placeholder="SG"
            {...form.register("destinationCountry", { required: true })}
            disabled={!canSave}
            required
          />

          <div>
            <Label htmlFor="malaysia-transfer-recipient">
              {t("processingActivityDetailsPage.malaysiaTransfer.recipient")}
            </Label>
            <Controller
              control={form.control}
              name="recipientThirdPartyId"
              rules={{ required: true }}
              render={({ field }) => (
                <Select
                  id="malaysia-transfer-recipient"
                  value={field.value}
                  onValueChange={field.onChange}
                  placeholder={t("processingActivityDetailsPage.malaysiaTransfer.recipientPlaceholder")}
                  disabled={!canSave || thirdParties.length === 0}
                >
                  {thirdParties.map(thirdParty => (
                    <Option key={thirdParty.id} value={thirdParty.id}>
                      {thirdParty.name}
                    </Option>
                  ))}
                </Select>
              )}
            />
          </div>

          <Field
            label={t("processingActivityDetailsPage.malaysiaTransfer.registrationNumber")}
            {...form.register("receiverRegistrationNumber")}
            disabled={!canSave}
          />

          <Field
            label={t("processingActivityDetailsPage.malaysiaTransfer.receiverContact")}
            {...form.register("receiverContact", { required: true })}
            disabled={!canSave}
            required
          />
        </div>

        <div>
          <Label htmlFor="malaysia-transfer-purpose">
            {t("processingActivityDetailsPage.malaysiaTransfer.transferPurpose")}
          </Label>
          <Textarea
            id="malaysia-transfer-purpose"
            rows={3}
            {...form.register("transferPurpose", { required: true })}
            disabled={!canSave}
            required
          />
        </div>
        <div>
          <Label htmlFor="malaysia-transfer-categories">
            {t("processingActivityDetailsPage.malaysiaTransfer.personalDataCategories")}
          </Label>
          <Textarea
            id="malaysia-transfer-categories"
            rows={3}
            {...form.register("personalDataCategories", { required: true })}
            disabled={!canSave}
            required
          />
        </div>
        <div>
          <Label htmlFor="malaysia-transfer-safeguards">
            {t("processingActivityDetailsPage.malaysiaTransfer.safeguards")}
          </Label>
          <Textarea
            id="malaysia-transfer-safeguards"
            rows={4}
            {...form.register("safeguards", { required: true })}
            disabled={!canSave}
            required
          />
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <Label htmlFor="malaysia-transfer-approval-status">
              {t("processingActivityDetailsPage.malaysiaTransfer.approvalStatus")}
            </Label>
            <Controller
              control={form.control}
              name="approvalStatus"
              render={({ field }) => (
                <Select
                  id="malaysia-transfer-approval-status"
                  value={field.value}
                  onValueChange={field.onChange}
                  disabled={!canSave}
                >
                  <Option value="PENDING">
                    {t("processingActivityDetailsPage.malaysiaTransfer.status.PENDING")}
                  </Option>
                  <Option value="APPROVED">
                    {t("processingActivityDetailsPage.malaysiaTransfer.status.APPROVED")}
                  </Option>
                  <Option value="REJECTED">
                    {t("processingActivityDetailsPage.malaysiaTransfer.status.REJECTED")}
                  </Option>
                </Select>
              )}
            />
          </div>
        </div>

        <div>
          <Label htmlFor="malaysia-transfer-approval-notes">
            {t("processingActivityDetailsPage.malaysiaTransfer.approvalNotes")}
          </Label>
          <Textarea
            id="malaysia-transfer-approval-notes"
            rows={3}
            {...form.register("approvalNotes", {
              required: approvalStatus === "REJECTED",
            })}
            disabled={!canSave}
            required={approvalStatus === "REJECTED"}
          />
        </div>

        <div>
          <Label htmlFor="malaysia-transfer-review-evidence">
            {t("processingActivityDetailsPage.malaysiaTransfer.reviewEvidence")}
          </Label>
          <Textarea
            id="malaysia-transfer-review-evidence"
            rows={3}
            {...form.register("reviewEvidence", {
              required: approvalStatus === "APPROVED",
            })}
            disabled={!canSave}
            required={approvalStatus === "APPROVED"}
          />
        </div>

        <p className="text-xs text-txt-tertiary">
          {t("processingActivityDetailsPage.malaysiaTransfer.disclaimer")}
        </p>

        {canSave && (
          <div className="flex justify-end">
            <Button type="submit" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting
                ? t("processingActivityDetailsPage.actions.saving")
                : t("processingActivityDetailsPage.malaysiaTransfer.save")}
            </Button>
          </div>
        )}
      </form>
    </section>
  );
}
