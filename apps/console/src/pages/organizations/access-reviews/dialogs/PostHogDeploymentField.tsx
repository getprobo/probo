// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Field, Option, Select } from "@probo/ui";

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
  const { __ } = useTranslate();

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
        <label className="text-sm font-medium">{__("Deployment")}</label>
        <Select
          value={deployment}
          onValueChange={(val: string) =>
            onChange(val === "SELF_HOSTED" ? { instanceUrl: "" } : { region: val })}
          placeholder={__("Select a deployment")}
        >
          <Option value="US">{__("PostHog Cloud (US)")}</Option>
          <Option value="EU">{__("PostHog Cloud (EU)")}</Option>
          <Option value="SELF_HOSTED">{__("Self-hosted")}</Option>
        </Select>
      </div>
      {deployment === "SELF_HOSTED" && (
        <Field
          label={__("Instance URL")}
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
