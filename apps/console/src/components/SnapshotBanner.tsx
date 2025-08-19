import { IconClock } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useLazyLoadQuery, graphql } from "react-relay";
import type { SnapshotBannerQuery } from "./__generated__/SnapshotBannerQuery.graphql";

const snapshotQuery = graphql`
  query SnapshotBannerQuery($snapshotId: ID!) {
    node(id: $snapshotId) {
      ... on Snapshot {
        id
        name
        createdAt
      }
    }
  }
`;

type Props = {
  snapshotId: string;
};

export function SnapshotBanner({ snapshotId }: Props) {
  const { __, dateFormat } = useTranslate();

  const data = useLazyLoadQuery<SnapshotBannerQuery>(snapshotQuery, { snapshotId });
  const snapshot = data.node;

  if (!snapshot) {
    return null;
  }

  return (
    <div className="bg-warning rounded-lg p-4 flex items-center gap-3">
      <IconClock className="text-warning-600 flex-shrink-0" size={20} />
      <div className="flex-1">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-medium text-warning-800">{__("Snapshot")} {snapshot.name}</span>
        </div>
        <p className="text-sm text-warning-700">
          {__("You are viewing a snapshot of data from")} {dateFormat(snapshot.createdAt, { year: "numeric", month: "short", day: "numeric" })}
        </p>
      </div>
    </div>
  );
}
