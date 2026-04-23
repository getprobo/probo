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
import { Card, Field, Input } from "@probo/ui";
import { Controller, useFormContext, useWatch } from "react-hook-form";

import { PlaceholderPreview } from "./PlaceholderPreview";
import { TRANSLATION_LABELS } from "./translationDefaults";
import type { TranslationFormValues } from "./TranslationEditor";

interface PlaceholderTranslationSectionProps {
  exampleCategoryName: string;
}

export function PlaceholderTranslationSection({
  exampleCategoryName,
}: PlaceholderTranslationSectionProps) {
  const { __ } = useTranslate();
  const { control } = useFormContext<TranslationFormValues>();

  const placeholderText = useWatch({ control, name: "placeholder_text" });
  const placeholderButton = useWatch({ control, name: "placeholder_button" });

  return (
    <div className="space-y-4">
      <h3 className="font-medium text-lg">{__("Placeholder")}</h3>
      <p className="text-sm text-txt-secondary">
        {__(
          "Shown in place of blocked content until the visitor gives consent.",
        )}
      </p>
      <div className="grid grid-cols-2 gap-6">
        <Card className="border p-4">
          <div className="space-y-4">
            <Controller
              control={control}
              name="placeholder_text"
              render={({ field }) => (
                <Field
                  label={__(TRANSLATION_LABELS.placeholder_text)}
                >
                  <Input {...field} />
                </Field>
              )}
            />
            <Controller
              control={control}
              name="placeholder_button"
              render={({ field }) => (
                <Field label={__(TRANSLATION_LABELS.placeholder_button)}>
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
