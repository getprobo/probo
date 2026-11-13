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

import { Field } from "@probo/ui/src/v2/form/Field";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

// Labels are technical identifiers (region code + hostname).
const DATADOG_SITES: { value: string; label: string }[] = [
  { value: "US1", label: "US1 (app.datadoghq.com)" },
  { value: "US3", label: "US3 (us3.datadoghq.com)" },
  { value: "US5", label: "US5 (us5.datadoghq.com)" },
  { value: "EU1", label: "EU1 (app.datadoghq.eu)" },
  { value: "AP1", label: "AP1 (ap1.datadoghq.com)" },
  { value: "AP2", label: "AP2 (ap2.datadoghq.com)" },
  { value: "UK1", label: "UK1 (uk1.datadoghq.com)" },
  { value: "US1-FED", label: "US1-FED (app.ddog-gov.com)" },
  { value: "US2-FED", label: "US2-FED (us2.ddog-gov.com)" },
];

interface DatadogSiteFieldProps {
  value: string | null;
  onValueChange: (value: string) => void;
}

export function DatadogSiteField({ value, onValueChange }: DatadogSiteFieldProps) {
  const { t } = useTranslation("organizations/settings/integrations");

  return (
    <Field required label={t("marketplacePage.fields.datadogSite")}>
      <Select value={value} onValueChange={(next: string | null) => onValueChange(next ?? "")}>
        <SelectTrigger placeholder={t("marketplacePage.fields.datadogSitePlaceholder")}>
          {(selected: string | null) => DATADOG_SITES.find(entry => entry.value === selected)?.label ?? null}
        </SelectTrigger>
        <SelectPopup>
          {DATADOG_SITES.map(entry => (
            <SelectItem key={entry.value} value={entry.value}>{entry.label}</SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </Field>
  );
}
