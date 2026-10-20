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
import { useState } from "react";
import { useTranslation } from "react-i18next";

import type { OAuthConnectForm_provider$key } from "#/__generated__/core/OAuthConnectForm_provider.graphql";

import { OAuthConnectForm } from "./OAuthConnectForm";

// Accepts a bare subdomain ("acme") or a pasted host
// ("https://acme.zendesk.com/") and returns the bare subdomain.
function cleanZendeskSubdomain(raw: string): string {
  let value = raw.trim();
  value = value.replace(/^https?:\/\//i, "");
  value = value.replace(/[/?#].*$/, "");
  value = value.replace(/\.zendesk\.com$/i, "");
  return value.trim();
}

export function ZendeskConnectForm({
  providerKey,
}: {
  providerKey: OAuthConnectForm_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [subdomain, setSubdomain] = useState("");
  const site = cleanZendeskSubdomain(subdomain);

  return (
    <OAuthConnectForm
      providerKey={providerKey}
      extras={{ site }}
      canSubmit={site !== ""}
    >
      <Field required label={t("marketplacePage.fields.zendeskSubdomain")}>
        <TextField
          name="subdomain"
          required
          value={subdomain}
          onValueChange={setSubdomain}
        />
      </Field>
    </OAuthConnectForm>
  );
}
