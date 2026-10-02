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
import { useState } from "react";
import { useTranslation } from "react-i18next";

import type { OAuthConnectForm_provider$key } from "#/__generated__/core/OAuthConnectForm_provider.graphql";

import { DATADOG_SITES } from "../_lib/connectorSettings";

import { OAuthConnectForm } from "./OAuthConnectForm";

export function DatadogConnectForm({
  organizationId,
  providerKey,
}: {
  organizationId: string;
  providerKey: OAuthConnectForm_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [site, setSite] = useState("US1");

  return (
    <OAuthConnectForm
      organizationId={organizationId}
      providerKey={providerKey}
      extras={{ site }}
    >
      <Field label={t("marketplacePage.fields.datadogSite")} required>
        <Select value={site} onValueChange={(value: string | null) => setSite(value ?? "US1")}>
          <SelectTrigger>
            {(value: string | null) => DATADOG_SITES.find(entry => entry.value === value)?.label ?? null}
          </SelectTrigger>
          <SelectPopup>
            {DATADOG_SITES.map(entry => (
              <SelectItem key={entry.value} value={entry.value}>{entry.label}</SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </Field>
    </OAuthConnectForm>
  );
}
