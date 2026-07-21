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
import { dateTimeFormat, humanizeSeconds } from "@probo/i18n";
import { Badge, Card, IconSquareBehindSquare2, PropertyRow, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { TrackerPatternPropertiesSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternPropertiesSection_trackerPattern.graphql";
import type { TrackerPatternPropertiesSectionMoveMutation } from "#/__generated__/core/TrackerPatternPropertiesSectionMoveMutation.graphql";

import { MoveToCategorySelect } from "./MoveToCategorySelect";

const trackerPatternPropertiesSectionFragment = graphql`
  fragment TrackerPatternPropertiesSection_trackerPattern on TrackerPattern {
    id
    pattern
    matchType
    trackerType
    source
    maxAgeSeconds
    description
    excluded
    detectedCount
    lastMatchedAt
    commonTrackerPatternId
    cookieCategory {
      id
      name
      kind
    }
    thirdParty {
      name
    }
    commonThirdParty {
      name
    }
  }
`;

const movePatternMutation = graphql`
  mutation TrackerPatternPropertiesSectionMoveMutation(
    $input: MoveTrackerPatternToCategoryInput!
  ) {
    moveTrackerPatternToCategory(input: $input) {
      trackerPattern {
        id
        cookieCategory {
          id
          name
          kind
        }
      }
      cookieBanner {
        id
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface TrackerPatternPropertiesSectionProps {
  trackerPatternKey: TrackerPatternPropertiesSection_trackerPattern$key;
}

export function TrackerPatternPropertiesSection({
  trackerPatternKey,
}: TrackerPatternPropertiesSectionProps) {
  const { toast } = useToast();
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const pattern = useFragment<TrackerPatternPropertiesSection_trackerPattern$key>(
    trackerPatternPropertiesSectionFragment,
    trackerPatternKey,
  );

  const [movePattern]
    = useMutation<TrackerPatternPropertiesSectionMoveMutation>(movePatternMutation);

  const handleMove = (targetCategoryId: string) => {
    if (targetCategoryId === pattern.cookieCategory?.id) {
      return;
    }
    movePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          targetCookieCategoryId: targetCategoryId,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: t("trackerProperties.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: t("trackerProperties.messages.successTitle"), description: t("trackerProperties.messages.moved"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("trackerProperties.errors.title"), description: formatError(t("trackerProperties.errors.move"), error), variant: "error" });
      },
    });
  };

  const typeBadges = {
    COOKIE: { variant: "warning" as const, label: t("trackerProperties.trackerTypes.cookie") },
    LOCAL_STORAGE: { variant: "info" as const, label: t("trackerProperties.trackerTypes.localStorage") },
    SESSION_STORAGE: { variant: "highlight" as const, label: t("trackerProperties.trackerTypes.sessionStorage") },
    INDEXED_DB: { variant: "success" as const, label: t("trackerProperties.trackerTypes.indexedDb") },
    CACHE_STORAGE: { variant: "outline" as const, label: t("trackerProperties.trackerTypes.cacheStorage") },
  };
  const sourceBadges = {
    SCRIPT: { variant: "info" as const, label: t("trackerProperties.sources.script") },
    PRE_EXISTING: { variant: "outline" as const, label: t("trackerProperties.sources.preExisting") },
    HTTP: { variant: "neutral" as const, label: t("trackerProperties.sources.http") },
    EXTENSION: { variant: "warning" as const, label: t("trackerProperties.sources.extension") },
  };
  const typeBadge = typeBadges[pattern.trackerType]
    ?? { variant: "neutral" as const, label: pattern.trackerType };
  const sourceBadge = pattern.source
    ? sourceBadges[pattern.source]
    ?? { variant: "neutral" as const, label: pattern.source }
    : null;

  return (
    <Card padded>
      <PropertyRow label={t("trackerProperties.properties.pattern")}>
        <span className="font-mono text-sm">{pattern.pattern}</span>
      </PropertyRow>
      <PropertyRow label={t("trackerProperties.properties.matchType")}>
        <span className="text-sm">{pattern.matchType === "EXACT" ? t("trackerProperties.matchTypes.exact") : t("trackerProperties.matchTypes.glob")}</span>
      </PropertyRow>
      <PropertyRow label={t("trackerProperties.properties.type")}>
        <Badge variant={typeBadge.variant}>{typeBadge.label}</Badge>
      </PropertyRow>
      {pattern.source && (
        <PropertyRow label={t("trackerProperties.properties.source")}>
          <Badge variant={sourceBadge?.variant}>
            {sourceBadge?.label}
          </Badge>
        </PropertyRow>
      )}
      <PropertyRow label={t("trackerProperties.properties.category")}>
        <MoveToCategorySelect
          currentCategoryId={pattern.cookieCategory?.id}
          currentCategoryName={pattern.cookieCategory?.name}
          highlight={!!pattern.cookieCategory && pattern.cookieCategory.kind !== "UNCATEGORISED"}
          onSelect={handleMove}
        />
      </PropertyRow>
      <PropertyRow label={t("trackerProperties.properties.thirdParty")}>
        {pattern.thirdParty
          ? (
              <div className="flex items-center gap-2">
                <span className="text-sm">{pattern.thirdParty.name}</span>
              </div>
            )
          : pattern.commonThirdParty
            ? (
                <div className="flex items-center gap-2">
                  <Badge variant="info">{t("trackerProperties.commonCatalog")}</Badge>
                  <span className="text-sm">{pattern.commonThirdParty.name}</span>
                </div>
              )
            : <span className="text-txt-tertiary text-sm">-</span>}
      </PropertyRow>
      <PropertyRow label={t("trackerProperties.properties.maxAge")}>
        <span className="text-sm">
          {humanizeSeconds(pattern.maxAgeSeconds ?? null, t)}
        </span>
      </PropertyRow>
      {pattern.description && (
        <>
          <PropertyRow label={t("trackerProperties.properties.description")}>
            <span className="text-sm">{pattern.description}</span>
          </PropertyRow>
          <PropertyRow label={t("trackerProperties.properties.descriptionSource")}>
            {pattern.commonTrackerPatternId
              ? (
                  <div className="flex items-center gap-2">
                    <Badge variant="info">{t("trackerProperties.commonCatalog")}</Badge>
                    <span className="font-mono text-xs text-txt-tertiary">{pattern.commonTrackerPatternId}</span>
                    <button
                      type="button"
                      className="p-1 rounded hover:bg-bg-hover transition-colors cursor-pointer"
                      onClick={() => {
                        const commonTrackerPatternId = pattern.commonTrackerPatternId;
                        if (!commonTrackerPatternId) {
                          return;
                        }
                        void (async () => {
                          try {
                            await navigator.clipboard.writeText(commonTrackerPatternId);
                            toast({ title: t("trackerProperties.messages.copiedTitle"), description: t("trackerProperties.messages.idCopied"), variant: "success" });
                          } catch {
                            toast({ title: t("trackerProperties.errors.title"), description: t("trackerProperties.errors.copy"), variant: "error" });
                          }
                        })();
                      }}
                    >
                      <IconSquareBehindSquare2 size={16} />
                    </button>
                  </div>
                )
              : <Badge variant="neutral">{t("trackerProperties.manual")}</Badge>}
          </PropertyRow>
        </>
      )}
      <PropertyRow label={t("trackerProperties.properties.excluded")}>
        <span className="text-sm">{pattern.excluded ? t("trackerProperties.boolean.yes") : t("trackerProperties.boolean.no")}</span>
      </PropertyRow>
      <PropertyRow label={t("trackerProperties.properties.detectedCount")}>
        <span className="text-sm">{pattern.detectedCount}</span>
      </PropertyRow>
      <PropertyRow label={t("trackerProperties.properties.lastMatched")}>
        {pattern.lastMatchedAt
          ? (
              <time dateTime={pattern.lastMatchedAt}>
                {dateTimeFormat(i18n.language, pattern.lastMatchedAt)}
              </time>
            )
          : <span className="text-txt-tertiary">-</span>}
      </PropertyRow>
    </Card>
  );
}
