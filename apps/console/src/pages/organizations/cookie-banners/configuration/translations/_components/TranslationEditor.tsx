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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { useMemo } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { TranslationEditorMutation } from "#/__generated__/core/TranslationEditorMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerTranslationsPage } from "../../../variants";
import {
  ALL_KEYS,
  type CategoryInfo,
  type CategoryTranslations,
  type TranslationFormValues,
} from "../_lib/translationDefaults";

import { BannerTranslationSection } from "./BannerTranslationSection";
import { PanelTranslationSection } from "./PanelTranslationSection";
import { PlaceholderTranslationSection } from "./PlaceholderTranslationSection";

const upsertTranslationMutation = graphql`
  mutation TranslationEditorMutation(
    $input: UpsertCookieBannerTranslationInput!
  ) {
    upsertCookieBannerTranslation(input: $input) {
      cookieBanner {
        id
        translations {
          id
          language
          translations
        }
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface TranslationEditorProps {
  cookieBannerId: string;
  language: string;
  existingTranslations: Record<string, string> | null;
  existingCategoryTranslations: CategoryTranslations | null;
  showBranding: boolean;
  categories: CategoryInfo[];
  necessaryCategoryName: string;
}

export function TranslationEditor({
  cookieBannerId,
  language,
  existingTranslations,
  existingCategoryTranslations,
  showBranding,
  categories,
  necessaryCategoryName,
}: TranslationEditorProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { form, actions } = cookieBannerTranslationsPage();
  const [upsertTranslation, isUpserting] = useMutation<TranslationEditorMutation>(
    upsertTranslationMutation,
    {
      successMessage: t("translationEditor.messages.saved"),
      errorToast: t("translationEditor.errors.save"),
    },
  );

  const defaultValues = useMemo(() => {
    const translations: Record<string, string> = {};
    for (const key of ALL_KEYS) {
      translations[key] = existingTranslations?.[key] ?? "";
    }

    const catDefaults: CategoryTranslations = {};
    for (const category of categories) {
      const existing = existingCategoryTranslations?.[category.id];
      catDefaults[category.id] = {
        name: existing?.name ?? "",
        description: existing?.description ?? "",
      };
    }

    return {
      ...translations,
      categories: catDefaults,
    } as TranslationFormValues;
  }, [existingTranslations, existingCategoryTranslations, categories]);

  const methods = useForm<TranslationFormValues>({
    defaultValues,
  });

  function handleSave(formData: TranslationFormValues) {
    const { categories: catTranslations, ...translations } = formData;
    const payload: Record<string, unknown> = { ...translations };

    const nonEmpty: CategoryTranslations = {};
    for (const [id, entry] of Object.entries(catTranslations)) {
      if (entry.name || entry.description) {
        nonEmpty[id] = entry;
      }
    }
    if (Object.keys(nonEmpty).length > 0) {
      payload.categories = nonEmpty;
    }

    void upsertTranslation({
      variables: {
        input: {
          cookieBannerId,
          language,
          translations: JSON.stringify(payload),
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <FormProvider {...methods}>
      <form
        className={form()}
        onSubmit={event => void methods.handleSubmit(handleSave)(event)}
      >
        <BannerTranslationSection showBranding={showBranding} />
        <PanelTranslationSection
          categories={categories}
          necessaryCategoryName={necessaryCategoryName}
        />
        <PlaceholderTranslationSection
          exampleCategoryName={
            categories[1]?.name
            ?? categories[0]?.name
            ?? t("translationEditor.exampleCategory")
          }
        />
        <div className={actions()}>
          <Button
            type="submit"
            variant="solid"
            color="neutral"
            highContrast
            loading={isUpserting}
          >
            {t("translationEditor.actions.save")}
          </Button>
        </div>
      </form>
    </FormProvider>
  );
}
