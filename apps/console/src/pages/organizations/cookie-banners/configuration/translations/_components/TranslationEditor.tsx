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

import { formatError, type GraphQLError } from "@probo/helpers";
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
      if (cat.kind === "UNCATEGORISED") continue;
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
            error as GraphQLError,
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
