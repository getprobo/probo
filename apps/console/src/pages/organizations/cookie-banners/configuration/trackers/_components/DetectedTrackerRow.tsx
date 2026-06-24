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

import { getTrackerSourceBadge } from "@probo/helpers";
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
              <Badge variant={getTrackerSourceBadge(tracker.source, __).variant}>
                {getTrackerSourceBadge(tracker.source, __).label}
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
