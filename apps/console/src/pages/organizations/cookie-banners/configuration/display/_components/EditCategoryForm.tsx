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

import { Button, Input, Textarea } from "@probo/ui";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

const GCM_CONSENT_TYPES = [
  "analytics_storage",
  "ad_storage",
  "ad_user_data",
  "ad_personalization",
  "functionality_storage",
  "personalization_storage",
  "security_storage",
] as const;

interface CategoryFormValues {
  name: string;
  slug: string;
  description: string;
  gcmConsentTypes: string[];
  posthogConsent: boolean;
}

interface EditCategoryFormProps {
  name: string;
  slug: string;
  description: string;
  kind: string;
  gcmConsentTypes: string[];
  posthogConsent: boolean;
  isUpdating: boolean;
  onSave: (name: string, slug: string, description: string, gcmConsentTypes: string[], posthogConsent: boolean) => void;
  onCancel: () => void;
}

export function EditCategoryForm({
  name,
  slug,
  description,
  kind,
  gcmConsentTypes,
  posthogConsent,
  isUpdating,
  onSave,
  onCancel,
}: EditCategoryFormProps) {
  const { t } = useTranslation("organizations/cookie-banners");

  const { register, handleSubmit, control } = useForm<CategoryFormValues>({
    defaultValues: {
      name,
      slug,
      description,
      gcmConsentTypes,
      posthogConsent,
    },
  });

  const onSubmit = (data: CategoryFormValues) => {
    onSave(data.name, data.slug, data.description, data.gcmConsentTypes, data.posthogConsent);
  };

  return (
    <div className="space-y-3">
      <Input
        {...register("name")}
        placeholder={t("editCategoryForm.fields.namePlaceholder")}
      />
      <Input
        {...register("slug", {
          pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
        })}
        placeholder={t("editCategoryForm.fields.slugPlaceholder")}
      />
      <Textarea
        {...register("description")}
        placeholder={t("editCategoryForm.fields.descriptionPlaceholder")}
        rows={2}
      />
      <div>
        <label className="text-sm font-medium">
          {t("editCategoryForm.googleConsentMode.title")}
        </label>
        <p className="text-xs text-muted-foreground mb-2">
          {t("editCategoryForm.googleConsentMode.description")}
        </p>
        <div className="flex flex-wrap gap-2">
          <Controller
            name="gcmConsentTypes"
            control={control}
            render={({ field }) => (
              <>
                {GCM_CONSENT_TYPES.map(type => (
                  <label
                    key={type}
                    className="flex items-center gap-1.5 text-xs cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={field.value.includes(type)}
                      onChange={() => {
                        const next = field.value.includes(type)
                          ? field.value.filter(t => t !== type)
                          : [...field.value, type];
                        field.onChange(next);
                      }}
                      className="rounded"
                    />
                    <code className="font-mono">{type}</code>
                  </label>
                ))}
              </>
            )}
          />
        </div>
      </div>
      {kind === "NORMAL" && (
        <div>
          <label className="text-sm font-medium">
            {t("editCategoryForm.posthog.title")}
          </label>
          <p className="text-xs text-muted-foreground mb-2">
            {t("editCategoryForm.posthog.description")}
          </p>
          <label className="flex items-center gap-1.5 text-xs cursor-pointer">
            <input
              type="checkbox"
              {...register("posthogConsent")}
              className="rounded"
            />
            <span>{t("editCategoryForm.posthog.checkbox")}</span>
          </label>
        </div>
      )}
      <div className="flex items-center gap-2">
        <Button
          onClick={() => void handleSubmit(onSubmit)()}
          disabled={isUpdating}
        >
          {isUpdating ? t("editCategoryForm.actions.saving") : t("editCategoryForm.actions.save")}
        </Button>
        <Button
          variant="secondary"
          onClick={onCancel}
        >
          {t("editCategoryForm.actions.cancel")}
        </Button>
      </div>
    </div>
  );
}
