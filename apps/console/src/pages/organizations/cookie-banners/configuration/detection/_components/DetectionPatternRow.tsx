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

import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { DetectionPatternRowFragment$key } from "#/__generated__/core/DetectionPatternRowFragment.graphql";

const detectionPatternFragment = graphql`
  fragment DetectionPatternRowFragment on CookiePattern {
    displayName
    matchType
    source
    description
    lastMatchedAt
    updatedAt
  }
`;

interface DetectionPatternRowProps {
  patternKey: DetectionPatternRowFragment$key;
}

export function DetectionPatternRow({ patternKey }: DetectionPatternRowProps) {
  const { __ } = useTranslate();
  const pattern = useFragment(detectionPatternFragment, patternKey);

  return (
    <Tr>
      <Td>
        <div className="flex flex-col min-w-0">
          <span className="font-medium">{pattern.displayName}</span>
          {pattern.description && (
            <span className="text-xs text-txt-tertiary wrap-break-word line-clamp-1">
              {pattern.description}
            </span>
          )}
        </div>
      </Td>
      <Td>
        <Badge variant={pattern.source === "SCRIPT" ? "info" : "neutral"}>
          {pattern.source === "SCRIPT" ? __("Script") : __("Pre-existing")}
        </Badge>
      </Td>
      <Td>
        {pattern.lastMatchedAt
          ? (
              <time dateTime={pattern.lastMatchedAt}>
                {formatDate(pattern.lastMatchedAt)}
              </time>
            )
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        <time dateTime={pattern.updatedAt}>
          {formatDate(pattern.updatedAt)}
        </time>
      </Td>
    </Tr>
  );
}
