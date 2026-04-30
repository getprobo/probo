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
import { Controller, useForm } from "react-hook-form";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { BannerSettingsForm_cookieBanner$key } from "#/__generated__/core/BannerSettingsForm_cookieBanner.graphql";
import type { BannerSettingsFormMutation } from "#/__generated__/core/BannerSettingsFormMutation.graphql";

const bannerSettingsFormFragment = graphql`
  fragment BannerSettingsForm_cookieBanner on CookieBanner {
    id
    name
    origin
    cookiePolicyUrl
    privacyPolicyUrl
    consentExpiryDays
    consentMode
    defaultLanguage
  }
`;

const updateBannerMutation = graphql`
  mutation BannerSettingsFormMutation($input: UpdateCookieBannerInput!) {
    updateCookieBanner(input: $input) {
      cookieBanner {
        id
        name
        cookiePolicyUrl
        privacyPolicyUrl
        consentExpiryDays
        consentMode
        defaultLanguage
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface BannerSettingsFormValues {
  name: string;
  cookiePolicyUrl: string;
  privacyPolicyUrl: string;
  consentExpiryDays: string;
  consentMode: "OPT_IN" | "OPT_OUT";
  defaultLanguage: string;
}

interface BannerSettingsFormProps {
  cookieBannerKey: BannerSettingsForm_cookieBanner$key;
}

export function BannerSettingsForm({ cookieBannerKey }: BannerSettingsFormProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const banner = useFragment(bannerSettingsFormFragment, cookieBannerKey);

  const [updateBanner, isUpdating] = useMutation<BannerSettingsFormMutation>(updateBannerMutation);

  const { register, handleSubmit, control } = useForm<BannerSettingsFormValues>({
    defaultValues: {
      name: banner.name,
      cookiePolicyUrl: banner.cookiePolicyUrl,
      privacyPolicyUrl: banner.privacyPolicyUrl ?? "",
      consentExpiryDays: String(banner.consentExpiryDays),
      consentMode: banner.consentMode,
      defaultLanguage: banner.defaultLanguage,
    },
  });

  const onSubmit = (data: BannerSettingsFormValues) => {
    updateBanner({
      variables: {
        input: {
          cookieBannerId: banner.id,
          name: data.name,
          cookiePolicyUrl: data.cookiePolicyUrl,
          privacyPolicyUrl: data.privacyPolicyUrl || undefined,
          consentExpiryDays: parseInt(data.consentExpiryDays, 10),
          consentMode: data.consentMode,
          defaultLanguage: data.defaultLanguage,
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
        <form className="space-y-4" onSubmit={e => void handleSubmit(onSubmit)(e)}>
          <Field label={__("Name")}>
            <Input {...register("name")} required />
          </Field>

          <Field label={__("Origin URL")}>
            <Input value={banner.origin} disabled />
          </Field>

          <Field label={__("Cookie Policy URL")}>
            <Input {...register("cookiePolicyUrl")} required />
          </Field>

          <Field label={__("Privacy Policy URL")}>
            <Input {...register("privacyPolicyUrl")} />
          </Field>

          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label>{__("Consent Expiry (days)")}</Label>
              <Input
                type="number"
                {...register("consentExpiryDays")}
                min="1"
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{__("Consent Mode")}</Label>
              <Controller
                name="consentMode"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <Option value="OPT_IN">{__("Opt-in")}</Option>
                    <Option value="OPT_OUT">{__("Opt-out")}</Option>
                  </Select>
                )}
              />
            </div>
            <div className="space-y-2">
              <Label>{__("Default Language")}</Label>
              <Controller
                name="defaultLanguage"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <Option value="en">{__("English")}</Option>
                    <Option value="fr">{__("French")}</Option>
                    <Option value="de">{__("German")}</Option>
                    <Option value="es">{__("Spanish")}</Option>
                  </Select>
                )}
              />
            </div>
          </div>

          <Button type="submit" disabled={isUpdating}>
            {isUpdating ? __("Saving...") : __("Save")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
