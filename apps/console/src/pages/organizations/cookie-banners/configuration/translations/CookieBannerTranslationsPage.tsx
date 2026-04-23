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
import { Option, Select } from "@probo/ui";
import { useMemo, useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerTranslationsPageQuery } from "#/__generated__/core/CookieBannerTranslationsPageQuery.graphql";

import { SUPPORTED_LANGUAGES } from "./_components/translationDefaults";
import { TranslationEditor } from "./_components/TranslationEditor";

export const cookieBannerTranslationsPageQuery = graphql`
  query CookieBannerTranslationsPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        defaultLanguage
        showBranding
        translations {
          id
          language
          translations
        }
        categories(first: 50, orderBy: { field: RANK, direction: ASC }) @required(action: THROW) {
          edges {
            node {
              id
              name
              slug
              description
              kind
            }
          }
        }
      }
    }
  }
`;

interface CookieBannerTranslationsPageProps {
  queryRef: PreloadedQuery<CookieBannerTranslationsPageQuery>;
}

export default function CookieBannerTranslationsPage({
  queryRef,
}: CookieBannerTranslationsPageProps) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerTranslationsPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;

  const existingLanguages = useMemo(
    () => banner.translations.map(t => t.language),
    [banner.translations],
  );

  const [selectedLanguage, setSelectedLanguage] = useState(
    () => banner.defaultLanguage,
  );

  const selectedTranslation = banner.translations.find(
    t => t.language === selectedLanguage,
  );

  const { uiStrings, categoryTranslations } = useMemo(() => {
    if (!selectedTranslation) {
      return { uiStrings: null, categoryTranslations: null };
    }
    try {
      const raw = JSON.parse(selectedTranslation.translations) as Record<string, unknown>;
      const ui: Record<string, string> = {};
      let cats: Record<string, { name: string; description: string }> | null = null;

      for (const [k, v] of Object.entries(raw)) {
        if (k === "categories" && typeof v === "object" && v !== null) {
          cats = v as Record<string, { name: string; description: string }>;
        } else if (typeof v === "string") {
          ui[k] = v;
        }
      }

      return { uiStrings: ui, categoryTranslations: cats };
    } catch {
      return { uiStrings: null, categoryTranslations: null };
    }
  }, [selectedTranslation]);

  const categories = useMemo(
    () =>
      banner.categories.edges.map(e => ({
        id: e.node.id,
        name: e.node.name,
        slug: e.node.slug,
        description: e.node.description,
        kind: e.node.kind,
      })),
    [banner.categories],
  );

  const necessaryCategoryName = useMemo(
    () => categories.find(c => c.kind === "NECESSARY")?.name ?? "Necessary",
    [categories],
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Select
          value={selectedLanguage}
          onValueChange={setSelectedLanguage}
        >
          {SUPPORTED_LANGUAGES.filter(
            l =>
              existingLanguages.includes(l.code)
              || l.code === selectedLanguage,
          ).map(l => (
            <Option key={l.code} value={l.code}>
              {l.label}
              {l.code === banner.defaultLanguage ? ` (${__("default")})` : ""}
            </Option>
          ))}
        </Select>
      </div>

      <TranslationEditor
        key={selectedLanguage}
        cookieBannerId={banner.id}
        language={selectedLanguage}
        existingTranslations={uiStrings}
        existingCategoryTranslations={categoryTranslations}
        showBranding={banner.showBranding}
        categories={categories}
        necessaryCategoryName={necessaryCategoryName}
      />

      {selectedLanguage === banner.defaultLanguage && (
        <p className="text-sm text-txt-secondary">
          {__(
            "This is the default language. These translations are shown when a visitor's language is not available.",
          )}
        </p>
      )}
    </div>
  );
}
