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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  PageHeader,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type { AssetGraphListQuery } from "#/__generated__/core/AssetGraphListQuery.graphql";
import type { AssetsListQuery } from "#/__generated__/core/AssetsListQuery.graphql";
import type { AssetsPageFragment$key } from "#/__generated__/core/AssetsPageFragment.graphql";
import { assetsQuery } from "#/hooks/graph/AssetGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AssetsTable } from "../../../components/assets/AssetsTable";
import { ReadOnlyAssetsTable } from "../../../components/assets/ReadOnlyAssetsTable";

import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import { PublishAssetListDialog } from "./dialogs/PublishAssetListDialog";

const paginatedAssetsFragment = graphql`
  fragment AssetsPageFragment on Organization
  @refetchable(queryName: "AssetsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "AssetOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    assets(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "AssetsPage_assets") {
      __id
      edges {
        node {
          # eslint-disable-next-line relay/unused-fields
          id
          # eslint-disable-next-line relay/unused-fields
          name
          # eslint-disable-next-line relay/unused-fields
          amount
          # eslint-disable-next-line relay/unused-fields
          assetType
          # eslint-disable-next-line relay/unused-fields
          dataTypesStored
          # eslint-disable-next-line relay/unused-fields
          owner {
            id
            # eslint-disable-next-line relay/unused-fields
            fullName
          }
          # eslint-disable-next-line relay/unused-fields
          vendors(first: 50) {
            edges {
              node {
                # eslint-disable-next-line relay/unused-fields
                id
                # eslint-disable-next-line relay/unused-fields
                name
                # eslint-disable-next-line relay/unused-fields
                websiteUrl
              }
            }
          }
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
  const navigate = useNavigate();

  const data = usePreloadedQuery<AssetGraphListQuery>(
    assetsQuery,
    props.queryRef,
  );
  const pagination = usePaginationFragment<AssetsListQuery, AssetsPageFragment$key>(
    paginatedAssetsFragment,
    data.node as AssetsPageFragment$key,
  );
  const assets = pagination.data.assets?.edges.map(edge => edge.node);
  const connectionId = pagination.data.assets.__id;
  const defaultApproverIds = (data.node.assetListDocument?.defaultApprovers ?? []).map(a => a.id);

  const canWrite = assets.some(asset => asset.canDelete || asset.canUpdate);
  usePageTitle(__("Assets"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Assets")}
        description={__(
          "Manage your organization's assets and their classifications.",
        )}
      >
        <div className="flex gap-2">
          {data.node.assetListDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${data.node.assetListDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {__("Document")}
              </Link>
            </Button>
          )}
          {data.node.canPublishAssets && (
            <PublishAssetListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {__("Publish")}
              </Button>
            </PublishAssetListDialog>
          )}
          {data.node.canCreateAsset && (
            <CreateAssetDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>{__("Add asset")}</Button>
            </CreateAssetDialog>
          )}
        </div>
      </PageHeader>
      {!canWrite
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
