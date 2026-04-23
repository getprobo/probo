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
                  <p className="text-xs text-txt-secondary mb-2">{"Use {{privacy_policy_link}} to insert the privacy policy link."}</p>
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
            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={control}
                name="button_customize"
                render={({ field }) => (
                  <Field label={__(TRANSLATION_LABELS.button_customize)}>
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
            privacyPolicyLinkText={privacyPolicyLinkText}
            showBranding={showBranding}
          />
        </div>
      </div>
    </div>
  );
}
