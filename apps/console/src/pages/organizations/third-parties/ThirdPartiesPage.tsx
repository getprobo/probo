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
  IconPageTextLine,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
  PageHeader,
  RiskBadge,
  TabItem,
  Tabs,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useState, useTransition } from "react";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { ThirdPartyGraphListQuery } from "#/__generated__/core/ThirdPartyGraphListQuery.graphql";
import type {
  ThirdPartyGraphPaginatedFragment$data,
  ThirdPartyGraphPaginatedFragment$key,
} from "#/__generated__/core/ThirdPartyGraphPaginatedFragment.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import {
  paginatedThirdPartiesFragment,
  thirdPartiesQuery,
  useDeleteThirdParty,
} from "#/hooks/graph/ThirdPartyGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { CreateThirdPartyDialog } from "./dialogs/CreateThirdPartyDialog";
import { PublishThirdPartyListDialog } from "./dialogs/PublishThirdPartyListDialog";

type ThirdParty = NodeOf<ThirdPartyGraphPaginatedFragment$data["thirdParties"]>;

type Props = {
  queryRef: PreloadedQuery<ThirdPartyGraphListQuery>;
};

export default function ThirdPartiesPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const data = usePreloadedQuery(thirdPartiesQuery, props.queryRef);
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    paginatedThirdPartiesFragment,
    data.node as ThirdPartyGraphPaginatedFragment$key,
  );

  const thirdParties = pagination.data.thirdParties?.edges.map(edge => edge.node);
  const connectionId = pagination.data.thirdParties.__id;
  const [, startTransition] = useTransition();
  const [firstLevelFilter, setFirstLevelFilter] = useState<boolean | null>(true);

  usePageTitle(__("Third parties"));

  const handleFilterChange = (firstLevel: boolean | null) => {
    setFirstLevelFilter(firstLevel);
    startTransition(() => {
      pagination.refetch(
        { filter: firstLevel !== null ? { firstLevel } : {} },
        { fetchPolicy: "store-and-network" },
      );
    });
  };

  const hasAnyAction
    = thirdParties.some(({ canUpdate, canDelete }) => canUpdate || canDelete);

  const thirdPartiesDocument = data.node?.thirdPartiesDocument;
  const defaultApproverIds
    = thirdPartiesDocument?.defaultApprovers?.map(a => a.id) ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Third parties")}
        description={__(
          "Third parties are external services and providers that your company uses. Add them to keep track of their risk and compliance status.",
        )}
      >
        <div className="flex gap-2">
          {thirdPartiesDocument && (
            <Button
              variant="secondary"
              icon={IconPageTextLine}
              onClick={() => void navigate(
                `/organizations/${organizationId}/documents/${thirdPartiesDocument.id}`,
              )}
            >
              {__("Document")}
            </Button>
          )}
          {data.node.canPublishThirdParty && (
            <PublishThirdPartyListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={documentId => void navigate(
                `/organizations/${organizationId}/documents/${documentId}`,
              )}
            >
              <Button variant="secondary" icon={IconUpload}>
                {__("Publish")}
              </Button>
            </PublishThirdPartyListDialog>
          )}
          {data.node.canCreateThirdParty && (
            <CreateThirdPartyDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>{__("Add third party")}</Button>
            </CreateThirdPartyDialog>
          )}
        </div>
      </PageHeader>
      <Tabs>
        <TabItem
          active={firstLevelFilter === true}
          onClick={() => handleFilterChange(true)}
        >
          {__("First Level")}
        </TabItem>
        <TabItem
          active={firstLevelFilter === null}
          onClick={() => handleFilterChange(null)}
        >
          {__("All")}
        </TabItem>
      </Tabs>
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Third party")}</SortableTh>
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
  const { __ } = useTranslate();
  const latestAssessment = thirdParty.riskAssessments?.edges[0]?.node;
  const deleteThirdParty = useDeleteThirdParty(thirdParty, connectionId);

  const thirdPartyUrl = `/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`;

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
