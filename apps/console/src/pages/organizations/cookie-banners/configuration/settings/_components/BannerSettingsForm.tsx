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

import { formatError } from "@probo/helpers";
import { Button, Card, Field, Input, Label, Option, Select, useToast } from "@probo/ui";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
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
  defaultLanguage: string;
}

interface BannerSettingsFormProps {
  cookieBannerKey: BannerSettingsForm_cookieBanner$key;
}

export function BannerSettingsForm({ cookieBannerKey }: BannerSettingsFormProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();

  const banner = useFragment(bannerSettingsFormFragment, cookieBannerKey);

  const [updateBanner, isUpdating] = useMutation<BannerSettingsFormMutation>(updateBannerMutation);

  const { register, handleSubmit, control } = useForm<BannerSettingsFormValues>({
    defaultValues: {
      name: banner.name,
      cookiePolicyUrl: banner.cookiePolicyUrl,
      privacyPolicyUrl: banner.privacyPolicyUrl ?? "",
      consentExpiryDays: String(banner.consentExpiryDays),
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
          defaultLanguage: data.defaultLanguage,
        },
      },
      onCompleted() {
        toast({ title: t("bannerSettingsForm.messages.successTitle"), description: t("bannerSettingsForm.messages.updated"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("bannerSettingsForm.errors.title"), description: formatError(t("bannerSettingsForm.errors.update"), error), variant: "error" });
      },
    });
  };

  return (
    <div className="space-y-4">
      <h3 className="font-medium">{t("bannerSettingsForm.title")}</h3>
      <Card className="border p-4">
        <form className="space-y-4" onSubmit={e => void handleSubmit(onSubmit)(e)}>
          <Field label={t("bannerSettingsForm.fields.name")}>
            <Input {...register("name")} required />
          </Field>

          <Field label={t("bannerSettingsForm.fields.origin")}>
            <Input value={banner.origin} disabled />
          </Field>

          <Field label={t("bannerSettingsForm.fields.cookiePolicyUrl")}>
            <Input {...register("cookiePolicyUrl")} required />
          </Field>

          <Field label={t("bannerSettingsForm.fields.privacyPolicyUrl")}>
            <Input {...register("privacyPolicyUrl")} />
          </Field>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t("bannerSettingsForm.fields.consentExpiryDays")}</Label>
              <Input
                type="number"
                {...register("consentExpiryDays")}
                min="1"
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{t("bannerSettingsForm.fields.defaultLanguage")}</Label>
              <Controller
                name="defaultLanguage"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <Option value="en">{t("bannerSettingsForm.languages.english")}</Option>
                    <Option value="fr">{t("bannerSettingsForm.languages.french")}</Option>
                    <Option value="de">{t("bannerSettingsForm.languages.german")}</Option>
                    <Option value="es">{t("bannerSettingsForm.languages.spanish")}</Option>
                  </Select>
                )}
              />
            </div>
          </div>

          <Button type="submit" disabled={isUpdating}>
            {isUpdating ? t("bannerSettingsForm.actions.saving") : t("bannerSettingsForm.actions.save")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
