import { getSnapshotTypeUrlPath } from "@probo/helpers";
import { useEffect } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate, useParams } from "react-router";

import type { SnapshotGraphNodeQuery } from "#/__generated__/core/SnapshotGraphNodeQuery.graphql";
import { PageError } from "#/components/PageError";
import { snapshotNodeQuery } from "#/hooks/graph/SnapshotGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  queryRef: PreloadedQuery<SnapshotGraphNodeQuery>;
};

export default function SnapshotDetailPage({ queryRef }: Props) {
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams();
  const data = usePreloadedQuery(snapshotNodeQuery, queryRef);

  useEffect(() => {
    if (!data.node || !data.node.type) {
      return;
    }

    const snapshot = data.node;
    const snapshotType = snapshot.type;
    const urlPath = getSnapshotTypeUrlPath(snapshotType);

    void navigate(
      `/organizations/${organizationId}/snapshots/${snapshotId}${urlPath}`,
      {
        replace: true,
      },
    );
  }, [data.node, navigate, organizationId, snapshotId]);

  if (!data.node || !data.node.type) {
    return <PageError />;
  }

  return null;
}
