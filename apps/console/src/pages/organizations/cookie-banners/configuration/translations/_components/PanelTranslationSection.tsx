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

      {categories.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-sm font-medium text-txt-secondary">
            {__("Category names")}
          </h4>
          <div className="grid grid-cols-2 gap-4">
            {categories.map(cat => (
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
