import { useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { snapshotNodeQuery } from "/hooks/graph/SnapshotGraph";
import type { SnapshotGraphNodeQuery } from "/hooks/graph/__generated__/SnapshotGraphNodeQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PageError } from "/components/PageError";
import { getSnapshotTypeUrlPath } from "@probo/helpers";

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

    navigate(`/organizations/${organizationId}/snapshots/${snapshotId}${urlPath}`, {
      replace: true,
    });
  }, [data.node, navigate, organizationId, snapshotId]);

  if (!data.node || !data.node.type) {
    return <PageError />;
  }

  return null;
}
