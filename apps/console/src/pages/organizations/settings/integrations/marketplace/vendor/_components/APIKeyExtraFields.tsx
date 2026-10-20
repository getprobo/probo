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
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { APIKeyExtraFields_provider$key } from "#/__generated__/core/APIKeyExtraFields_provider.graphql";

import { PostHogDeploymentField } from "./PostHogDeploymentField";

const apiKeyExtraFieldsFragment = graphql`
  fragment APIKeyExtraFields_provider on ConnectorProviderInfo {
    provider
    apiKeyExtraSettings {
      key
      label
      required
    }
  }
`;

interface APIKeyExtraFieldsProps {
  providerKey: APIKeyExtraFields_provider$key;
  extras: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
}

export function APIKeyExtraFields({
  providerKey,
  extras,
  onChange,
}: APIKeyExtraFieldsProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const provider = useFragment(apiKeyExtraFieldsFragment, providerKey);
  const extraSettings = provider.apiKeyExtraSettings;

  if (provider.provider === "POSTHOG") {
    return <PostHogDeploymentField values={extras} onChange={onChange} />;
  }

  if (provider.provider === "SEGMENT") {
    return (
      <Field required label={t("marketplacePage.fields.region")}>
        <Select
          value={extras.region ?? null}
          onValueChange={(value: string | null) => {
            onChange({ ...extras, region: value ?? "" });
          }}
        >
          <SelectTrigger placeholder={t("marketplacePage.fields.regionPlaceholder")}>
            {(value: string | null) => {
              if (value === "US") {
                return t("marketplacePage.fields.regionUnitedStates");
              }
              if (value === "EU") {
                return t("marketplacePage.fields.regionEurope");
              }
              return null;
            }}
          </SelectTrigger>
          <SelectPopup>
            <SelectItem value="US">{t("marketplacePage.fields.regionUnitedStates")}</SelectItem>
            <SelectItem value="EU">{t("marketplacePage.fields.regionEurope")}</SelectItem>
          </SelectPopup>
        </Select>
      </Field>
    );
  }

  return extraSettings.map(setting => (
    <Field key={setting.key} required={setting.required} label={setting.label}>
      <TextField
        name={setting.key}
        required={setting.required}
        value={extras[setting.key] ?? ""}
        onValueChange={value => onChange({ ...extras, [setting.key]: value })}
      />
    </Field>
  ));
}
