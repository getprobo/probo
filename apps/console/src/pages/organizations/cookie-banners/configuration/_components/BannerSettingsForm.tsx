// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, Input, Label, Option, Select, useToast } from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { BannerSettingsForm_cookieBanner$key } from "#/__generated__/core/BannerSettingsForm_cookieBanner.graphql";
import type { BannerSettingsFormMutation } from "#/__generated__/core/BannerSettingsFormMutation.graphql";

const bannerSettingsFormFragment = graphql`
  fragment BannerSettingsForm_cookieBanner on CookieBanner {
    id
    name
    origin
    privacyPolicyUrl
    consentExpiryDays
    consentMode
  }
`;

const updateBannerMutation = graphql`
  mutation BannerSettingsFormMutation($input: UpdateCookieBannerInput!) {
    updateCookieBanner(input: $input) {
      cookieBanner {
        id
        name
        origin
        privacyPolicyUrl
        consentExpiryDays
        consentMode
      }
    }
  }
`;

interface BannerSettingsFormProps {
  cookieBannerKey: BannerSettingsForm_cookieBanner$key;
}

export function BannerSettingsForm({ cookieBannerKey }: BannerSettingsFormProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const banner = useFragment(bannerSettingsFormFragment, cookieBannerKey);

  const [commitMutation, isInFlight] = useMutation<BannerSettingsFormMutation>(updateBannerMutation);

  const [name, setName] = useState(banner.name);
  const [origin, setOrigin] = useState(banner.origin);
  const [privacyPolicyUrl, setPrivacyPolicyUrl] = useState(banner.privacyPolicyUrl);
  const [consentExpiryDays, setConsentExpiryDays] = useState(String(banner.consentExpiryDays));
  const [consentMode, setConsentMode] = useState(banner.consentMode);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    commitMutation({
      variables: {
        input: {
          cookieBannerId: banner.id,
          name,
          origin,
          privacyPolicyUrl,
          consentExpiryDays: parseInt(consentExpiryDays, 10),
          consentMode: consentMode,
        },
      },
      onCompleted() {
        toast({ title: __("Success"), description: __("Banner settings updated"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to update"), error as GraphQLError), variant: "error" });
      },
    });
  };

  return (
    <div className="space-y-4">
      <h3 className="font-medium">{__("Settings")}</h3>
      <Card className="border p-4">
        <form className="space-y-4" onSubmit={handleSubmit}>
          <Field label={__("Name")}>
            <Input value={name} onChange={e => setName(e.target.value)} required />
          </Field>

          <Field label={__("Origin URL")}>
            <Input value={origin} onChange={e => setOrigin(e.target.value)} required />
          </Field>

          <Field label={__("Privacy Policy URL")}>
            <Input value={privacyPolicyUrl} onChange={e => setPrivacyPolicyUrl(e.target.value)} required />
          </Field>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{__("Consent Expiry (days)")}</Label>
              <Input
                type="number"
                value={consentExpiryDays}
                onChange={e => setConsentExpiryDays(e.target.value)}
                min="1"
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{__("Consent Mode")}</Label>
              <Select value={consentMode} onValueChange={v => setConsentMode(v as "OPT_IN" | "OPT_OUT")}>
                <Option value="OPT_IN">{__("Opt-in")}</Option>
                <Option value="OPT_OUT">{__("Opt-out")}</Option>
              </Select>
            </div>
          </div>

          <Button type="submit" disabled={isInFlight}>
            {isInFlight ? __("Saving...") : __("Save")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
