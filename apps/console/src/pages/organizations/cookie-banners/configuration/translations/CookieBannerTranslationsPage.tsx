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

import { usePageTitle } from "@probo/hooks";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerTranslationsPageQuery } from "#/__generated__/core/CookieBannerTranslationsPageQuery.graphql";

import { CookieBannerPageHeader } from "../../_components/CookieBannerPageHeader";
import { cookieBannerTranslationsPage } from "../../variants";

import { TranslationEditor } from "./_components/TranslationEditor";
import { SUPPORTED_LANGUAGES } from "./_lib/translationDefaults";

export const cookieBannerTranslationsPageQuery = graphql`
  query CookieBannerTranslationsPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
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
        categories(first: 50, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED }) @required(action: THROW) {
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

export function CookieBannerTranslationsPage({
  queryRef,
}: CookieBannerTranslationsPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const title = t("translationsPage.title");
  usePageTitle(title);
  const { root, toolbar, language } = cookieBannerTranslationsPage();
  const data = usePreloadedQuery<CookieBannerTranslationsPageQuery>(
    cookieBannerTranslationsPageQuery,
    queryRef,
  );

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  const [selectedLanguage, setSelectedLanguage] = useState(
    () => banner.defaultLanguage,
  );

  const selectedTranslation = banner.translations.find(
    translation => translation.language === selectedLanguage,
  );

  const { uiStrings, categoryTranslations } = useMemo(() => {
    if (!selectedTranslation) {
      return { uiStrings: null, categoryTranslations: null };
    }
    try {
      const raw = JSON.parse(selectedTranslation.translations) as Record<string, unknown>;
      const ui: Record<string, string> = {};
      let cats: Record<string, { name: string; description: string }> | null = null;

      for (const [key, value] of Object.entries(raw)) {
        if (key === "categories" && typeof value === "object" && value !== null) {
          cats = value as Record<string, { name: string; description: string }>;
        } else if (typeof value === "string") {
          ui[key] = value;
        }
      }

      return { uiStrings: ui, categoryTranslations: cats };
    } catch {
      return { uiStrings: null, categoryTranslations: null };
    }
  }, [selectedTranslation]);

  const categories = useMemo(
    () =>
      banner.categories.edges.map(edge => ({
        id: edge.node.id,
        name: edge.node.name,
        slug: edge.node.slug,
        description: edge.node.description,
        kind: edge.node.kind,
      })),
    [banner.categories],
  );

  const necessaryCategoryName = useMemo(
    () => categories.find(category => category.kind === "NECESSARY")?.name
      ?? t("translationsPage.necessaryFallback"),
    [categories, t],
  );

  return (
    <div className={root()}>
      <CookieBannerPageHeader
        title={title}
        description={t("translationsPage.description")}
      />
      <div className={toolbar()}>
        <div className={language()}>
          <Select
            value={selectedLanguage}
            onValueChange={(value: string | null) => {
              if (value != null) {
                setSelectedLanguage(value);
              }
            }}
          >
            <SelectTrigger size={2} aria-label={t("translationsPage.language")}>
              {(value: string | null) => {
                const selected = SUPPORTED_LANGUAGES.find(item => item.code === value);
                if (selected == null) {
                  return "";
                }
                return selected.code === banner.defaultLanguage
                  ? t("translationsPage.languageDefault", { language: selected.label })
                  : selected.label;
              }}
            </SelectTrigger>
            <SelectPopup>
              {SUPPORTED_LANGUAGES.map(item => (
                <SelectItem key={item.code} value={item.code}>
                  {item.code === banner.defaultLanguage
                    ? t("translationsPage.languageDefault", { language: item.label })
                    : item.label}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
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
        <p>
          <Text size={2} color="faint">
            {t("translationsPage.defaultLanguageDescription")}
          </Text>
        </p>
      )}
    </div>
  );
}
