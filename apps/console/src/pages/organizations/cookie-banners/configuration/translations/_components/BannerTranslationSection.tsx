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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Controller, useFormContext, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { cookieBannerTranslationsPage } from "../../../variants";
import type { TranslationFormValues } from "../_lib/translationDefaults";

import { BannerPreview } from "./BannerPreview";

interface BannerTranslationSectionProps {
  showBranding: boolean;
}

export function BannerTranslationSection({
  showBranding,
}: BannerTranslationSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { control } = useFormContext<TranslationFormValues>();
  const { section, split, fields, pair, preview } = cookieBannerTranslationsPage();

  const bannerTitle = useWatch({ control, name: "banner_title" });
  const bannerDescription = useWatch({ control, name: "banner_description" });
  const buttonAcceptAll = useWatch({ control, name: "button_accept_all" });
  const buttonRejectAll = useWatch({ control, name: "button_reject_all" });
  const buttonCustomize = useWatch({ control, name: "button_customize" });
  const cookiePolicyLinkText = useWatch({
    control,
    name: "cookie_policy_link_text",
  });
  const privacyPolicyLinkText = useWatch({
    control,
    name: "privacy_policy_link_text",
  });

  return (
    <section className={section()}>
      <Heading level={2} size={4} weight="medium" highContrast>
        {t("bannerTranslationSection.title")}
      </Heading>
      <div className={split()}>
        <Card size={2} variant="soft">
          <div className={fields()}>
            <Controller
              control={control}
              name="banner_title"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.bannerTitle")}>
                  <TextField
                    name={field.name}
                    value={field.value}
                    onValueChange={field.onChange}
                  />
                </Field>
              )}
            />
            <Controller
              control={control}
              name="banner_description"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.bannerDescription")}>
                  <Textarea
                    name={field.name}
                    rows={3}
                    value={field.value}
                    onChange={field.onChange}
                  />
                </Field>
              )}
            />
            <Text size={1} color="faint">
              {t("bannerTranslationSection.policyLinkHelp")}
            </Text>
            <div className={pair()}>
              <Controller
                control={control}
                name="button_accept_all"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.acceptAllButton")}>
                    <TextField
                      name={field.name}
                      value={field.value}
                      onValueChange={field.onChange}
                    />
                  </Field>
                )}
              />
              <Controller
                control={control}
                name="button_reject_all"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.rejectAllButton")}>
                    <TextField
                      name={field.name}
                      value={field.value}
                      onValueChange={field.onChange}
                    />
                  </Field>
                )}
              />
            </div>
            <Controller
              control={control}
              name="button_customize"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.customizeButton")}>
                  <TextField
                    name={field.name}
                    value={field.value}
                    onValueChange={field.onChange}
                  />
                </Field>
              )}
            />
            <div className={pair()}>
              <Controller
                control={control}
                name="cookie_policy_link_text"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.cookiePolicyLinkText")}>
                    <TextField
                      name={field.name}
                      value={field.value}
                      onValueChange={field.onChange}
                    />
                  </Field>
                )}
              />
              <Controller
                control={control}
                name="privacy_policy_link_text"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.privacyPolicyLinkText")}>
                    <TextField
                      name={field.name}
                      value={field.value}
                      onValueChange={field.onChange}
                    />
                  </Field>
                )}
              />
            </div>
          </div>
        </Card>
        <div className={preview()}>
          <BannerPreview
            bannerTitle={bannerTitle}
            bannerDescription={bannerDescription}
            buttonAcceptAll={buttonAcceptAll}
            buttonRejectAll={buttonRejectAll}
            buttonCustomize={buttonCustomize}
            cookiePolicyLinkText={cookiePolicyLinkText}
            privacyPolicyLinkText={privacyPolicyLinkText}
            showBranding={showBranding}
          />
        </div>
      </div>
    </section>
  );
}
