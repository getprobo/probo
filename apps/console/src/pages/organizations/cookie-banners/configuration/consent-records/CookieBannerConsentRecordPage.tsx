// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Breadcrumb, Card, PageHeader, PropertyRow } from "@probo/ui";
import { useMemo } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CookieBannerConsentRecordPageQuery } from "#/__generated__/core/CookieBannerConsentRecordPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  formatAnonymizedIp,
  getActionLabel,
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
              duration
              description
            }
          }
        }
        ipAddress
        userAgent
        sdkVersion
        consentData
        createdAt
      }
    }
  }
`;

interface CookieBannerConsentRecordPageProps {
  queryRef: PreloadedQuery<CookieBannerConsentRecordPageQuery>;
}

export default function CookieBannerConsentRecordPage({
  queryRef,
}: CookieBannerConsentRecordPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(cookieBannerConsentRecordPageQuery, queryRef);

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

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Cookie Banners"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: bannerName,
            to: `/organizations/${organizationId}/cookie-banners/${bannerId}`,
          },
          {
            label: __("Consent Records"),
            to: `/organizations/${organizationId}/cookie-banners/${bannerId}/consent-records`,
          },
          {
            label: record.id,
          },
        ]}
      />

      <PageHeader title={__("Consent Record")} />

      <Card padded>
        <PropertyRow label={__("Visitor ID")}>
          <span className="font-mono text-sm">{record.visitorId}</span>
        </PropertyRow>
        <PropertyRow label={__("Action")}>
          <Badge variant={getActionVariant(record.action)}>
            {getActionLabel(record.action, __)}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Banner Version")}>
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
        <PropertyRow label={__("IP Address")}>
          <span className="font-mono text-sm">
            {record.ipAddress
              ? formatAnonymizedIp(record.ipAddress)
              : "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={__("User Agent")}>
          <span className="font-mono text-sm truncate max-w-md" title={record.userAgent ?? undefined}>
            {record.userAgent ?? "-"}
          </span>
        </PropertyRow>
        <PropertyRow label={__("SDK Version")}>
          <span className="font-mono text-sm">{record.sdkVersion}</span>
        </PropertyRow>
        <PropertyRow label={__("Date")}>
          <time dateTime={record.createdAt}>
            {formatDate(record.createdAt)}
          </time>
        </PropertyRow>
      </Card>

      <Card padded>
        <h3 className="text-lg font-semibold mb-4">{__("Consent Data")}</h3>
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
                              (
                              {__("Required")}
                              )
                            </span>
                          )}
                        </div>
                        <Badge
                          variant={consented ? "success" : "danger"}
                          size="sm"
                        >
                          {consented ? __("Accepted") : __("Rejected")}
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
                                  {__("Cookie")}
                                </th>
                                <th className="font-medium pb-1 pr-4">
                                  {__("Duration")}
                                </th>
                                <th className="font-medium pb-1">
                                  {__("Description")}
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
                                    {cookie.duration}
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
