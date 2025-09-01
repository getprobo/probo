import { IconClock } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useLazyLoadQuery, graphql } from "react-relay";
import { useLocation } from "react-router";
import type { SnapshotBannerQuery } from "./__generated__/SnapshotBannerQuery.graphql";
import { getSnapshotTypeUrlPath } from "@probo/helpers";

const snapshotQuery = graphql`
  query SnapshotBannerQuery($snapshotId: ID!) {
    node(id: $snapshotId) {
      ... on Snapshot {
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
  const { __, dateFormat } = useTranslate();
  const location = useLocation();

  const data = useLazyLoadQuery<SnapshotBannerQuery>(snapshotQuery, { snapshotId });
  const snapshot = data.node;

  if (!snapshot) {
    return null;
  }

  if (snapshot.type && !isSnapshotTypeValidForUrl(snapshot.type, location.pathname)) {
    throw new Error("PAGE_NOT_FOUND");
  }

  return (
    <div className="bg-warning rounded-lg p-4 flex items-center gap-3">
      <IconClock className="text-warning-600 flex-shrink-0" size={20} />
      <div className="flex-1">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-medium text-warning-800">{__("Snapshot")} {snapshot.name}</span>
        </div>
        <p className="text-sm text-warning-700">
          {__("You are viewing a snapshot from")} {dateFormat(snapshot.createdAt, { year: "numeric", month: "short", day: "numeric" })}
        </p>
      </div>
    </div>
  );
}
