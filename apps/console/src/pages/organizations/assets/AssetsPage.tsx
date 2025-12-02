import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconPlusLarge,
  PageHeader,
} from "@probo/ui";
import { use } from "react";
import {
  graphql,
  usePaginationFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import { AssetsTable } from "../../../components/assets/AssetsTable";
import { ReadOnlyAssetsTable } from "../../../components/assets/ReadOnlyAssetsTable";
import type { AssetsPageFragment$key } from "./__generated__/AssetsPageFragment.graphql";
import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import { SnapshotBanner } from "/components/SnapshotBanner";
import {
  assetsQuery,
} from "/hooks/graph/AssetGraph";
import type { AssetGraphListQuery } from "/hooks/graph/__generated__/AssetGraphListQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PermissionsContext } from "/providers/PermissionsContext";

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

  const data = usePreloadedQuery<AssetGraphListQuery>(assetsQuery, props.queryRef);
  const pagination = usePaginationFragment(
    paginatedAssetsFragment,
    data.node as AssetsPageFragment$key,
  );
  const assets = pagination.data.assets?.edges.map((edge) => edge.node);
  const connectionId = pagination.data.assets.__id;

  const { isAuthorized } = use(PermissionsContext);
  const canWrite = (
    isAuthorized("Asset", "updateAsset") ||
    isAuthorized("Asset", "deleteAsset")
  );
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
        {!isSnapshotMode && isAuthorized("Organization", "createAsset") && (
          <CreateAssetDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add asset")}</Button>
          </CreateAssetDialog>
        )}
      </PageHeader>
      {isSnapshotMode || !canWrite ?
        <ReadOnlyAssetsTable pagination={pagination} assets={assets} />
        :
        <AssetsTable
          connectionId={connectionId}
          pagination={pagination}
          assets={assets}
        />
      }
    </div>
  );
}
