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

import {
  certifications,
  objectEntries,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  Badge,
  Button,
  Card,
  Combobox,
  ComboboxItem,
  IconCrossLargeX,
  IconPlusLarge,
  Input,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { clsx } from "clsx";
import type { ComponentProps, FocusEvent } from "react";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyAssurancePageFragment$key } from "#/__generated__/core/ThirdPartyAssurancePageFragment.graphql";
import type { ThirdPartyAssurancePageQuery } from "#/__generated__/core/ThirdPartyAssurancePageQuery.graphql";
import type { ThirdPartyAssurancePageRefetchQuery } from "#/__generated__/core/ThirdPartyAssurancePageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { UploadComplianceReportDialog } from "../_components/UploadComplianceReportDialog";
import { useUpdateThirdParty } from "../_lib/useUpdateThirdParty";

import { ThirdPartyComplianceReportRow } from "./_components/ThirdPartyComplianceReportRow";

const complianceReportsFragment = graphql`
  fragment ThirdPartyAssurancePageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyAssurancePageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyComplianceReportOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    canUploadComplianceReport: permission(
      action: "core:thirdParty-compliance-report:upload"
    )
    complianceReports(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyAssurancePage_complianceReports") {
      __id
      edges {
        node {
          id
          ...ThirdPartyComplianceReportRow_report
        }
      }
    }
  }
`;

export const thirdPartyAssurancePageQuery = graphql`
  query ThirdPartyAssurancePageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        name
        canUpdate: permission(action: "core:thirdParty:update")
        certifications
        statusPageUrl
        termsOfServiceUrl
        privacyPolicyUrl
        serviceLevelAgreementUrl
        dataProcessingAgreementUrl
        securityPageUrl
        trustPageUrl
        ...ThirdPartyAssurancePageFragment
      }
    }
  }
`;

interface ThirdPartyAssurancePageProps {
  queryRef: PreloadedQuery<ThirdPartyAssurancePageQuery>;
}

type UrlFieldName
  = | "statusPageUrl"
    | "termsOfServiceUrl"
    | "privacyPolicyUrl"
    | "serviceLevelAgreementUrl"
    | "dataProcessingAgreementUrl"
    | "securityPageUrl"
    | "trustPageUrl";

function normalizeUrl(value: string | null | undefined): string | null {
  const trimmed = value?.trim() ?? "";
  return trimmed.length > 0 ? trimmed : null;
}

export function ThirdPartyAssurancePage({ queryRef }: ThirdPartyAssurancePageProps) {
  const data = usePreloadedQuery<ThirdPartyAssurancePageQuery>(
    thirdPartyAssurancePageQuery,
    queryRef,
  );
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;
  const { t } = useTranslation();
  const [update] = useUpdateThirdParty();
  const [pendingCertifications, setPendingCertifications] = useState<
    readonly string[] | null
  >(null);
  const [pendingUrls, setPendingUrls] = useState<Partial<Record<UrlFieldName, string | null>>>({});
  const urlSavesRef = useRef<Partial<Record<UrlFieldName, Promise<void>>>>({});
  const certificationsValue = pendingCertifications ?? thirdParty.certifications;

  const { data: reportsData, refetch, ...pagination } = usePaginationFragment<
    ThirdPartyAssurancePageRefetchQuery,
    ThirdPartyAssurancePageFragment$key
  >(complianceReportsFragment, thirdParty);

  const connectionId = reportsData.complianceReports.__id;
  const reports = reportsData.complianceReports.edges.map(edge => edge.node);

  const urls = [
    {
      name: "statusPageUrl",
      label: t("thirdPartyAssurancePage.urlLabels.statusPage"),
      value: thirdParty.statusPageUrl,
    },
    {
      name: "termsOfServiceUrl",
      label: t("thirdPartyAssurancePage.urlLabels.termsOfService"),
      value: thirdParty.termsOfServiceUrl,
    },
    {
      name: "privacyPolicyUrl",
      label: t("thirdPartyAssurancePage.urlLabels.privacyPolicy"),
      value: thirdParty.privacyPolicyUrl,
    },
    {
      name: "serviceLevelAgreementUrl",
      label: t("thirdPartyAssurancePage.urlLabels.serviceLevelAgreement"),
      value: thirdParty.serviceLevelAgreementUrl,
    },
    {
      name: "dataProcessingAgreementUrl",
      label: t("thirdPartyAssurancePage.urlLabels.dataProcessingAgreement"),
      value: thirdParty.dataProcessingAgreementUrl,
    },
    {
      name: "securityPageUrl",
      label: t("thirdPartyAssurancePage.urlLabels.securityPage"),
      value: thirdParty.securityPageUrl,
    },
    {
      name: "trustPageUrl",
      label: t("thirdPartyAssurancePage.urlLabels.trustPage"),
      value: thirdParty.trustPageUrl,
    },
  ] as const satisfies ReadonlyArray<{
    name: UrlFieldName;
    label: string;
    value: string | null | undefined;
  }>;

  usePageTitle(t("thirdPartyAssurancePage.pageTitle", { name: thirdParty.name }));

  function handleCertificationsChange(next: string[]) {
    setPendingCertifications(next);
    void update(thirdParty.id, "certifications", next)
      .then(() => {
        setPendingCertifications(null);
      })
      .catch(() => {
        setPendingCertifications(null);
      });
  }

  function handleUrlBlur(
    field: UrlFieldName,
    current: string | null | undefined,
    event: FocusEvent<HTMLInputElement>,
  ) {
    const next = normalizeUrl(event.target.value);
    if (next === normalizeUrl(current)) {
      return;
    }
    setPendingUrls(prev => ({ ...prev, [field]: next }));
    const previous = urlSavesRef.current[field] ?? Promise.resolve();
    const save = previous
      .catch(() => undefined)
      .then(() => update(thirdParty.id, field, next))
      .then(() => {
        setPendingUrls((prev) => {
          if (prev[field] !== next) {
            return prev;
          }
          const { [field]: _cleared, ...rest } = prev;
          return rest;
        });
      })
      .catch(() => {
        setPendingUrls((prev) => {
          if (prev[field] !== next) {
            return prev;
          }
          const { [field]: _cleared, ...rest } = prev;
          return rest;
        });
      });
    urlSavesRef.current[field] = save;
  }

  return (
    <div className="space-y-12">
      <PageHeader
        title={t("thirdPartyAssurancePage.title")}
        description={t("thirdPartyAssurancePage.description")}
      />

      <div className="space-y-4">
        <h2 className="text-base font-medium">
          {t("thirdPartyAssurancePage.sections.certifications")}
        </h2>
        <Card padded>
          <Certifications
            onValueChange={handleCertificationsChange}
            value={certificationsValue}
            readOnly={!thirdParty.canUpdate || pendingCertifications !== null}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("thirdPartyAssurancePage.sections.links")}</h2>
        <Card className="divide-y divide-border-low">
          {urls.map((url) => {
            const savedValue = pendingUrls[url.name] !== undefined
              ? (normalizeUrl(pendingUrls[url.name]) ?? "")
              : (normalizeUrl(url.value) ?? "");
            return (
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
                  key={savedValue}
                  className="p-4 focus:bg-tertiary-pressed outline-none"
                  id={url.name}
                  defaultValue={savedValue}
                  onBlur={event => handleUrlBlur(url.name, savedValue, event)}
                  type="text"
                  placeholder="https://..."
                  variant="ghost"
                  disabled={!thirdParty.canUpdate}
                />
              </div>
            );
          })}
        </Card>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between gap-4">
          <h2 className="text-base font-medium">
            {t("thirdPartyAssurancePage.sections.reports")}
          </h2>
          {reportsData.canUploadComplianceReport && (
            <UploadComplianceReportDialog
              thirdPartyId={reportsData.id}
              connectionId={connectionId}
            >
              <Button icon={IconPlusLarge}>{t("thirdPartyAssurancePage.actions.addReport")}</Button>
            </UploadComplianceReportDialog>
          )}
        </div>

        <SortableTable
          {...pagination}
          refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
        >
          <Thead>
            <Tr>
              <Th>{t("thirdPartyAssurancePage.columns.reportName")}</Th>
              <SortableTh field="REPORT_DATE">{t("thirdPartyAssurancePage.columns.reportDate")}</SortableTh>
              <Th>{t("thirdPartyAssurancePage.columns.validUntil")}</Th>
              <Th>{t("thirdPartyAssurancePage.columns.fileSize")}</Th>
              {reports.length > 0 && <Th>{t("thirdPartyAssurancePage.columns.actions")}</Th>}
            </Tr>
          </Thead>
          <Tbody>
            {reports.map(report => (
              <ThirdPartyComplianceReportRow
                key={report.id}
                reportKey={report}
                connectionId={connectionId}
              />
            ))}
          </Tbody>
        </SortableTable>
      </div>
    </div>
  );
}

interface CertificationsProps {
  value: readonly string[];
  onValueChange: (value: string[]) => void;
  readOnly?: boolean;
}

function Certifications(props: CertificationsProps) {
  const categorizedCertifications = Object.values(certifications).flat();
  const { t } = useTranslation();
  // Only the badge that was just added animates in; `starting:` styles are
  // entry-only, so a shared flag would tag every later-mounted badge.
  const [addedCertification, setAddedCertification] = useState<string | null>(null);
  const categories = objectEntries(certifications)
    .map(
      ([key, value]) =>
        [key, value.filter(c => props.value.includes(c))] as const,
    )
    .filter(([, certs]) => certs.length > 0);
  const customCertifications = props.value.filter(
    c => !categorizedCertifications.includes(c),
  );
  if (customCertifications.length > 0) {
    categories.push(["custom", customCertifications]);
  }

  const addCertificate = (name: string) => {
    if (props.value.includes(name)) {
      return;
    }
    setAddedCertification(name);
    props.onValueChange([...props.value, name]);
  };

  const removeCertificate = (name: string) => {
    setAddedCertification(current => (current === name ? null : current));
    props.onValueChange(props.value.filter(v => v !== name));
  };

  return (
    <div className="space-y-6">
      {categories.map(([key, certs]) => (
        <div key={key} className="space-y-2">
          <div className="text-sm font-medium text-txt-secondary">
            {t(`thirdPartyAssurancePage.categories.${key}`)}
          </div>
          <div className="flex flex-wrap gap-2">
            {certs.map(certification => (
              <Badge asChild size="md" key={certification}>
                {props.readOnly
                  ? (
                      <span>{certification}</span>
                    )
                  : (
                      <button
                        onClick={() => removeCertificate(certification)}
                        type="button"
                        className={clsx(
                          "hover:bg-subtle-hover cursor-pointer",
                          addedCertification === certification
                          && "starting:opacity-0 starting:w-0 w-max transition-all duration-500 starting:bg-accent",
                        )}
                      >
                        {certification}
                        <div className="w-0 overflow-hidden group-hover:w-4 duration-200">
                          <IconCrossLargeX size={12} />
                        </div>
                      </button>
                    )}
              </Badge>
            ))}
          </div>
        </div>
      ))}
      {!props.readOnly && (
        <CertificationInput
          certifications={categorizedCertifications.filter(
            c => !props.value.includes(c),
          )}
          onAdd={addCertificate}
        />
      )}
    </div>
  );
}

function CertificationInput({
  certifications,
  onAdd,
}: {
  certifications: string[];
  onAdd: (name: string) => void;
}) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const isCustom = !certifications.includes(search.trim());
  const filteredCertifications = certifications.filter(c =>
    c.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="flex items-center gap-2">
      <Combobox
        autoSelect
        resetValueOnHide
        onSelect={onAdd}
        onSearch={setSearch}
        placeholder={t("thirdPartyAssurancePage.placeholders.add")}
      >
        {filteredCertifications.map(certification => (
          <ComboboxItem key={certification} value={certification}>
            {certification}
          </ComboboxItem>
        ))}
        {isCustom && search.trim().length >= 2 && (
          <ComboboxItem value={search.trim()}>
            <IconPlusLarge size={20} />
            {t("thirdPartyAssurancePage.addCustom", { name: search })}
          </ComboboxItem>
        )}
      </Combobox>
    </div>
  );
}
