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

import type { RegistryAssetGraphListQuery } from "#/__generated__/core/RegistryAssetGraphListQuery.graphql";
import type { RegistryAssetsListQuery } from "#/__generated__/core/RegistryAssetsListQuery.graphql";
import type { RegistryAssetsPageFragment$key } from "#/__generated__/core/RegistryAssetsPageFragment.graphql";
import { registryAssetsQuery } from "#/hooks/graph/RegistryAssetGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { RegistryAssetsTable } from "../../../components/registry-assets/RegistryAssetsTable";
import { ReadOnlyRegistryAssetsTable } from "../../../components/registry-assets/ReadOnlyRegistryAssetsTable";

import { CreateRegistryAssetDialog } from "./dialogs/CreateRegistryAssetDialog";
import { PublishRegistryAssetListDialog } from "./dialogs/PublishRegistryAssetListDialog";

const paginatedAssetsFragment = graphql`
  fragment RegistryAssetsPageFragment on Organization
  @refetchable(queryName: "RegistryAssetsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "RegistryAssetOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    registryAssets(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "RegistryAssetsPage_registryAssets") {
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
          registryAssetType
          # eslint-disable-next-line relay/unused-fields
          dataTypesStored
          # eslint-disable-next-line relay/unused-fields
          owner {
            id
            # eslint-disable-next-line relay/unused-fields
            fullName
            # eslint-disable-next-line relay/unused-fields
            emailAddress
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
          canUpdate: permission(action: "core:registry-asset:update")
          canDelete: permission(action: "core:registry-asset:delete")
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<RegistryAssetGraphListQuery>;
};

export default function RegistryAssetsPage(props: Props) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const data = usePreloadedQuery<RegistryAssetGraphListQuery>(
    registryAssetsQuery,
    props.queryRef,
  );
  const pagination = usePaginationFragment<RegistryAssetsListQuery, RegistryAssetsPageFragment$key>(
    paginatedAssetsFragment,
    data.node as RegistryAssetsPageFragment$key,
  );
  const assets = pagination.data.registryAssets?.edges.map(edge => edge.node);
  const connectionId = pagination.data.registryAssets.__id;
  const defaultApproverIds = (data.node.registryAssetListDocument?.defaultApprovers ?? []).map(a => a.id);

  const canWrite = assets.some(asset => asset.canDelete || asset.canUpdate);
  usePageTitle(t("assetsPage.title"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("assetsPage.title")}
        description={t("assetsPage.description")}
      >
        <div className="flex gap-2">
          {data.node.registryAssetListDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/governance/documents/${data.node.registryAssetListDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("assetsPage.actions.document")}
              </Link>
            </Button>
          )}
          {data.node.canPublishAssets && (
            <PublishRegistryAssetListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/governance/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {t("assetsPage.actions.publish")}
              </Button>
            </PublishRegistryAssetListDialog>
          )}
          {data.node.canCreateRegistryAsset && (
            <CreateRegistryAssetDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>
                {t("assetsPage.actions.addAsset")}
              </Button>
            </CreateRegistryAssetDialog>
          )}
        </div>
      </PageHeader>
      {!canWrite
        ? (
            <ReadOnlyRegistryAssetsTable pagination={pagination} assets={assets} />
          )
        : (
            <RegistryAssetsTable
              connectionId={connectionId}
              pagination={pagination}
              assets={assets}
            />
          )}
    </div>
  );
}
