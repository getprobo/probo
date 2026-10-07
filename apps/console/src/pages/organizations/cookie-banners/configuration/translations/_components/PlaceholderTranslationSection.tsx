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
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Controller, useFormContext, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { cookieBannerTranslationsPage } from "../../../variants";
import type { TranslationFormValues } from "../_lib/translationDefaults";

import { PlaceholderPreview } from "./PlaceholderPreview";

interface PlaceholderTranslationSectionProps {
  exampleCategoryName: string;
}

export function PlaceholderTranslationSection({
  exampleCategoryName,
}: PlaceholderTranslationSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { control } = useFormContext<TranslationFormValues>();
  const { section, split, fields, preview } = cookieBannerTranslationsPage();

  const placeholderText = useWatch({ control, name: "placeholder_text" });
  const placeholderButton = useWatch({ control, name: "placeholder_button" });

  return (
    <section className={section()}>
      <Heading level={2} size={4} weight="medium" highContrast>
        {t("placeholderTranslationSection.title")}
      </Heading>
      <Text size={2} color="faint">
        {t("placeholderTranslationSection.description")}
      </Text>
      <div className={split()}>
        <Card size={2} variant="soft">
          <div className={fields()}>
            <Controller
              control={control}
              name="placeholder_text"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.placeholderText")}>
                  <TextField
                    name={field.name}
                    value={field.value}
                    onValueChange={field.onChange}
                  />
                </Field>
              )}
            />
            <Text size={1} color="faint">
              {t("placeholderTranslationSection.categoryHelp")}
            </Text>
            <Controller
              control={control}
              name="placeholder_button"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.placeholderButton")}>
                  <TextField
                    name={field.name}
                    value={field.value}
                    onValueChange={field.onChange}
                  />
                </Field>
              )}
            />
          </div>
        </Card>
        <div className={preview()}>
          <PlaceholderPreview
            placeholderText={placeholderText}
            placeholderButton={placeholderButton}
            categoryName={exampleCategoryName}
          />
        </div>
      </div>
    </section>
  );
}
