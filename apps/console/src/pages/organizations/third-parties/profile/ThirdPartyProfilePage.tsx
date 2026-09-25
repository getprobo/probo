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

import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  Field,
  Option,
  PageHeader,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { ThirdPartyProfilePageQuery } from "#/__generated__/core/ThirdPartyProfilePageQuery.graphql";
import type { ThirdPartyCategory } from "#/__generated__/core/useThirdPartyProfileFormFragment.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { CountriesField } from "#/components/form/CountriesField";

import { useThirdPartyProfileForm } from "./_lib/useThirdPartyProfileForm";

export const thirdPartyProfilePageQuery = graphql`
  query ThirdPartyProfilePageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        name
        canUpdate: permission(action: "core:thirdParty:update")
        ...useThirdPartyProfileFormFragment
      }
    }
  }
`;

interface ThirdPartyProfilePageProps {
  queryRef: PreloadedQuery<ThirdPartyProfilePageQuery>;
}

export function ThirdPartyProfilePage({ queryRef }: ThirdPartyProfilePageProps) {
  const data = usePreloadedQuery<ThirdPartyProfilePageQuery>(
    thirdPartyProfilePageQuery,
    queryRef,
  );
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;

  const { t } = useTranslation();
  const thirdPartyCategories: { value: ThirdPartyCategory; label: string }[] = [
    { value: "ANALYTICS", label: t("thirdPartyProfilePage.categories.analytics") },
    { value: "CLOUD_MONITORING", label: t("thirdPartyProfilePage.categories.cloudMonitoring") },
    { value: "CLOUD_PROVIDER", label: t("thirdPartyProfilePage.categories.cloudProvider") },
    { value: "COLLABORATION", label: t("thirdPartyProfilePage.categories.collaboration") },
    { value: "CUSTOMER_SUPPORT", label: t("thirdPartyProfilePage.categories.customerSupport") },
    {
      value: "DATA_STORAGE_AND_PROCESSING",
      label: t("thirdPartyProfilePage.categories.dataStorageAndProcessing"),
    },
    { value: "DOCUMENT_MANAGEMENT", label: t("thirdPartyProfilePage.categories.documentManagement") },
    { value: "EMPLOYEE_MANAGEMENT", label: t("thirdPartyProfilePage.categories.employeeManagement") },
    { value: "ENGINEERING", label: t("thirdPartyProfilePage.categories.engineering") },
    { value: "FINANCE", label: t("thirdPartyProfilePage.categories.finance") },
    { value: "IDENTITY_PROVIDER", label: t("thirdPartyProfilePage.categories.identityProvider") },
    { value: "IT", label: t("thirdPartyProfilePage.categories.it") },
    { value: "MARKETING", label: t("thirdPartyProfilePage.categories.marketing") },
    { value: "OFFICE_OPERATIONS", label: t("thirdPartyProfilePage.categories.officeOperations") },
    { value: "OTHER", label: t("thirdPartyProfilePage.categories.other") },
    { value: "PASSWORD_MANAGEMENT", label: t("thirdPartyProfilePage.categories.passwordManagement") },
    { value: "PRODUCT_AND_DESIGN", label: t("thirdPartyProfilePage.categories.productAndDesign") },
    { value: "PROFESSIONAL_SERVICES", label: t("thirdPartyProfilePage.categories.professionalServices") },
    { value: "RECRUITING", label: t("thirdPartyProfilePage.categories.recruiting") },
    { value: "SALES", label: t("thirdPartyProfilePage.categories.sales") },
    { value: "SECURITY", label: t("thirdPartyProfilePage.categories.security") },
    { value: "VERSION_CONTROL", label: t("thirdPartyProfilePage.categories.versionControl") },
  ];

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useThirdPartyProfileForm(thirdParty);

  usePageTitle(t("thirdPartyProfilePage.pageTitle", { name: thirdParty.name }));

  const isFormDisabled = isSubmitting || !thirdParty.canUpdate;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartyProfilePage.title")}
        description={t("thirdPartyProfilePage.description")}
      />
      <form
        onSubmit={!thirdParty.canUpdate
          ? undefined
          : e => void handleSubmit(e)}
        className="space-y-12"
      >
        <div className="space-y-4">
          <h2 className="text-base font-medium">{t("thirdPartyProfilePage.sections.details")}</h2>
          <Card className="space-y-4" padded>
            <Field
              {...register("name")}
              label={t("thirdPartyProfilePage.fields.name")}
              type="text"
              error={errors.name?.message}
              disabled={isFormDisabled}
            />
            <Field
              {...register("description")}
              label={t("thirdPartyProfilePage.fields.description")}
              type="textarea"
              error={errors.description?.message}
              disabled={isFormDisabled}
            />
            <ControlledField
              control={control}
              name="category"
              type="select"
              label={t("thirdPartyProfilePage.fields.category")}
              placeholder={t("thirdPartyProfilePage.placeholders.category")}
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
              label={t("thirdPartyProfilePage.fields.legalName")}
              type="text"
              error={errors.legalName?.message}
              disabled={isFormDisabled}
            />
            <Field
              {...register("headquarterAddress")}
              label={t("thirdPartyProfilePage.fields.headquarterAddress")}
              type="textarea"
              error={errors.headquarterAddress?.message}
              disabled={isFormDisabled}
            />
            <Field
              {...register("websiteUrl")}
              label={t("thirdPartyProfilePage.fields.websiteUrl")}
              type="text"
              error={errors.websiteUrl?.message}
              disabled={isFormDisabled}
            />
          </Card>
        </div>

        <div className="space-y-4">
          <h2 className="text-base font-medium">{t("thirdPartyProfilePage.sections.countries")}</h2>
          <Card padded>
            <CountriesField
              control={control}
              name="countries"
              disabled={isFormDisabled}
            />
          </Card>
        </div>

        <div className="flex justify-end">
          {thirdParty.canUpdate && (
            <Button type="submit" disabled={isSubmitting}>
              {t("thirdPartyProfilePage.actions.update")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
