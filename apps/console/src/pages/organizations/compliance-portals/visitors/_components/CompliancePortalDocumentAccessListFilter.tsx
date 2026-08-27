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

import { getCompliancePortalDocumentAccessStatusLabel } from "@probo/helpers";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

import {
  type DocumentAccessListStatusOption,
  documentAccessListStatusOptions,
  isDocumentAccessListStatusOption,
  useDocumentAccessListFilters,
} from "../_lib/useDocumentAccessListFilters";
import { documentAccessList } from "../variants";

export function CompliancePortalDocumentAccessListFilter() {
  const { t } = useTranslation("organizations/compliance-portals");
  const { t: tRoot } = useTranslation();
  const { status, setStatus } = useDocumentAccessListFilters();
  const { filter } = documentAccessList();
  const allLabel = t("documentAccessList.filter.all");

  function optionLabel(option: DocumentAccessListStatusOption): string {
    switch (option) {
      case "none":
        return t("documentAccessList.filter.none");
      case "requested":
        return getCompliancePortalDocumentAccessStatusLabel("REQUESTED", tRoot);
      case "granted":
        return getCompliancePortalDocumentAccessStatusLabel("GRANTED", tRoot);
      case "revoked":
        return getCompliancePortalDocumentAccessStatusLabel("REVOKED", tRoot);
      case "rejected":
        return getCompliancePortalDocumentAccessStatusLabel("REJECTED", tRoot);
    }
  }

  return (
    <div className={filter()}>
      <Select
        value={status === "all" ? null : status}
        onValueChange={(value: string | null) => {
          if (value == null) {
            setStatus("all");
            return;
          }
          if (isDocumentAccessListStatusOption(value)) {
            setStatus(value);
          }
        }}
      >
        <SelectTrigger
          size={2}
          placeholder={allLabel}
          aria-label={t("documentAccessList.filter.label")}
        >
          {(value: DocumentAccessListStatusOption | null) => (
            value != null ? optionLabel(value) : allLabel
          )}
        </SelectTrigger>
        <SelectPopup align="start">
          <SelectItem value={null}>{allLabel}</SelectItem>
          {documentAccessListStatusOptions.map(option => (
            <SelectItem key={option} value={option}>
              {optionLabel(option)}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
}
