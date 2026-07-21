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

import { downloadFile } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  Button,
  Card,
  Field,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  Input,
  Option,
} from "@probo/ui";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyOverviewPageBusinessAssociateAgreementFragment$key } from "#/__generated__/core/ThirdPartyOverviewPageBusinessAssociateAgreementFragment.graphql";
import type { ThirdPartyOverviewPageDataPrivacyAgreementFragment$key } from "#/__generated__/core/ThirdPartyOverviewPageDataPrivacyAgreementFragment.graphql";
import type { ThirdPartyOverviewPageQuery } from "#/__generated__/core/ThirdPartyOverviewPageQuery.graphql";
import type { ThirdPartyCategory } from "#/__generated__/core/useThirdPartyFormFragment.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { CountriesField } from "#/components/form/CountriesField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useThirdPartyForm } from "#/hooks/forms/useThirdPartyForm";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DeleteBusinessAssociateAgreementDialog } from "../_components/DeleteBusinessAssociateAgreementDialog";
import { DeleteDataPrivacyAgreementDialog } from "../_components/DeleteDataPrivacyAgreementDialog";
import { EditBusinessAssociateAgreementDialog } from "../_components/EditBusinessAssociateAgreementDialog";
import { EditDataPrivacyAgreementDialog } from "../_components/EditDataPrivacyAgreementDialog";
import { UploadBusinessAssociateAgreementDialog } from "../_components/UploadBusinessAssociateAgreementDialog";
import { UploadDataPrivacyAgreementDialog } from "../_components/UploadDataPrivacyAgreementDialog";

const thirdPartyBusinessAssociateAgreementFragment = graphql`
  fragment ThirdPartyOverviewPageBusinessAssociateAgreementFragment on ThirdParty {
    businessAssociateAgreement {
      id
      file {
        fileName
        downloadUrl
      }
      validFrom
      validUntil
      canUpdate: permission(
        action: "core:thirdParty-business-associate-agreement:update"
      )
      canDelete: permission(
        action: "core:thirdParty-business-associate-agreement:delete"
      )
    }
  }
`;

const thirdPartyDataPrivacyAgreementFragment = graphql`
  fragment ThirdPartyOverviewPageDataPrivacyAgreementFragment on ThirdParty {
    dataPrivacyAgreement {
      id
      file {
        fileName
        downloadUrl
      }
      validFrom
      validUntil
      canUpdate: permission(action: "core:thirdParty-data-privacy-agreement:update")
      canDelete: permission(action: "core:thirdParty-data-privacy-agreement:delete")
    }
  }
`;

export const thirdPartyOverviewPageQuery = graphql`
  query ThirdPartyOverviewPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        name
        canUpdate: permission(action: "core:thirdParty:update")
        canUploadBAA: permission(
          action: "core:thirdParty-business-associate-agreement:upload"
        )
        canUploadDPA: permission(
          action: "core:thirdParty-data-privacy-agreement:upload"
        )
        ...useThirdPartyFormFragment
        ...ThirdPartyOverviewPageBusinessAssociateAgreementFragment
        ...ThirdPartyOverviewPageDataPrivacyAgreementFragment
      }
    }
  }
`;

interface ThirdPartyOverviewPageProps {
  queryRef: PreloadedQuery<ThirdPartyOverviewPageQuery>;
}

export default function ThirdPartyOverviewPage(props: ThirdPartyOverviewPageProps) {
  const data = usePreloadedQuery<ThirdPartyOverviewPageQuery>(thirdPartyOverviewPageQuery, props.queryRef);
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;

  const { t, i18n } = useTranslation();
  const thirdPartyCategories: { value: ThirdPartyCategory; label: string }[] = [
    { value: "ANALYTICS", label: t("thirdPartyOverviewPage.categories.analytics") },
    { value: "CLOUD_MONITORING", label: t("thirdPartyOverviewPage.categories.cloudMonitoring") },
    { value: "CLOUD_PROVIDER", label: t("thirdPartyOverviewPage.categories.cloudProvider") },
    { value: "COLLABORATION", label: t("thirdPartyOverviewPage.categories.collaboration") },
    { value: "CUSTOMER_SUPPORT", label: t("thirdPartyOverviewPage.categories.customerSupport") },
    {
      value: "DATA_STORAGE_AND_PROCESSING",
      label: t("thirdPartyOverviewPage.categories.dataStorageAndProcessing"),
    },
    { value: "DOCUMENT_MANAGEMENT", label: t("thirdPartyOverviewPage.categories.documentManagement") },
    { value: "EMPLOYEE_MANAGEMENT", label: t("thirdPartyOverviewPage.categories.employeeManagement") },
    { value: "ENGINEERING", label: t("thirdPartyOverviewPage.categories.engineering") },
    { value: "FINANCE", label: t("thirdPartyOverviewPage.categories.finance") },
    { value: "IDENTITY_PROVIDER", label: t("thirdPartyOverviewPage.categories.identityProvider") },
    { value: "IT", label: t("thirdPartyOverviewPage.categories.it") },
    { value: "MARKETING", label: t("thirdPartyOverviewPage.categories.marketing") },
    { value: "OFFICE_OPERATIONS", label: t("thirdPartyOverviewPage.categories.officeOperations") },
    { value: "OTHER", label: t("thirdPartyOverviewPage.categories.other") },
    { value: "PASSWORD_MANAGEMENT", label: t("thirdPartyOverviewPage.categories.passwordManagement") },
    { value: "PRODUCT_AND_DESIGN", label: t("thirdPartyOverviewPage.categories.productAndDesign") },
    { value: "PROFESSIONAL_SERVICES", label: t("thirdPartyOverviewPage.categories.professionalServices") },
    { value: "RECRUITING", label: t("thirdPartyOverviewPage.categories.recruiting") },
    { value: "SALES", label: t("thirdPartyOverviewPage.categories.sales") },
    { value: "SECURITY", label: t("thirdPartyOverviewPage.categories.security") },
    { value: "VERSION_CONTROL", label: t("thirdPartyOverviewPage.categories.versionControl") },
  ];
  const organizationId = useOrganizationId();

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useThirdPartyForm(thirdParty);

  const thirdPartyWithBAA
    = useFragment<ThirdPartyOverviewPageBusinessAssociateAgreementFragment$key>(
      thirdPartyBusinessAssociateAgreementFragment,
      thirdParty,
    );
  const businessAssociateAgreement = thirdPartyWithBAA.businessAssociateAgreement;

  const thirdPartyWithDPA
    = useFragment<ThirdPartyOverviewPageDataPrivacyAgreementFragment$key>(
      thirdPartyDataPrivacyAgreementFragment,
      thirdParty,
    );
  const dataPrivacyAgreement = thirdPartyWithDPA.dataPrivacyAgreement;

  const urls = useMemo(
    () =>
      [
        { name: "statusPageUrl", label: t("thirdPartyOverviewPage.urlLabels.statusPage") },
        { name: "termsOfServiceUrl", label: t("thirdPartyOverviewPage.urlLabels.termsOfService") },
        { name: "privacyPolicyUrl", label: t("thirdPartyOverviewPage.urlLabels.privacyPolicy") },
        {
          name: "serviceLevelAgreementUrl",
          label: t("thirdPartyOverviewPage.urlLabels.serviceLevelAgreement"),
        },
        {
          name: "dataProcessingAgreementUrl",
          label: t("thirdPartyOverviewPage.urlLabels.dataProcessingAgreement"),
        },
        { name: "securityPageUrl", label: t("thirdPartyOverviewPage.urlLabels.securityPage") },
        { name: "trustPageUrl", label: t("thirdPartyOverviewPage.urlLabels.trustPage") },
      ] as const,
    [t],
  );

  usePageTitle(t("thirdPartyOverviewPage.pageTitle", { name: thirdParty.name }));

  const isFormDisabled = isSubmitting || !thirdParty.canUpdate;

  return (
    <form
      onSubmit={!thirdParty.canUpdate
        ? undefined
        : e => void handleSubmit(e)}
      className="space-y-12"
    >
      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("thirdPartyOverviewPage.sections.details")}</h2>
        <Card className="space-y-4" padded>
          <Field
            {...register("name")}
            label={t("thirdPartyOverviewPage.fields.name")}
            type="text"
            error={errors.name?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("description")}
            label={t("thirdPartyOverviewPage.fields.description")}
            type="textarea"
            error={errors.description?.message}
            disabled={isFormDisabled}
          />
          <ControlledField
            control={control}
            name="category"
            type="select"
            label={t("thirdPartyOverviewPage.fields.category")}
            placeholder={t("thirdPartyOverviewPage.placeholders.category")}
            error={errors.category?.message}
            disabled={isFormDisabled}
          >
            {thirdPartyCategories.map(category => (
              <Option key={category.value} value={category.value}>
                {category.label}
              </Option>
            ))}
          </ControlledField>
          <Field
            {...register("legalName")}
            label={t("thirdPartyOverviewPage.fields.legalName")}
            type="text"
            error={errors.legalName?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("headquarterAddress")}
            label={t("thirdPartyOverviewPage.fields.headquarterAddress")}
            type="textarea"
            error={errors.headquarterAddress?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("websiteUrl")}
            label={t("thirdPartyOverviewPage.fields.websiteUrl")}
            type="text"
            error={errors.websiteUrl?.message}
            disabled={isFormDisabled}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("thirdPartyOverviewPage.sections.countries")}</h2>
        <Card padded>
          <CountriesField
            control={control}
            name="countries"
            disabled={isFormDisabled}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("thirdPartyOverviewPage.sections.ownership")}</h2>
        <Card className="space-y-4" padded>
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="businessOwnerId"
            label={t("thirdPartyOverviewPage.fields.businessOwner")}
            error={errors.businessOwnerId?.message}
            disabled={isFormDisabled}
            optional={true}
          />
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="securityOwnerId"
            label={t("thirdPartyOverviewPage.fields.securityOwner")}
            error={errors.securityOwnerId?.message}
            disabled={isFormDisabled}
            optional={true}
          />
        </Card>
      </div>

      <div className="space-y-4 mb-4">
        <h2 className="text-base font-medium">{t("thirdPartyOverviewPage.sections.links")}</h2>
        <Card className="divide-y divide-border-low">
          {urls.map(url => (
            <div
              key={url.name}
              className="grid grid-cols-2 items-center divide-x divide-border-low"
            >
              <label
                className="p-4 text-sm font-medium text-txt-secondary"
                htmlFor={url.name}
              >
                {url.label}
              </label>
              <Input
                className="p-4 focus:bg-tertiary-pressed outline-none"
                id={url.name}
                key={url.name}
                {...register(url.name)}
                type="text"
                placeholder="https://..."
                variant="ghost"
                disabled={isFormDisabled}
              />
            </div>
          ))}
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("thirdPartyOverviewPage.sections.dataAgreements")}</h2>
        <Card className="space-y-4" padded>
          <div className="flex items-center justify-between p-4 border border-border-low rounded-lg">
            <div className="flex-1">
              <h3 className="font-medium text-txt-primary">
                {t("thirdPartyOverviewPage.agreements.businessAssociate")}
              </h3>
              <p className="text-sm text-txt-secondary mt-1">
                {businessAssociateAgreement
                  ? businessAssociateAgreement.file.fileName
                  : t("thirdPartyOverviewPage.agreements.noBusinessAssociate")}
              </p>
              {(businessAssociateAgreement?.validFrom
                || businessAssociateAgreement?.validUntil) && (
                <p className="text-xs text-txt-secondary mt-1">
                  {formatValidity(
                    businessAssociateAgreement.validFrom,
                    businessAssociateAgreement.validUntil,
                    i18n.language,
                    t,
                  )}
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              {businessAssociateAgreement
                ? (
                    <>
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() =>
                          downloadFile(
                            businessAssociateAgreement.file.downloadUrl,
                            businessAssociateAgreement.file.fileName,
                          )}
                      >
                        {t("thirdPartyOverviewPage.actions.downloadPdf")}
                      </Button>
                      {businessAssociateAgreement.canUpdate && (
                        <EditBusinessAssociateAgreementDialog
                          thirdPartyId={thirdParty.id}
                          agreement={{
                            validFrom: businessAssociateAgreement.validFrom,
                            validUntil: businessAssociateAgreement.validUntil,
                          }}
                          onSuccess={() => window.location.reload()}
                        >
                          <Button variant="quaternary" icon={IconPencil} />
                        </EditBusinessAssociateAgreementDialog>
                      )}
                      {businessAssociateAgreement.canDelete && (
                        <DeleteBusinessAssociateAgreementDialog
                          thirdPartyId={thirdParty.id}
                          fileName={businessAssociateAgreement.file.fileName}
                          onSuccess={() => window.location.reload()}
                        >
                          <Button variant="quaternary" icon={IconTrashCan} />
                        </DeleteBusinessAssociateAgreementDialog>
                      )}
                    </>
                  )
                : (
                    thirdParty.canUploadBAA && (
                      <UploadBusinessAssociateAgreementDialog
                        thirdPartyId={thirdParty.id}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="secondary" icon={IconPlusLarge}>
                          {t("thirdPartyOverviewPage.actions.upload")}
                        </Button>
                      </UploadBusinessAssociateAgreementDialog>
                    )
                  )}
            </div>
          </div>

          <div className="flex items-center justify-between p-4 border border-border-low rounded-lg">
            <div className="flex-1">
              <h3 className="font-medium text-txt-primary">
                {t("thirdPartyOverviewPage.agreements.dataPrivacy")}
              </h3>
              <p className="text-sm text-txt-secondary mt-1">
                {dataPrivacyAgreement
                  ? dataPrivacyAgreement.file.fileName
                  : t("thirdPartyOverviewPage.agreements.noDataPrivacy")}
              </p>
              {(dataPrivacyAgreement?.validFrom
                || dataPrivacyAgreement?.validUntil) && (
                <p className="text-xs text-txt-secondary mt-1">
                  {formatValidity(
                    dataPrivacyAgreement.validFrom,
                    dataPrivacyAgreement.validUntil,
                    i18n.language,
                    t,
                  )}
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              {dataPrivacyAgreement
                ? (
                    <>
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() =>
                          downloadFile(
                            dataPrivacyAgreement.file.downloadUrl,
                            dataPrivacyAgreement.file.fileName,
                          )}
                      >
                        {t("thirdPartyOverviewPage.actions.downloadPdf")}
                      </Button>
                      {dataPrivacyAgreement.canUpdate && (
                        <EditDataPrivacyAgreementDialog
                          thirdPartyId={thirdParty.id}
                          agreement={{
                            validFrom: dataPrivacyAgreement.validFrom,
                            validUntil: dataPrivacyAgreement.validUntil,
                          }}
                          onSuccess={() => window.location.reload()}
                        >
                          <Button variant="quaternary" icon={IconPencil} />
                        </EditDataPrivacyAgreementDialog>
                      )}
                      {dataPrivacyAgreement.canDelete && (
                        <DeleteDataPrivacyAgreementDialog
                          thirdPartyId={thirdParty.id}
                          fileName={dataPrivacyAgreement.file.fileName}
                          onSuccess={() => window.location.reload()}
                        >
                          <Button variant="quaternary" icon={IconTrashCan} />
                        </DeleteDataPrivacyAgreementDialog>
                      )}
                    </>
                  )
                : (
                    thirdParty.canUploadDPA && (
                      <UploadDataPrivacyAgreementDialog
                        thirdPartyId={thirdParty.id}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="secondary" icon={IconPlusLarge}>
                          {t("thirdPartyOverviewPage.actions.upload")}
                        </Button>
                      </UploadDataPrivacyAgreementDialog>
                    )
                  )}
            </div>
          </div>
        </Card>
      </div>

      <div className="flex justify-end">
        {thirdParty.canUpdate && (
          <Button type="submit" disabled={isSubmitting}>
            {t("thirdPartyOverviewPage.actions.update")}
          </Button>
        )}
      </div>
    </form>
  );
}

function formatValidity(
  validFrom: string | null | undefined,
  validUntil: string | null | undefined,
  language: string,
  t: ReturnType<typeof useTranslation>["t"],
) {
  if (validFrom && validUntil) {
    return t("thirdPartyOverviewPage.agreements.validity.range", {
      from: dateFormat(language, validFrom),
      until: dateFormat(language, validUntil),
    });
  }
  if (validFrom) return t("thirdPartyOverviewPage.agreements.validity.from", { date: dateFormat(language, validFrom) });
  if (validUntil) return t("thirdPartyOverviewPage.agreements.validity.until", { date: dateFormat(language, validUntil) });
  return "";
}
