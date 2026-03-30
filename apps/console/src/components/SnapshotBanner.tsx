// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import {
  formatDate,
  getSnapshotTypeLabel,
  getSnapshotTypeUrlPath,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { IconClock } from "@probo/ui";
import { graphql, useLazyLoadQuery } from "react-relay";
import { useLocation } from "react-router";

import type { SnapshotBannerQuery } from "#/__generated__/core/SnapshotBannerQuery.graphql";

const snapshotQuery = graphql`
  query SnapshotBannerQuery($snapshotId: ID!) {
    node(id: $snapshotId) {
      ... on Snapshot {
        # eslint-disable-next-line relay/unused-fields
        id
        name
        type
        createdAt
      }
    }
  }
`;

const isSnapshotTypeValidForUrl = (type: string, pathname: string) => {
  const urlPath = getSnapshotTypeUrlPath(type);
  return pathname.includes(urlPath);
};

type Props = {
  snapshotId: string;
};

export function SnapshotBanner({ snapshotId }: Props) {
  const { __ } = useTranslate();
  const location = useLocation();

  const data = useLazyLoadQuery<SnapshotBannerQuery>(snapshotQuery, {
    snapshotId,
  });
  const snapshot = data.node;

  if (!snapshot) {
    return null;
  }

  if (
    snapshot.type
    && !isSnapshotTypeValidForUrl(snapshot.type, location.pathname)
  ) {
    throw new Error("PAGE_NOT_FOUND");
  }

  return (
    <div className="bg-warning rounded-lg p-4 flex items-center gap-3">
      <IconClock className="text-warning-600 flex-shrink-0" size={20} />
      <div className="flex-1">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-medium text-warning-800">
            {__("Snapshot")}
            {" "}
            {snapshot.name}
          </span>
        </div>
        <p className="text-sm text-warning-700">
          {sprintf(
            __("You are viewing a %s snapshot from %s"),
            getSnapshotTypeLabel(__, snapshot.type).toLocaleLowerCase(),
            formatDate(snapshot.createdAt),
          )}
        </p>
      </div>
    </div>
  );
}
