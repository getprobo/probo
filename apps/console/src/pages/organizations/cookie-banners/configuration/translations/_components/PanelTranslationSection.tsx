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
import type {
  CategoryInfo,
  TranslationFormValues,
} from "../_lib/translationDefaults";

import { PanelPreview } from "./PanelPreview";

interface PanelTranslationSectionProps {
  categories: CategoryInfo[];
  necessaryCategoryName: string;
}

export function PanelTranslationSection({
  categories,
  necessaryCategoryName,
}: PanelTranslationSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { control } = useFormContext<TranslationFormValues>();
  const { section, split, fields, pair, preview, categoryGrid } = cookieBannerTranslationsPage();

  const panelTitle = useWatch({ control, name: "panel_title" });
  const panelDescription = useWatch({ control, name: "panel_description" });
  const buttonAcceptAll = useWatch({ control, name: "button_accept_all" });
  const buttonRejectAll = useWatch({ control, name: "button_reject_all" });
  const buttonSave = useWatch({ control, name: "button_save" });
  const categoryTranslations = useWatch({ control, name: "categories" });

  const necessaryCategory = categories.find(category => category.kind === "NECESSARY");
  const translatedNecessaryName = necessaryCategory == null
    ? necessaryCategoryName
    : (categoryTranslations?.[necessaryCategory.id]?.name || necessaryCategoryName);

  const previewCategories = categories.map((category) => {
    const translated = categoryTranslations?.[category.id];
    return {
      name: translated?.name || category.name,
      description: translated?.description || category.description,
      isNecessary: category.kind === "NECESSARY",
    };
  });

  return (
    <section className={section()}>
      <Heading level={2} size={4} weight="medium" highContrast>
        {t("panelTranslationSection.title")}
      </Heading>
      <div className={split()}>
        <Card size={2} variant="soft">
          <div className={fields()}>
            <Controller
              control={control}
              name="panel_title"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.panelTitle")}>
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
              name="panel_description"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.panelDescription")}>
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
              {t("panelTranslationSection.necessaryCategoryHelp")}
            </Text>
            <Controller
              control={control}
              name="button_save"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.saveButton")}>
                  <TextField
                    name={field.name}
                    value={field.value}
                    onValueChange={field.onChange}
                  />
                </Field>
              )}
            />
            <Text size={2} weight="medium" highContrast>
              {t("panelTranslationSection.accessibilityLabels")}
            </Text>
            <div className={pair()}>
              <Controller
                control={control}
                name="aria_close"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.ariaClose")}>
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
                name="aria_cookie_settings"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.ariaCookieSettings")}>
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
                name="aria_show_details"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.ariaShowDetails")}>
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
                name="aria_hide_details"
                render={({ field }) => (
                  <Field label={t("translationEditor.labels.ariaHideDetails")}>
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
        <div className={fields()}>
          <Text size={2} weight="medium" highContrast>
            {t("panelTranslationSection.categoryNames")}
          </Text>
          <div className={categoryGrid()}>
            {categories.map(category => (
              <Card key={category.id} size={2} variant="soft">
                <div className={fields()}>
                  <Text size={1} color="faint">
                    {category.name}
                    {" "}
                    {`(${category.slug})`}
                  </Text>
                  <Controller
                    control={control}
                    name={`categories.${category.id}.name`}
                    render={({ field }) => (
                      <Field label={t("panelTranslationSection.translatedName")}>
                        <TextField
                          name={field.name}
                          value={field.value}
                          placeholder={category.name}
                          onValueChange={field.onChange}
                        />
                      </Field>
                    )}
                  />
                  <Controller
                    control={control}
                    name={`categories.${category.id}.description`}
                    render={({ field }) => (
                      <Field label={t("panelTranslationSection.translatedDescription")}>
                        <Textarea
                          name={field.name}
                          rows={2}
                          value={field.value}
                          placeholder={category.description}
                          onChange={field.onChange}
                        />
                      </Field>
                    )}
                  />
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}
