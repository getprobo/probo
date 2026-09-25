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
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import {
  cleanZendeskSubdomain,
  connectOAuthProvider,
  DATADOG_SITES,
} from "#/pages/organizations/access-reviews/dialogs/_lib/connectorSettings";

import { ConnectFormFooter, type ConnectVendorDriver } from "./ConnectFormFooter";

export function OAuthConnectForm({
  organizationId,
  driver,
}: {
  organizationId: string;
  driver: ConnectVendorDriver;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [datadogSite, setDatadogSite] = useState("US1");
  const [zendeskSubdomain, setZendeskSubdomain] = useState("");
  const zendeskReady = driver.provider !== "ZENDESK" || cleanZendeskSubdomain(zendeskSubdomain) !== "";

  return (
    <form
      className="flex flex-col gap-4"
      onSubmit={(event) => {
        event.preventDefault();
        if (!zendeskReady) {
          return;
        }
        const extras: Record<string, string> | undefined = driver.provider === "DATADOG"
          ? { site: datadogSite }
          : driver.provider === "ZENDESK"
            ? { subdomain: cleanZendeskSubdomain(zendeskSubdomain) }
            : undefined;
        connectOAuthProvider(organizationId, driver.provider, driver.oauth2Scopes, extras);
      }}
    >
      {driver.provider === "DATADOG" && (
        <Field label={t("marketplacePage.fields.datadogSite")} required>
          <Select value={datadogSite} onValueChange={(value: string | null) => setDatadogSite(value ?? "US1")}>
            <SelectTrigger>
              {(value: string | null) => DATADOG_SITES.find(site => site.value === value)?.label ?? null}
            </SelectTrigger>
            <SelectPopup>
              {DATADOG_SITES.map(site => (
                <SelectItem key={site.value} value={site.value}>{site.label}</SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </Field>
      )}
      {driver.provider === "ZENDESK" && (
        <Field label={t("marketplacePage.fields.zendeskSubdomain")} required>
          <TextField
            value={zendeskSubdomain}
            onChange={event => setZendeskSubdomain(event.target.value)}
          />
        </Field>
      )}
      <ConnectFormFooter
        documentationUrl={driver.documentationUrl}
        disabled={!zendeskReady}
      />
    </form>
  );
}
