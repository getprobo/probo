// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { usePageTitle } from "@probo/hooks";
import {
  Button,
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  PageHeader,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
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
          thirdParties(first: 50) {
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
  const { t } = useTranslation();
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
  usePageTitle(t("assetsPage.title"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("assetsPage.title")}
        description={t("assetsPage.description")}
      >
        <div className="flex gap-2">
          {data.node.assetListDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${data.node.assetListDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("assetsPage.actions.document")}
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
                {t("assetsPage.actions.publish")}
              </Button>
            </PublishAssetListDialog>
          )}
          {data.node.canCreateAsset && (
            <CreateAssetDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>
                {t("assetsPage.actions.addAsset")}
              </Button>
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
