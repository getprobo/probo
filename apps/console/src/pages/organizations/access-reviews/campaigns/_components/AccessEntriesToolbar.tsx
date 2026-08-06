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

import { IconMagnifyingGlass, Input } from "@probo/ui";
import { useTranslation } from "react-i18next";

import { FilterMultiSelect, type FilterMultiSelectOption } from "./FilterMultiSelect";

type Props = {
  emailFilter: string;
  onEmailFilterChange: (value: string) => void;
  connectorOptions: ReadonlyArray<FilterMultiSelectOption>;
  connectorFilter: ReadonlyArray<string>;
  onConnectorFilterChange: (value: string[]) => void;
  mfaFilter: ReadonlyArray<string>;
  onMfaFilterChange: (value: string[]) => void;
  authMethodFilter: ReadonlyArray<string>;
  onAuthMethodFilterChange: (value: string[]) => void;
  adminFilter: ReadonlyArray<string>;
  onAdminFilterChange: (value: string[]) => void;
  activeFilter: ReadonlyArray<string>;
  onActiveFilterChange: (value: string[]) => void;
};

// Filter row styled like the compliance-portal subprocessors toolbar:
// compact selects on the left, expanding search field on the right.
export function AccessEntriesToolbar({
  emailFilter,
  onEmailFilterChange,
  connectorOptions,
  connectorFilter,
  onConnectorFilterChange,
  mfaFilter,
  onMfaFilterChange,
  authMethodFilter,
  onAuthMethodFilterChange,
  adminFilter,
  onAdminFilterChange,
  activeFilter,
  onActiveFilterChange,
}: Props) {
  const { t } = useTranslation();

  return (
    <div className="flex min-h-10 flex-wrap items-center gap-3">
      <div className="w-44 max-sm:w-full">
        <FilterMultiSelect
          placeholder={t("campaignDetailPage.filters.allConnectors")}
          options={connectorOptions}
          value={connectorFilter}
          onChange={onConnectorFilterChange}
        />
      </div>
      <div className="w-40 max-sm:w-full">
        <FilterMultiSelect
          placeholder={t("campaignDetailPage.filters.allMfa")}
          options={[
            { value: "ENABLED", label: t("campaignDetailPage.mfaStatus.enabled") },
            { value: "DISABLED", label: t("campaignDetailPage.mfaStatus.disabled") },
            { value: "UNKNOWN", label: t("campaignDetailPage.mfaStatus.unknown") },
          ]}
          value={mfaFilter}
          onChange={onMfaFilterChange}
        />
      </div>
      <div className="w-44 max-sm:w-full">
        <FilterMultiSelect
          placeholder={t("campaignDetailPage.filters.allAuthMethods")}
          options={[
            { value: "SSO", label: t("campaignDetailPage.authMethod.sso") },
            { value: "PASSWORD", label: t("campaignDetailPage.authMethod.password") },
            { value: "API_KEY", label: t("campaignDetailPage.authMethod.api_key") },
            {
              value: "SERVICE_ACCOUNT",
              label: t("campaignDetailPage.authMethod.service_account"),
            },
            { value: "UNKNOWN", label: t("campaignDetailPage.authMethod.unknown") },
          ]}
          value={authMethodFilter}
          onChange={onAuthMethodFilterChange}
        />
      </div>
      <div className="w-36 max-sm:w-full">
        <FilterMultiSelect
          placeholder={t("campaignDetailPage.filters.allAdmin")}
          options={[
            { value: "YES", label: t("campaignDetailPage.values.yes") },
            { value: "NO", label: t("campaignDetailPage.values.no") },
          ]}
          value={adminFilter}
          onChange={onAdminFilterChange}
        />
      </div>
      <div className="w-40 max-sm:w-full">
        <FilterMultiSelect
          placeholder={t("campaignDetailPage.filters.allAccountStatuses")}
          options={[
            { value: "ACTIVE", label: t("campaignDetailPage.accountStatus.active") },
            { value: "DISABLED", label: t("campaignDetailPage.accountStatus.disabled") },
            { value: "UNKNOWN", label: t("campaignDetailPage.accountStatus.unknown") },
          ]}
          value={activeFilter}
          onChange={onActiveFilterChange}
        />
      </div>
      <div className="min-w-60 flex-1 max-sm:w-full max-sm:min-w-0">
        <Input
          icon={IconMagnifyingGlass}
          value={emailFilter}
          onValueChange={onEmailFilterChange}
          placeholder={t("campaignDetailPage.filters.searchEmail")}
        />
      </div>
    </div>
  );
}
