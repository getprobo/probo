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
import { ALL_KEYS } from "./translationDefaults";

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

export type TranslationFormValues = Record<string, string>;

interface TranslationEditorProps {
  cookieBannerId: string;
  language: string;
  existingTranslations: Record<string, string> | null;
  showBranding: boolean;
  categoryNames: string[];
  necessaryCategoryName: string;
}

export function TranslationEditor({
  cookieBannerId,
  language,
  existingTranslations,
  showBranding,
  categoryNames,
  necessaryCategoryName,
}: TranslationEditorProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [upsertTranslation, isUpserting]
    = useMutation<TranslationEditorMutation>(upsertTranslationMutation);

  const defaultValues = useMemo(() => {
    const values: Record<string, string> = {};
    for (const key of ALL_KEYS) {
      values[key] = existingTranslations?.[key] ?? "";
    }
    return values;
  }, [existingTranslations]);

  const methods = useForm<TranslationFormValues>({
    defaultValues,
  });

  const handleSave = (formData: TranslationFormValues) => {
    upsertTranslation({
      variables: {
        input: {
          cookieBannerId,
          language,
          translations: JSON.stringify(formData),
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
          categoryNames={categoryNames}
          necessaryCategoryName={necessaryCategoryName}
        />
        <PlaceholderTranslationSection
          exampleCategoryName={categoryNames[1] ?? categoryNames[0] ?? "Analytics"}
        />

        <Button type="submit" disabled={isUpserting}>
          {isUpserting ? __("Saving...") : __("Save translations")}
        </Button>
      </form>
    </FormProvider>
  );
}
