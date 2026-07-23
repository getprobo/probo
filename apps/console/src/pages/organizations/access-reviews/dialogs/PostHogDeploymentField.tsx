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

import { Field, Option, Select } from "@probo/ui";
import { useTranslation } from "react-i18next";

// PostHog is a single provider spanning Cloud (region us/eu) and self-hosted
// (instance URL). The API-key form surfaces this as one deployment choice;
// the two settings are mutually exclusive, so picking one clears the other.

type PostHogDeploymentFieldProps = {
  values: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
};

export function PostHogDeploymentField({
  values,
  onChange,
}: PostHogDeploymentFieldProps) {
  const { t } = useTranslation();

  const region = values.region ?? "";
  let deployment = "";
  if (region === "US" || region === "EU") {
    deployment = region;
  } else if ("instanceUrl" in values) {
    deployment = "SELF_HOSTED";
  }

  return (
    <>
      <div className="space-y-1.5">
        <label className="text-sm font-medium">
          {t("postHogDeploymentField.deployment.label")}
        </label>
        <Select
          value={deployment}
          onValueChange={(val: string) =>
            onChange(val === "SELF_HOSTED" ? { instanceUrl: "" } : { region: val })}
          placeholder={t("postHogDeploymentField.deployment.placeholder")}
        >
          <Option value="US">
            {t("postHogDeploymentField.deployment.usCloud")}
          </Option>
          <Option value="EU">
            {t("postHogDeploymentField.deployment.euCloud")}
          </Option>
          <Option value="SELF_HOSTED">
            {t("postHogDeploymentField.deployment.selfHosted")}
          </Option>
        </Select>
      </div>
      {deployment === "SELF_HOSTED" && (
        <Field
          label={t("postHogDeploymentField.instanceUrl")}
          value={values.instanceUrl ?? ""}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            onChange({ instanceUrl: e.target.value })}
          required
        />
      )}
    </>
  );
}

// isPostHogDeploymentSelected reports whether a valid PostHog deployment has
// been chosen: a Cloud region (us/eu) or a non-empty self-hosted instance URL.
export function isPostHogDeploymentSelected(
  values: Record<string, string>,
): boolean {
  return (
    values.region === "US"
    || values.region === "EU"
    || !!values.instanceUrl?.trim()
  );
}
