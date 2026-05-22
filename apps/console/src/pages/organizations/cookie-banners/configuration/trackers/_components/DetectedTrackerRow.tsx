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
import { Badge, Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { DetectedTrackerRow_detectedTracker$key } from "#/__generated__/core/DetectedTrackerRow_detectedTracker.graphql";

const detectedTrackerFragment = graphql`
  fragment DetectedTrackerRow_detectedTracker on DetectedTracker {
    id
    identifier
    initiatorUrl
    maxAgeSeconds
    source
    lastDetectedAt
  }
`;

function sourceBadge(source: string, __: (s: string) => string) {
  switch (source) {
    case "SCRIPT": return { label: __("Script"), variant: "info" as const };
    case "PRE_EXISTING": return { label: __("Pre-existing"), variant: "outline" as const };
    default: return { label: source, variant: "neutral" as const };
  }
}

interface DetectedTrackerRowProps {
  detectedTrackerKey: DetectedTrackerRow_detectedTracker$key;
}

export function DetectedTrackerRow({ detectedTrackerKey }: DetectedTrackerRowProps) {
  const { __ } = useTranslate();
  const tracker = useFragment(detectedTrackerFragment, detectedTrackerKey);

  return (
    <Tr>
      <Td>
        <span className="font-mono text-xs break-all max-w-xs inline-block">{tracker.identifier}</span>
      </Td>
      <Td>
        {tracker.initiatorUrl
          ? <span className="font-mono text-xs break-all max-w-xs inline-block">{tracker.initiatorUrl}</span>
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        {tracker.maxAgeSeconds != null
          ? <span className="text-sm">{tracker.maxAgeSeconds}</span>
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        {tracker.source
          ? (
              <Badge variant={sourceBadge(tracker.source, __).variant}>
                {sourceBadge(tracker.source, __).label}
              </Badge>
            )
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        <time dateTime={tracker.lastDetectedAt}>
          {new Date(tracker.lastDetectedAt).toLocaleString()}
        </time>
      </Td>
    </Tr>
  );
}
