// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  Avatar,
  Button,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  RiskBadge,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import type { ComponentProps } from "react";
import {
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { ThirdPartyThirdPartiesPageDeleteMutation } from "#/__generated__/core/ThirdPartyThirdPartiesPageDeleteMutation.graphql";
import type { ThirdPartyThirdPartiesPageFragment$key } from "#/__generated__/core/ThirdPartyThirdPartiesPageFragment.graphql";
import type { ThirdPartyThirdPartiesPagePaginationQuery } from "#/__generated__/core/ThirdPartyThirdPartiesPagePaginationQuery.graphql";
import type { ThirdPartyThirdPartiesPageQuery } from "#/__generated__/core/ThirdPartyThirdPartiesPageQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AddChildThirdPartyDialog } from "../_components/AddChildThirdPartyDialog";

// Keep in sync with coredata.MaxThirdPartyLevel on the backend.
const MAX_THIRD_PARTY_LEVEL = 4;

export const thirdPartyThirdPartiesPageQuery = graphql`
  query ThirdPartyThirdPartiesPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        name
        level
        ancestors {
          name
        }
        canUpdate: permission(action: "core:thirdParty:update")
        ...ThirdPartyThirdPartiesPageFragment
      }
    }
  }
`;

const paginatedFragment = graphql`
  fragment ThirdPartyThirdPartiesPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyThirdPartiesPagePaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    childThirdParties(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyThirdPartiesPageFragment_childThirdParties", filters: []) {
      __id
      edges {
        node {
          id
          name
          websiteUrl
          riskAssessments(
            first: 1
            orderBy: { direction: DESC, field: CREATED_AT }
          ) {
            edges {
              node {
                id
                createdAt
                dataSensitivity
                businessImpact
              }
            }
          }
        }
      }
    }
  }
`;

const deleteChildMutation = graphql`
  mutation ThirdPartyThirdPartiesPageDeleteMutation(
    $input: DeleteThirdPartyInput!
    $connections: [ID!]!
  ) {
    deleteThirdParty(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

interface Props {
  queryRef: PreloadedQuery<ThirdPartyThirdPartiesPageQuery>;
}

export default function ThirdPartyThirdPartiesPage({ queryRef }: Props) {
  const { node } = usePreloadedQuery(thirdPartyThirdPartiesPageQuery, queryRef);
  const thirdParty = node.__typename === "ThirdParty" ? node : null;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();

  const pagination = usePaginationFragment<
    ThirdPartyThirdPartiesPagePaginationQuery,
    ThirdPartyThirdPartiesPageFragment$key
  >(paginatedFragment, thirdParty as ThirdPartyThirdPartiesPageFragment$key);
  const [deleteChild] = useMutation<ThirdPartyThirdPartiesPageDeleteMutation>(deleteChildMutation);

  usePageTitle((thirdParty?.name ?? "") + " - " + __("Third Parties"));

  if (!thirdParty) {
    return null;
  }

  const connectionId = pagination.data.childThirdParties.__id;
  const childThirdParties = pagination.data.childThirdParties.edges.map(edge => edge.node);

  // Strip any existing " (...)" suffix so the chain stays clean even when a
  // parent's stored name is itself already qualified.
  const baseName = (name: string) => name.replace(/\s*\([^)]*\)\s*$/, "");
  const parentAncestors = thirdParty.ancestors ?? [];
  const parentNamePath = [
    ...parentAncestors.map(ancestor => baseName(ancestor.name)),
    baseName(thirdParty.name),
  ];

  const handleDelete = (childId: string, childName: string) => {
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          deleteChild({
            variables: {
              input: { thirdPartyId: childId },
              connections: [connectionId],
            },
            onCompleted: () => resolve(),
            onError: err => reject(err),
          });
        }),
      {
        message: `${__("Delete")} "${childName}"?`,
      },
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Third Parties")}
        description={__("Manage third parties linked to this third party.")}
      >
        {thirdParty.canUpdate && thirdParty.level < MAX_THIRD_PARTY_LEVEL && (
          <AddChildThirdPartyDialog
            parentThirdPartyId={thirdParty.id}
            parentNamePath={parentNamePath}
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>{__("Add third party")}</Button>
          </AddChildThirdPartyDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={pagination.refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Third party")}</SortableTh>
            <Th>{__("Accessed At")}</Th>
            <Th>{__("Data Risk")}</Th>
            <Th>{__("Business Risk")}</Th>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {childThirdParties.map((child) => {
            const latestAssessment = child.riskAssessments?.edges[0]?.node;

            return (
              <Tr
                key={child.id}
                to={`/organizations/${organizationId}/third-parties/${child.id}/overview`}
              >
                <Td>
                  <div className="flex gap-2 items-center">
                    <Avatar name={child.name} src={faviconUrl(child.websiteUrl)} />
                    <div>{child.name}</div>
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
                <Td noLink width={50} className="text-end">
                  {thirdParty.canUpdate && (
                    <Button
                      variant="tertiary"
                      icon={IconTrashCan}
                      onClick={() => handleDelete(child.id, child.name)}
                    />
                  )}
                </Td>
              </Tr>
            );
          })}
        </Tbody>
      </SortableTable>
    </div>
  );
}
