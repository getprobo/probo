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

import { useTranslate } from "@probo/i18n";
import { Card, Field, Input, Textarea } from "@probo/ui";
import { Controller, useFormContext, useWatch } from "react-hook-form";

import { BannerPreview } from "./BannerPreview";
import { TRANSLATION_LABELS } from "./translationDefaults";
import type { TranslationFormValues } from "./TranslationEditor";

interface BannerTranslationSectionProps {
  showBranding: boolean;
}

export function BannerTranslationSection({
  showBranding,
}: BannerTranslationSectionProps) {
  const { __ } = useTranslate();
  const { control } = useFormContext<TranslationFormValues>();

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
    <div className="space-y-4">
      <h3 className="font-medium text-lg">{__("Banner")}</h3>
      <div className="grid grid-cols-2 gap-6">
        <Card className="border p-4">
          <div className="space-y-4">
            <Controller
              control={control}
              name="banner_title"
              render={({ field }) => (
                <Field label={__(TRANSLATION_LABELS.banner_title)}>
                  <Input {...field} />
                </Field>
              )}
            />
            <Controller
              control={control}
              name="banner_description"
              render={({ field }) => (
                <Field
                  label={__(TRANSLATION_LABELS.banner_description)}
                >
                  <p className="text-xs text-txt-secondary mb-2">{"Use {{cookie_policy_link}} and {{privacy_policy_link}} to insert policy links."}</p>
                  <Textarea {...field} rows={3} />
                </Field>
              )}
            />
            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={control}
                name="button_accept_all"
                render={({ field }) => (
                  <Field label={__(TRANSLATION_LABELS.button_accept_all)}>
                    <Input {...field} />
                  </Field>
                )}
              />
              <Controller
                control={control}
                name="button_reject_all"
                render={({ field }) => (
                  <Field label={__(TRANSLATION_LABELS.button_reject_all)}>
                    <Input {...field} />
                  </Field>
                )}
              />
            </div>
            <Controller
              control={control}
              name="button_customize"
              render={({ field }) => (
                <Field label={__(TRANSLATION_LABELS.button_customize)}>
                  <Input {...field} />
                </Field>
              )}
            />
            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={control}
                name="cookie_policy_link_text"
                render={({ field }) => (
                  <Field
                    label={__(TRANSLATION_LABELS.cookie_policy_link_text)}
                  >
                    <Input {...field} />
                  </Field>
                )}
              />
              <Controller
                control={control}
                name="privacy_policy_link_text"
                render={({ field }) => (
                  <Field
                    label={__(TRANSLATION_LABELS.privacy_policy_link_text)}
                  >
                    <Input {...field} />
                  </Field>
                )}
              />
            </div>
          </div>
        </Card>

        <div className="flex items-start justify-center rounded-lg border border-border-low bg-[repeating-conic-gradient(#e5e7eb_0%_25%,transparent_0%_50%)] bg-size-[20px_20px] p-6">
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
    </div>
  );
}
