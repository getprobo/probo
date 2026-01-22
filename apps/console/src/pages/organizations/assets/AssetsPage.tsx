import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, PageHeader } from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type { AssetGraphListQuery } from "#/__generated__/core/AssetGraphListQuery.graphql";
import type { AssetsPageFragment$key } from "#/__generated__/core/AssetsPageFragment.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { assetsQuery } from "#/hooks/graph/AssetGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AssetsTable } from "../../../components/assets/AssetsTable";
import { ReadOnlyAssetsTable } from "../../../components/assets/ReadOnlyAssetsTable";

import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";

const paginatedAssetsFragment = graphql`
  fragment AssetsPageFragment on Organization
  @refetchable(queryName: "AssetsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "AssetOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    assets(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
      filter: { snapshotId: $snapshotId }
    ) @connection(key: "AssetsPage_assets", filters: ["filter"]) {
      __id
      edges {
        node {
          id
          snapshotId
          name
          amount
          assetType
          dataTypesStored
          owner {
            id
            fullName
          }
          vendors(first: 50) {
            edges {
              node {
                id
                name
                websiteUrl
              }
            }
          }
          createdAt
          canUpdate: permission(action: "core:asset:update")
          canDelete: permission(action: "core:asset:delete")
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<AssetGraphListQuery>;
};

export default function AssetsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const data = usePreloadedQuery<AssetGraphListQuery>(
    assetsQuery,
    props.queryRef,
  );
  const pagination = usePaginationFragment(
    paginatedAssetsFragment,
    data.node as AssetsPageFragment$key,
  );
  const assets = pagination.data.assets?.edges.map(edge => edge.node);
  const connectionId = pagination.data.assets.__id;

  const canWrite = assets.some(asset => asset.canDelete || asset.canUpdate);
  usePageTitle(__("Assets"));

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Assets")}
        description={__(
          "Manage your organization's assets and their classifications.",
        )}
      >
        {!isSnapshotMode && data.node.canCreateAsset && (
          <CreateAssetDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add asset")}</Button>
          </CreateAssetDialog>
        )}
      </PageHeader>
      {isSnapshotMode || !canWrite
        ? (
            <ReadOnlyAssetsTable pagination={pagination} assets={assets} />
          )
        : (
            <AssetsTable
              connectionId={connectionId}
              pagination={pagination}
              assets={assets}
            />
          )}
    </div>
  );
}
