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

import { faviconUrl, formatDate } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  RiskBadge,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type { ThirdPartyGraphListQuery } from "#/__generated__/core/ThirdPartyGraphListQuery.graphql";
import type {
  ThirdPartyGraphPaginatedFragment$data,
  ThirdPartyGraphPaginatedFragment$key,
} from "#/__generated__/core/ThirdPartyGraphPaginatedFragment.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import {
  paginatedThirdPartiesFragment,
  useDeleteThirdParty,
  thirdPartiesQuery,
} from "#/hooks/graph/ThirdPartyGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { CreateThirdPartyDialog } from "./dialogs/CreateThirdPartyDialog";

type ThirdParty = NodeOf<ThirdPartyGraphPaginatedFragment$data["third parties"]>;

type Props = {
  queryRef: PreloadedQuery<ThirdPartyGraphListQuery>;
};

export default function ThirdPartiesPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const data = usePreloadedQuery(thirdPartiesQuery, props.queryRef);
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    paginatedThirdPartiesFragment,
    data.node as ThirdPartyGraphPaginatedFragment$key,
  );

  const thirdParties = pagination.data.thirdParties?.edges.map(edge => edge.node);
  const connectionId = pagination.data.thirdParties.__id;

  usePageTitle(__("Third Parties"));

  const hasAnyAction
    = !isSnapshotMode
      && thirdParties.some(({ canUpdate, canDelete }) => canUpdate || canDelete);

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Third Parties")}
        description={__(
          "Third parties are third-party services that your company uses. Add them to keep track of their risk and compliance status.",
        )}
      >
        {!isSnapshotMode && data.node.canCreateThirdParty && (
          <CreateThirdPartyDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add third party")}</Button>
          </CreateThirdPartyDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Third Party")}</SortableTh>
            <Th>{__("Accessed At")}</Th>
            <Th>{__("Data Risk")}</Th>
            <Th>{__("Business Risk")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {thirdParties?.map(thirdParty => (
            <ThirdPartyRow
              key={thirdParty.id}
              thirdParty={thirdParty}
              organizationId={organizationId}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function ThirdPartyRow({
  thirdParty,
  organizationId,
  connectionId,
  hasAnyAction,
}: {
  thirdParty: ThirdParty;
  organizationId: string;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const { __ } = useTranslate();
  const latestAssessment = thirdParty.riskAssessments?.edges[0]?.node;
  const deleteThirdParty = useDeleteThirdParty(thirdParty, connectionId);

  const thirdPartyUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/third-parties/${thirdParty.id}/overview`
      : `/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`;

  return (
    <>
      <Tr to={thirdPartyUrl}>
        <Td>
          <div className="flex gap-2 items-center">
            <Avatar name={thirdParty.name} src={faviconUrl(thirdParty.websiteUrl)} />
            <div>{thirdParty.name}</div>
          </div>
        </Td>
        <Td>
          {latestAssessment?.createdAt
            ? formatDate(latestAssessment.createdAt)
            : __("Not assessed")}
        </Td>
        <Td>
          <RiskBadge level={latestAssessment?.dataSensitivity ?? "NONE"} />
        </Td>
        <Td>
          <RiskBadge level={latestAssessment?.businessImpact ?? "NONE"} />
        </Td>
        {hasAnyAction && (
          <Td noLink width={50} className="text-end">
            <ActionDropdown>
              {thirdParty.canDelete && (
                <DropdownItem
                  onClick={deleteThirdParty}
                  variant="danger"
                  icon={IconTrashCan}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          </Td>
        )}
      </Tr>
    </>
  );
}
