// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { faviconUrl } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { ThirdPartySupplyChainPageDeleteMutation } from "#/__generated__/core/ThirdPartySupplyChainPageDeleteMutation.graphql";
import type { ThirdPartySupplyChainPageFragment$key } from "#/__generated__/core/ThirdPartySupplyChainPageFragment.graphql";
import type { ThirdPartySupplyChainPagePaginationQuery } from "#/__generated__/core/ThirdPartySupplyChainPagePaginationQuery.graphql";
import type { ThirdPartySupplyChainPageQuery } from "#/__generated__/core/ThirdPartySupplyChainPageQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AddChildThirdPartyDialog } from "../_components/AddChildThirdPartyDialog";
import { thirdPartyHref } from "../_lib/thirdPartySections";

// Keep in sync with coredata.MaxThirdPartyLevel on the backend.
const MAX_THIRD_PARTY_LEVEL = 4;

export const thirdPartySupplyChainPageQuery = graphql`
  query ThirdPartySupplyChainPageQuery($thirdPartyId: ID!) {
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
        ...ThirdPartySupplyChainPageFragment
      }
    }
  }
`;

const paginatedFragment = graphql`
  fragment ThirdPartySupplyChainPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartySupplyChainPagePaginationQuery")
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
    ) @connection(key: "ThirdPartySupplyChainPageFragment_childThirdParties", filters: []) {
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
  mutation ThirdPartySupplyChainPageDeleteMutation(
    $input: DeleteThirdPartyInput!
    $connections: [ID!]!
  ) {
    deleteThirdParty(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartySupplyChainPageProps {
  queryRef: PreloadedQuery<ThirdPartySupplyChainPageQuery>;
}

export function ThirdPartySupplyChainPage({ queryRef }: ThirdPartySupplyChainPageProps) {
  const { node } = usePreloadedQuery<ThirdPartySupplyChainPageQuery>(
    thirdPartySupplyChainPageQuery,
    queryRef,
  );
  const thirdParty = node.__typename === "ThirdParty" ? node : null;
  const { t, i18n } = useTranslation();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();

  const pagination = usePaginationFragment<
    ThirdPartySupplyChainPagePaginationQuery,
    ThirdPartySupplyChainPageFragment$key
  >(paginatedFragment, thirdParty as ThirdPartySupplyChainPageFragment$key);
  const [deleteChild] = useMutation<ThirdPartySupplyChainPageDeleteMutation>(deleteChildMutation);

  usePageTitle(t("thirdPartySupplyChainPage.pageTitle", { name: thirdParty?.name ?? "" }));

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
        message: t("thirdPartySupplyChainPage.deleteConfirmation", { name: childName }),
      },
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartySupplyChainPage.title")}
        description={t("thirdPartySupplyChainPage.description")}
      >
        {thirdParty.canUpdate && thirdParty.level < MAX_THIRD_PARTY_LEVEL && (
          <AddChildThirdPartyDialog
            parentThirdPartyId={thirdParty.id}
            parentNamePath={parentNamePath}
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>{t("thirdPartySupplyChainPage.actions.add")}</Button>
          </AddChildThirdPartyDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={pagination.refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
        <Thead>
          <Tr>
            <SortableTh field="NAME">{t("thirdPartySupplyChainPage.columns.thirdParty")}</SortableTh>
            <Th>{t("thirdPartySupplyChainPage.columns.assessedAt")}</Th>
            <Th>{t("thirdPartySupplyChainPage.columns.dataRisk")}</Th>
            <Th>{t("thirdPartySupplyChainPage.columns.businessRisk")}</Th>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {childThirdParties.map((child) => {
            const latestAssessment = child.riskAssessments?.edges[0]?.node;

            return (
              <Tr
                key={child.id}
                to={thirdPartyHref(organizationId, child.id)}
              >
                <Td>
                  <div className="flex gap-2 items-center">
                    <Avatar name={child.name} src={faviconUrl(child.websiteUrl)} />
                    <div>{child.name}</div>
                  </div>
                </Td>
                <Td>
                  {latestAssessment?.createdAt
                    ? dateFormat(i18n.language, latestAssessment.createdAt)
                    : t("thirdPartySupplyChainPage.notAssessed")}
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
