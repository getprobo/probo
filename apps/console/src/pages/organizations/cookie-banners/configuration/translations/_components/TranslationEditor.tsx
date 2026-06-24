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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, useToast } from "@probo/ui";
import { useMemo } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { TranslationEditorMutation } from "#/__generated__/core/TranslationEditorMutation.graphql";

import { BannerTranslationSection } from "./BannerTranslationSection";
import { PanelTranslationSection } from "./PanelTranslationSection";
import { PlaceholderTranslationSection } from "./PlaceholderTranslationSection";
import { ALL_KEYS, type TranslationKey } from "./translationDefaults";

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

export type TranslationFormValues = Record<TranslationKey, string> & {
  categories: CategoryTranslations;
};

export interface CategoryInfo {
  id: string;
  name: string;
  slug: string;
  description: string;
  kind: string;
}

export type CategoryTranslations = Record<
  string,
  { name: string; description: string }
>;

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
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [upsertTranslation, isUpserting]
    = useMutation<TranslationEditorMutation>(upsertTranslationMutation);

  const defaultValues = useMemo(() => {
    const translations: Record<string, string> = {};
    for (const key of ALL_KEYS) {
      translations[key] = existingTranslations?.[key] ?? "";
    }

    const catDefaults: CategoryTranslations = {};
    for (const cat of categories) {
      const existing = existingCategoryTranslations?.[cat.id];
      catDefaults[cat.id] = {
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

  const handleSave = (formData: TranslationFormValues) => {
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

    upsertTranslation({
      variables: {
        input: {
          cookieBannerId,
          language,
          translations: JSON.stringify(payload),
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Translation saved"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to save translation"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <FormProvider {...methods}>
      <form
        className="space-y-8"
        onSubmit={e => void methods.handleSubmit(handleSave)(e)}
      >
        <BannerTranslationSection showBranding={showBranding} />
        <PanelTranslationSection
          categories={categories}
          necessaryCategoryName={necessaryCategoryName}
        />
        <PlaceholderTranslationSection
          exampleCategoryName={categories[1]?.name ?? categories[0]?.name ?? "Analytics"}
        />

        <Button type="submit" disabled={isUpserting}>
          {isUpserting ? __("Saving...") : __("Save translations")}
        </Button>
      </form>
    </FormProvider>
  );
}
