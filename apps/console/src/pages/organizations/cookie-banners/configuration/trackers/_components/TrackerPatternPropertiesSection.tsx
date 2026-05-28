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

import { humanizeSeconds } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Card, PropertyRow } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { TrackerPatternPropertiesSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternPropertiesSection_trackerPattern.graphql";

const trackerPatternPropertiesSectionFragment = graphql`
  fragment TrackerPatternPropertiesSection_trackerPattern on TrackerPattern {
    pattern
    matchType
    trackerType
    source
    maxAgeSeconds
    description
    excluded
    detectedCount
    lastMatchedAt
    cookieCategory {
      name
    }
    thirdParty {
      id
      name
    }
    commonThirdParty {
      id
      name
      logoUrl
    }
  }
`;

function trackerTypeBadge(type: string, __: (s: string) => string) {
  switch (type) {
    case "COOKIE": return { label: __("Cookie"), variant: "warning" as const };
    case "LOCAL_STORAGE": return { label: __("localStorage"), variant: "info" as const };
    case "SESSION_STORAGE": return { label: __("sessionStorage"), variant: "highlight" as const };
    case "INDEXED_DB": return { label: __("IndexedDB"), variant: "success" as const };
    case "CACHE_STORAGE": return { label: __("Cache Storage"), variant: "outline" as const };
    default: return { label: type, variant: "neutral" as const };
  }
}

function sourceBadge(source: string, __: (s: string) => string) {
  switch (source) {
    case "SCRIPT": return { label: __("Script"), variant: "info" as const };
    case "PRE_EXISTING": return { label: __("Pre-existing"), variant: "outline" as const };
    default: return { label: source, variant: "neutral" as const };
  }
}

interface TrackerPatternPropertiesSectionProps {
  trackerPatternKey: TrackerPatternPropertiesSection_trackerPattern$key;
}

export function TrackerPatternPropertiesSection({
  trackerPatternKey,
}: TrackerPatternPropertiesSectionProps) {
  const { __ } = useTranslate();
  const pattern = useFragment(
    trackerPatternPropertiesSectionFragment,
    trackerPatternKey,
  );

  const typeBadge = trackerTypeBadge(pattern.trackerType, __);

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
          <Badge variant={sourceBadge(pattern.source, __).variant}>
            {sourceBadge(pattern.source, __).label}
          </Badge>
        </PropertyRow>
      )}
      <PropertyRow label={__("Category")}>
        <span className="text-sm">
          {pattern.cookieCategory?.name ?? "-"}
        </span>
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
        <PropertyRow label={__("Description")}>
          <span className="text-sm">{pattern.description}</span>
        </PropertyRow>
      )}
      <PropertyRow label={__("Excluded")}>
        <span className="text-sm">{pattern.excluded ? __("Yes") : __("No")}</span>
      </PropertyRow>
      <PropertyRow label={__("Detected Count")}>
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
