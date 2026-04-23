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

import { PanelPreview } from "./PanelPreview";
import { TRANSLATION_LABELS } from "./translationDefaults";
import type { TranslationFormValues } from "./TranslationEditor";

interface PanelTranslationSectionProps {
  categoryNames: string[];
  necessaryCategoryName: string;
}

export function PanelTranslationSection({
  categoryNames,
  necessaryCategoryName,
}: PanelTranslationSectionProps) {
  const { __ } = useTranslate();
  const { control } = useFormContext<TranslationFormValues>();

  const panelTitle = useWatch({ control, name: "panel_title" });
  const panelDescription = useWatch({ control, name: "panel_description" });
  const buttonSave = useWatch({ control, name: "button_save" });

  return (
    <div className="space-y-4">
      <h3 className="font-medium text-lg">{__("Preferences panel")}</h3>
      <div className="grid grid-cols-2 gap-6">
        <Card className="border p-4">
          <div className="space-y-4">
            <Controller
              control={control}
              name="panel_title"
              render={({ field }) => (
                <Field label={__(TRANSLATION_LABELS.panel_title)}>
                  <Input {...field} />
                </Field>
              )}
            />
            <Controller
              control={control}
              name="panel_description"
              render={({ field }) => (
                <Field
                  label={__(TRANSLATION_LABELS.panel_description)}
                  // description={__(
                  //   "Use {{necessary_category}} to insert the necessary category name.",
                  // )}
                >
                  <Textarea {...field} rows={3} />
                </Field>
              )}
            />
            <Controller
              control={control}
              name="button_save"
              render={({ field }) => (
                <Field label={__(TRANSLATION_LABELS.button_save)}>
                  <Input {...field} />
                </Field>
              )}
            />

            <div className="space-y-4 border-t border-border-low pt-4">
              <h4 className="text-sm font-medium text-txt-secondary">
                {__("Accessibility labels")}
              </h4>
              <div className="grid grid-cols-2 gap-4">
                <Controller
                  control={control}
                  name="aria_close"
                  render={({ field }) => (
                    <Field label={__(TRANSLATION_LABELS.aria_close)}>
                      <Input {...field} />
                    </Field>
                  )}
                />
                <Controller
                  control={control}
                  name="aria_cookie_settings"
                  render={({ field }) => (
                    <Field
                      label={__(TRANSLATION_LABELS.aria_cookie_settings)}
                    >
                      <Input {...field} />
                    </Field>
                  )}
                />
                <Controller
                  control={control}
                  name="aria_show_details"
                  render={({ field }) => (
                    <Field label={__(TRANSLATION_LABELS.aria_show_details)}>
                      <Input {...field} />
                    </Field>
                  )}
                />
                <Controller
                  control={control}
                  name="aria_hide_details"
                  render={({ field }) => (
                    <Field label={__(TRANSLATION_LABELS.aria_hide_details)}>
                      <Input {...field} />
                    </Field>
                  )}
                />
              </div>
            </div>
          </div>
        </Card>

        <div className="flex items-start justify-center rounded-lg border border-border-low bg-[repeating-conic-gradient(#e5e7eb_0%_25%,transparent_0%_50%)] bg-size-[20px_20px] p-6">
          <PanelPreview
            panelTitle={panelTitle}
            panelDescription={panelDescription}
            buttonSave={buttonSave}
            categoryNames={categoryNames}
            necessaryCategoryName={necessaryCategoryName}
          />
        </div>
      </div>
    </div>
  );
}
