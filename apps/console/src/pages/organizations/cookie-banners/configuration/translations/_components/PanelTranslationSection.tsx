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
import type {
  CategoryInfo,
  TranslationFormValues,
} from "./TranslationEditor";

interface PanelTranslationSectionProps {
  categories: CategoryInfo[];
  necessaryCategoryName: string;
}

export function PanelTranslationSection({
  categories,
  necessaryCategoryName,
}: PanelTranslationSectionProps) {
  const { __ } = useTranslate();
  const { control } = useFormContext<TranslationFormValues>();

  const panelTitle = useWatch({ control, name: "panel_title" });
  const panelDescription = useWatch({ control, name: "panel_description" });
  const buttonAcceptAll = useWatch({ control, name: "button_accept_all" });
  const buttonRejectAll = useWatch({ control, name: "button_reject_all" });
  const buttonSave = useWatch({ control, name: "button_save" });
  const categoryTranslations = useWatch({ control, name: "categories" });

  const visibleCategories = categories.filter(c => c.kind !== "UNCATEGORISED");

  const translatedNecessaryName = (() => {
    const necessaryCat = categories.find(c => c.kind === "NECESSARY");
    if (!necessaryCat) return necessaryCategoryName;
    return categoryTranslations?.[necessaryCat.id]?.name || necessaryCategoryName;
  })();

  const previewCategories = categories.map((c) => {
    const translated = categoryTranslations?.[c.id];
    return {
      name: translated?.name || c.name,
      description: translated?.description || c.description,
      isNecessary: c.kind === "NECESSARY",
    };
  });

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
                >
                  <p className="text-xs text-txt-secondary mb-2">{__("Use {{necessary_category}} to refer to the required cookies category name.")}</p>
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
            buttonAcceptAll={buttonAcceptAll}
            buttonRejectAll={buttonRejectAll}
            buttonSave={buttonSave}
            categories={previewCategories}
            necessaryCategoryName={translatedNecessaryName}
          />
        </div>
      </div>

      {visibleCategories.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-sm font-medium text-txt-secondary">
            {__("Category names")}
          </h4>
          <div className="grid grid-cols-2 gap-4">
            {visibleCategories.map(cat => (
              <Card key={cat.id} className="border p-4 space-y-3">
                <div className="text-sm text-txt-secondary">
                  {cat.name}
                  {" "}
                  <span className="text-txt-secondary/60">
                    {`(${cat.slug})`}
                  </span>
                </div>
                <Controller
                  control={control}
                  name={`categories.${cat.id}.name`}
                  render={({ field }) => (
                    <Field label={__("Translated name")}>
                      <Input {...field} placeholder={cat.name} />
                    </Field>
                  )}
                />
                <Controller
                  control={control}
                  name={`categories.${cat.id}.description`}
                  render={({ field }) => (
                    <Field label={__("Translated description")}>
                      <Textarea {...field} placeholder={cat.description} rows={2} />
                    </Field>
                  )}
                />
              </Card>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
