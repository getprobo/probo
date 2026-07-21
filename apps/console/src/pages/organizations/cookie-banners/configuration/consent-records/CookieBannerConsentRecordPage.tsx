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

import { dateFormat, humanizeSeconds } from "@probo/i18n";
import { Badge, Breadcrumb, Card, PageHeader, PropertyRow } from "@probo/ui";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CookieBannerConsentRecordPageQuery } from "#/__generated__/core/CookieBannerConsentRecordPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  formatAnonymizedIp,
  getActionVariant,
} from "./_components/consentRecordHelpers";

export const cookieBannerConsentRecordPageQuery = graphql`
  query CookieBannerConsentRecordPageQuery($consentRecordId: ID!) {
    node(id: $consentRecordId) @required(action: THROW) {
      __typename
      ... on CookieConsentRecord {
        id
        visitorId
        action
        cookieBanner @required(action: THROW) {
          id
          name
        }
        cookieBannerVersion @required(action: THROW) {
          id
          version
          categories {
            name
            slug
            description
            kind
            cookies {
              name
              trackerType
              maxAgeSeconds
              description
            }
          }
        }
        ipAddress
        userAgent
        sdkVersion
        regulation
        regulationSource
        countryCode
        consentData
        createdAt
      }
    }
  }
`;

interface CookieBannerConsentRecordPageProps {
  queryRef: PreloadedQuery<CookieBannerConsentRecordPageQuery>;
}

const PERSISTENT_TRACKER_TYPES = new Set([
  "LOCAL_STORAGE",
  "INDEXED_DB",
  "CACHE_STORAGE",
]);

export default function CookieBannerConsentRecordPage({
  queryRef,
}: CookieBannerConsentRecordPageProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<CookieBannerConsentRecordPageQuery>(cookieBannerConsentRecordPageQuery, queryRef);

  if (data.node.__typename !== "CookieConsentRecord") {
    throw new Error("invalid type for node");
  }

  const record = data.node;
  const bannerId = record.cookieBanner.id;
  const bannerName = record.cookieBanner.name;

  const consentMap = useMemo(() => {
    try {
      return JSON.parse(record.consentData) as Record<string, boolean>;
    } catch {
      return {};
    }
  }, [record.consentData]);

  const categories = record.cookieBannerVersion.categories;
  const formatDuration = (seconds: number | null, trackerType?: string | null) => {
    if (seconds === null || seconds <= 0) {
      return trackerType && PERSISTENT_TRACKER_TYPES.has(trackerType)
        ? t("duration.persistent")
        : t("duration.session");
    }
    return humanizeSeconds(seconds, t);
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("consentRecordPage.breadcrumbs.index"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: bannerName,
            to: `/organizations/${organizationId}/cookie-banners/${bannerId}`,
          },
          {
            label: t("consentRecordPage.breadcrumbs.records"),
            to: `/organizations/${organizationId}/cookie-banners/${bannerId}/consent-records`,
          },
          {
            label: record.id,
          },
        ]}
      />

      <PageHeader title={t("consentRecordPage.title")} />

      <Card padded>
        <PropertyRow label={t("consentRecordPage.properties.visitorId")}>
          <span className="font-mono text-sm">{record.visitorId}</span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.action")}>
          <Badge variant={getActionVariant(record.action)}>
            {t(`consentRecordPage.actions.${record.action.toLowerCase()}`)}
          </Badge>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.bannerVersion")}>
          {record.cookieBannerVersion
            ? (
                <span className="font-mono text-sm">
                  {record.cookieBannerVersion.version}
                </span>
              )
            : (
                <span className="text-txt-tertiary">-</span>
              )}
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.ipAddress")}>
          <span className="font-mono text-sm">
            {record.ipAddress
              ? formatAnonymizedIp(record.ipAddress)
              : "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.userAgent")}>
          <span className="font-mono text-sm break-all">
            {record.userAgent ?? "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.sdkVersion")}>
          <span className="font-mono text-sm">{record.sdkVersion}</span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.regulation")}>
          <span className="font-mono text-sm">
            {record.regulation || "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.regulationSource")}>
          <span className="font-mono text-sm">
            {record.regulationSource || "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.country")}>
          <span className="font-mono text-sm">
            {record.countryCode || "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={t("consentRecordPage.properties.date")}>
          <time dateTime={record.createdAt}>
            {dateFormat(i18n.language, record.createdAt)}
          </time>
        </PropertyRow>
      </Card>

      <Card padded>
        <h3 className="text-lg font-semibold mb-4">{t("consentRecordPage.consentData")}</h3>
        {categories.length > 0
          ? (
              <div className="space-y-4">
                {categories.map((category) => {
                  const consented = consentMap[category.slug];
                  return (
                    <div
                      key={category.slug}
                      className="border-b border-border-low pb-4 last:border-b-0 last:pb-0"
                    >
                      <div className="flex items-center justify-between mb-1">
                        <div>
                          <span className="font-medium">{category.name}</span>
                          {category.kind === "NECESSARY" && (
                            <span className="ml-2 text-xs text-txt-tertiary">
                              {t("consentRecordPage.required")}
                            </span>
                          )}
                        </div>
                        <Badge
                          variant={consented ? "success" : "danger"}
                          size="sm"
                        >
                          {consented ? t("consentRecordPage.consent.accepted") : t("consentRecordPage.consent.rejected")}
                        </Badge>
                      </div>
                      {category.description && (
                        <p className="text-sm text-txt-secondary mb-2">
                          {category.description}
                        </p>
                      )}
                      {category.cookies.length > 0 && (
                        <div className="ml-4 mt-2">
                          <table className="w-full text-sm">
                            <thead>
                              <tr className="text-left text-txt-tertiary">
                                <th className="font-medium pb-1 pr-4">
                                  {t("consentRecordPage.cookies.columns.cookie")}
                                </th>
                                <th className="font-medium pb-1 pr-4">
                                  {t("consentRecordPage.cookies.columns.duration")}
                                </th>
                                <th className="font-medium pb-1">
                                  {t("consentRecordPage.cookies.columns.description")}
                                </th>
                              </tr>
                            </thead>
                            <tbody>
                              {category.cookies.map(cookie => (
                                <tr key={cookie.name}>
                                  <td className="py-1 pr-4 font-mono text-xs">
                                    {cookie.name}
                                  </td>
                                  <td className="py-1 pr-4 text-txt-secondary">
                                    {formatDuration(cookie.maxAgeSeconds ?? null, cookie.trackerType)}
                                  </td>
                                  <td className="py-1 text-txt-secondary">
                                    {cookie.description}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )
          : (
              <p className="text-sm text-txt-tertiary font-mono">
                {record.consentData}
              </p>
            )}
      </Card>
    </div>
  );
}
