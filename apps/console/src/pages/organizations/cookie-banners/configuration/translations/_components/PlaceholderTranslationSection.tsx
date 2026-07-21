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

import { Card, Field, Input } from "@probo/ui";
import { Controller, useFormContext, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { PlaceholderPreview } from "./PlaceholderPreview";
import type { TranslationFormValues } from "./TranslationEditor";

interface PlaceholderTranslationSectionProps {
  exampleCategoryName: string;
}

export function PlaceholderTranslationSection({
  exampleCategoryName,
}: PlaceholderTranslationSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { control } = useFormContext<TranslationFormValues>();

  const placeholderText = useWatch({ control, name: "placeholder_text" });
  const placeholderButton = useWatch({ control, name: "placeholder_button" });

  return (
    <div className="space-y-4">
      <h3 className="font-medium text-lg">
        {t("placeholderTranslationSection.title")}
      </h3>
      <p className="text-sm text-txt-secondary">
        {t("placeholderTranslationSection.description")}
      </p>
      <div className="grid grid-cols-2 gap-6">
        <Card className="border p-4">
          <div className="space-y-4">
            <Controller
              control={control}
              name="placeholder_text"
              render={({ field }) => (
                <Field
                  label={t("translationEditor.labels.placeholderText")}
                >
                  <p className="text-xs text-txt-secondary mb-2">
                    {t("placeholderTranslationSection.categoryHelp")}
                  </p>
                  <Input {...field} />
                </Field>
              )}
            />
            <Controller
              control={control}
              name="placeholder_button"
              render={({ field }) => (
                <Field label={t("translationEditor.labels.placeholderButton")}>
                  <Input {...field} />
                </Field>
              )}
            />
          </div>
        </Card>

        <div className="flex items-start justify-center rounded-lg border border-border-low bg-[repeating-conic-gradient(#e5e7eb_0%_25%,transparent_0%_50%)] bg-size-[20px_20px] p-6">
          <PlaceholderPreview
            placeholderText={placeholderText}
            placeholderButton={placeholderButton}
            categoryName={exampleCategoryName}
          />
        </div>
      </div>
    </div>
  );
}
