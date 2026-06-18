// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError, getTrackerSourceBadge, getTrackerTypeBadge, type GraphQLError, humanizeSeconds } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Card, IconSquareBehindSquare2, PropertyRow, useToast } from "@probo/ui";
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
  const { __ } = useTranslate();
  const pattern = useFragment(
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
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: __("Success"), description: __("Cookie moved"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to move cookie"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const typeBadge = getTrackerTypeBadge(pattern.trackerType, __);

  return (
    <Card padded>
      <PropertyRow label={__("Pattern")}>
        <span className="font-mono text-sm">{pattern.pattern}</span>
      </PropertyRow>
      <PropertyRow label={__("Match Type")}>
        <span className="text-sm">{pattern.matchType === "EXACT" ? __("Exact") : __("Glob")}</span>
      </PropertyRow>
      <PropertyRow label={__("Type")}>
        <Badge variant={typeBadge.variant}>{typeBadge.label}</Badge>
      </PropertyRow>
      {pattern.source && (
        <PropertyRow label={__("Source")}>
          <Badge variant={getTrackerSourceBadge(pattern.source, __).variant}>
            {getTrackerSourceBadge(pattern.source, __).label}
          </Badge>
        </PropertyRow>
      )}
      <PropertyRow label={__("Category")}>
        <MoveToCategorySelect
          currentCategoryId={pattern.cookieCategory?.id}
          currentCategoryName={pattern.cookieCategory?.name}
          highlight={!!pattern.cookieCategory && pattern.cookieCategory.kind !== "UNCATEGORISED"}
          onSelect={handleMove}
        />
      </PropertyRow>
      <PropertyRow label={__("Third party")}>
        {pattern.thirdParty
          ? (
              <div className="flex items-center gap-2">
                <span className="text-sm">{pattern.thirdParty.name}</span>
              </div>
            )
          : pattern.commonThirdParty
            ? (
                <div className="flex items-center gap-2">
                  <Badge variant="info">{__("Common catalog")}</Badge>
                  <span className="text-sm">{pattern.commonThirdParty.name}</span>
                </div>
              )
            : <span className="text-txt-tertiary text-sm">-</span>}
      </PropertyRow>
      <PropertyRow label={__("Max Age")}>
        <span className="text-sm">
          {humanizeSeconds(pattern.maxAgeSeconds ?? null)}
        </span>
      </PropertyRow>
      {pattern.description && (
        <>
          <PropertyRow label={__("Description")}>
            <span className="text-sm">{pattern.description}</span>
          </PropertyRow>
          <PropertyRow label={__("Description source")}>
            {pattern.commonTrackerPatternId
              ? (
                  <div className="flex items-center gap-2">
                    <Badge variant="info">{__("Common catalog")}</Badge>
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
                            toast({ title: __("Copied"), description: __("Common Tracker ID copied to clipboard"), variant: "success" });
                          } catch {
                            toast({ title: __("Error"), description: __("Failed to copy Common Tracker ID"), variant: "error" });
                          }
                        })();
                      }}
                    >
                      <IconSquareBehindSquare2 size={16} />
                    </button>
                  </div>
                )
              : <Badge variant="neutral">{__("Manual")}</Badge>}
          </PropertyRow>
        </>
      )}
      <PropertyRow label={__("Excluded")}>
        <span className="text-sm">{pattern.excluded ? __("Yes") : __("No")}</span>
      </PropertyRow>
      <PropertyRow label={__("Distinct Trackers Detected")}>
        <span className="text-sm">{pattern.detectedCount}</span>
      </PropertyRow>
      <PropertyRow label={__("Last Matched")}>
        {pattern.lastMatchedAt
          ? (
              <time dateTime={pattern.lastMatchedAt}>
                {new Date(pattern.lastMatchedAt).toLocaleString()}
              </time>
            )
          : <span className="text-txt-tertiary">-</span>}
      </PropertyRow>
    </Card>
  );
}
